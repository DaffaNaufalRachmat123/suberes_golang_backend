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

  # Connect to postgres maintenance DB as superuser
  docker exec -i "${container}" psql -U postgres -v ON_ERROR_STOP=1 <<SQL
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
  docker exec -i "${container}" psql -U postgres -d "${db_name}" -v ON_ERROR_STOP=1 <<SQL
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
# Bootstrap staging DB
# -----------------------------------------------------------------------------
STAG_ENV_FILE="${APP_ROOT}/.env.staging"
if [[ -f "${STAG_ENV_FILE}" ]]; then
  # Load staging vars into a subshell so they don't pollute the outer scope
  STAG_DATABASE_VAL="$(set -a; source "${STAG_ENV_FILE}"; set +a; echo "${STAG_DATABASE:-}")"
  STAG_USERNAME_VAL="$(set -a; source "${STAG_ENV_FILE}"; set +a; echo "${STAG_USERNAME:-}")"
  STAG_PASSWORD_VAL="$(set -a; source "${STAG_ENV_FILE}"; set +a; echo "${STAG_PASSWORD:-}")"

  if [[ -n "${STAG_DATABASE_VAL}" && -n "${STAG_USERNAME_VAL}" && -n "${STAG_PASSWORD_VAL}" ]]; then
    bootstrap_db "suberes_postgres_stag" "${STAG_USERNAME_VAL}" "${STAG_PASSWORD_VAL}" "${STAG_DATABASE_VAL}"
  else
    echo "[provision] Skip staging DB bootstrap: STAG_DATABASE/STAG_USERNAME/STAG_PASSWORD incomplete"
  fi
else
  echo "[provision] Skip staging DB bootstrap: ${STAG_ENV_FILE} not found"
fi

# -----------------------------------------------------------------------------
# Bootstrap production DB
# -----------------------------------------------------------------------------
PROD_ENV_FILE="${APP_ROOT}/.env.production"
if [[ -f "${PROD_ENV_FILE}" ]]; then
  PROD_DATABASE_VAL="$(set -a; source "${PROD_ENV_FILE}"; set +a; echo "${PROD_DATABASE:-}")"
  PROD_USERNAME_VAL="$(set -a; source "${PROD_ENV_FILE}"; set +a; echo "${PROD_USERNAME:-}")"
  PROD_PASSWORD_VAL="$(set -a; source "${PROD_ENV_FILE}"; set +a; echo "${PROD_PASSWORD:-}")"

  if [[ -n "${PROD_DATABASE_VAL}" && -n "${PROD_USERNAME_VAL}" && -n "${PROD_PASSWORD_VAL}" ]]; then
    bootstrap_db "suberes_postgres" "${PROD_USERNAME_VAL}" "${PROD_PASSWORD_VAL}" "${PROD_DATABASE_VAL}"
  else
    echo "[provision] Skip production DB bootstrap: PROD_DATABASE/PROD_USERNAME/PROD_PASSWORD incomplete"
  fi
else
  echo "[provision] Skip production DB bootstrap: ${PROD_ENV_FILE} not found"
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

docker compose -f "$COMPOSE_FILE" up -d nginx

docker compose -f "$COMPOSE_FILE" run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "$SSL_PRIMARY_DOMAIN" \
  -d "$SSL_ALT_DOMAIN" \
  --email "$SSL_EMAIL" \
  --agree-tos --no-eff-email --keep-until-expiring

docker compose -f "$COMPOSE_FILE" exec -T nginx nginx -s reload || docker compose -f "$COMPOSE_FILE" restart nginx

echo "[provision] SSL bootstrap completed for ${DEPLOY_ENV}"