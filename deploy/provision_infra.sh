#!/usr/bin/env bash
# deploy/provision_infra.sh
# Ensures DB user/database privileges and SSL certificates for staging/production.
set -euo pipefail

# -----------------------------------------------------------------------------
# bootstrap_db <container_name> <db_user> <db_pass> <db_name>
#   Connects as postgres superuser via docker exec (no password required since
#   peer/trust auth is used inside the container), then:
#     1. Creates the role if it does not exist (or updates the password)
#     2. Creates the database if it does not exist
#     3. Grants full privileges on the database and public schema
# -----------------------------------------------------------------------------
bootstrap_db() {
  local container="$1"
  local db_user="$2"
  local db_pass="$3"
  local db_name="$4"

  if ! docker ps --format '{{.Names}}' | grep -qx "${container}"; then
    echo "[provision] Skip DB bootstrap for '${container}': container not running"
    return 0
  fi

  echo "[provision] Checking role '${db_user}' and database '${db_name}' in '${container}'"

  # Connect as db_user to the 'postgres' maintenance database.
  # When POSTGRES_USER is a custom value, that user IS the superuser —
  # Docker does not create a separate 'postgres' role.
  # We must specify -d postgres explicitly; otherwise psql defaults to a
  # database named after the user (which doesn't exist yet).
  docker exec -i "${container}" psql -U "${db_user}" -d postgres -v ON_ERROR_STOP=1 <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${db_user}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${db_user}', '${db_pass}');
    RAISE NOTICE 'Role % created.', '${db_user}';
  ELSE
    EXECUTE format('ALTER ROLE %I WITH LOGIN PASSWORD %L', '${db_user}', '${db_pass}');
    RAISE NOTICE 'Role % already exists — password synced.', '${db_user}';
  END IF;
END
\$\$;

SELECT format('CREATE DATABASE %I OWNER %I', '${db_name}', '${db_user}')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${db_name}')
\gexec

GRANT ALL PRIVILEGES ON DATABASE "${db_name}" TO "${db_user}";
SQL

  # Grant schema-level privileges inside the target database
  docker exec -i "${container}" psql -U "${db_user}" -d "${db_name}" -v ON_ERROR_STOP=1 <<SQL
GRANT ALL ON SCHEMA public TO "${db_user}";
GRANT ALL PRIVILEGES ON ALL TABLES    IN SCHEMA public TO "${db_user}";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "${db_user}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES    TO "${db_user}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "${db_user}";
SQL

  echo "[provision] DB bootstrap done for '${db_name}' in '${container}'"
}

DEPLOY_ENV="${DEPLOY_ENV:-production}"
COMPOSE_FILE="${COMPOSE_FILE:-/opt/suberes/docker-compose.production.yml}"
APP_ROOT="$(cd "$(dirname "$COMPOSE_FILE")" && pwd)"

# Derive the env file for docker compose --env-file (must match DEPLOY_ENV)
if [[ "${DEPLOY_ENV}" == "staging" ]]; then
  COMPOSE_ENV_FILE="${APP_ROOT}/.env.staging"
else
  COMPOSE_ENV_FILE="${APP_ROOT}/.env.production"
fi

# -----------------------------------------------------------------------------
# Prepare isolated env directories
# -----------------------------------------------------------------------------

STAGING_DIR="/opt/suberes/staging"
PRODUCTION_DIR="/opt/suberes/production"

# Create staging dir if missing
if [[ ! -d "$STAGING_DIR" ]]; then
  mkdir -p "$STAGING_DIR"

  chown deployer:deployer "$STAGING_DIR" || true
  chmod 755 "$STAGING_DIR"
fi

# Create production dir if missing
if [[ ! -d "$PRODUCTION_DIR" ]]; then
  sudo mkdir -p "$PRODUCTION_DIR"

  sudo chown root:root "$PRODUCTION_DIR"
  sudo chmod 700 "$PRODUCTION_DIR"
fi

# -----------------------------------------------------------------------------
# Sync staging env
# -----------------------------------------------------------------------------

if [[ -f "${APP_ROOT}/.env.staging" ]]; then
  cp "${APP_ROOT}/.env.staging" "${STAGING_DIR}/.env"

  chown deployer:deployer "${STAGING_DIR}/.env" || true
  chmod 640 "${STAGING_DIR}/.env"

  echo "[provision] Synced staging env"
else
  echo "[provision] Skip staging env sync: .env.staging not found"
fi

# -----------------------------------------------------------------------------
# Sync production env
# -----------------------------------------------------------------------------

if [[ -f "${APP_ROOT}/.env.production" ]]; then
  sudo cp "${APP_ROOT}/.env.production" "${PRODUCTION_DIR}/.env"

  sudo chown root:root "${PRODUCTION_DIR}/.env"
  sudo chmod 600 "${PRODUCTION_DIR}/.env"

  echo "[provision] Synced production env"
else
  echo "[provision] Skip production env sync: .env.production not found"
fi

# -----------------------------------------------------------------------------
# Start postgres so DB bootstrap can connect via docker exec
# -----------------------------------------------------------------------------
if [[ -f "${COMPOSE_ENV_FILE}" ]]; then
  # Source env file early so STAG_*/PROD_* vars are available for pg_isready -U
  set -a; source "${COMPOSE_ENV_FILE}"; set +a

  if [[ "${DEPLOY_ENV}" == "staging" ]]; then
    _PG_USER="${STAG_USERNAME:-postgres}"
  else
    _PG_USER="${PROD_USERNAME:-postgres}"
  fi

  echo "[provision] Starting postgres for ${DEPLOY_ENV}..."
  docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" up -d postgres

  echo "[provision] Waiting for postgres to become healthy (max 60s)..."
  POSTGRES_HEALTHY=0
  for i in $(seq 1 12); do
    sleep 5
    if docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
        pg_isready -U "${_PG_USER}" > /dev/null 2>&1; then
      POSTGRES_HEALTHY=1
      echo "[provision] Postgres is ready"
      break
    fi
    echo "[provision] Waiting for postgres... attempt ${i}/12"
  done

  if [[ "${POSTGRES_HEALTHY}" -ne 1 ]]; then
    echo "[provision] ERROR: postgres did not become healthy in time, aborting"
    exit 1
  fi
else
  echo "[provision] Skip postgres startup: ${COMPOSE_ENV_FILE} not found"
fi

# Env vars already sourced above from COMPOSE_ENV_FILE
if [[ "${DEPLOY_ENV}" == "staging" ]]; then
  if [[ -n "${STAG_DATABASE:-}" && -n "${STAG_USERNAME:-}" && -n "${STAG_PASSWORD:-}" ]]; then
    bootstrap_db "suberes_postgres_stag" "${STAG_USERNAME}" "${STAG_PASSWORD}" "${STAG_DATABASE}"
  else
    echo "[provision] Skip staging DB bootstrap: STAG_DATABASE/STAG_USERNAME/STAG_PASSWORD incomplete"
  fi
elif [[ "${DEPLOY_ENV}" == "production" ]]; then
  if [[ -n "${PROD_DATABASE:-}" && -n "${PROD_USERNAME:-}" && -n "${PROD_PASSWORD:-}" ]]; then
    bootstrap_db "suberes_postgres" "${PROD_USERNAME}" "${PROD_PASSWORD}" "${PROD_DATABASE}"
  else
    echo "[provision] Skip production DB bootstrap: PROD_DATABASE/PROD_USERNAME/PROD_PASSWORD incomplete"
  fi
fi

# SSL domain depends on DEPLOY_ENV (used below)
if [[ "${DEPLOY_ENV}" == "staging" ]]; then
  SSL_PRIMARY_DOMAIN="staging.suberes.com"
  SSL_ALT_DOMAIN="www.staging.suberes.com"
else
  SSL_PRIMARY_DOMAIN="suberes.com"
  SSL_ALT_DOMAIN="www.suberes.com"
fi

if [[ -z "${SSL_EMAIL:-}" ]]; then
  echo "[provision] Skip SSL bootstrap: SSL_EMAIL is empty"
  exit 0
fi

echo "[provision] Ensuring SSL certificate for ${SSL_PRIMARY_DOMAIN} and ${SSL_ALT_DOMAIN}"

docker compose --env-file "${COMPOSE_ENV_FILE}" -f "$COMPOSE_FILE" up -d nginx

docker compose --env-file "${COMPOSE_ENV_FILE}" -f "$COMPOSE_FILE" run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "$SSL_PRIMARY_DOMAIN" \
  -d "$SSL_ALT_DOMAIN" \
  --email "$SSL_EMAIL" \
  --agree-tos --no-eff-email --keep-until-expiring

docker compose --env-file "${COMPOSE_ENV_FILE}" -f "$COMPOSE_FILE" exec -T nginx nginx -s reload \
  || docker compose --env-file "${COMPOSE_ENV_FILE}" -f "$COMPOSE_FILE" restart nginx

echo "[provision] SSL bootstrap completed for ${DEPLOY_ENV}"