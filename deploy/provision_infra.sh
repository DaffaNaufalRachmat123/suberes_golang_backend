#!/usr/bin/env bash
# deploy/provision_infra.sh
# Ensures DB user/database privileges and SSL certificates for staging/production.
set -euo pipefail

run_psql() {
  local db_name="$1"
  local sql_payload="$2"
  docker run --rm -i \
    -e PGPASSWORD="${DB_ADMIN_PASS}" \
    postgres:16-alpine \
    psql "host=${DB_HOST} port=${DB_PORT} user=${DB_ADMIN_USER} dbname=${db_name} sslmode=prefer" \
    -v ON_ERROR_STOP=1 <<SQL
${sql_payload}
SQL
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

if [[ "$DEPLOY_ENV" == "staging" ]]; then
  ENV_FILE="${APP_ROOT}/.env.staging"
  DB_HOST_DEFAULT_VAR="STAG_HOST"
  DB_PORT_DEFAULT_VAR="STAG_PORT"
  DB_NAME_VAR="STAG_DATABASE"
  DB_USER_VAR="STAG_USERNAME"
  DB_PASS_VAR="STAG_PASSWORD"
  DB_ADMIN_USER_DEFAULT="${STAG_ADMIN_USERNAME:-postgres}"
  DB_ADMIN_PASS_DEFAULT="${STAG_ADMIN_PASSWORD:-}"
  SSL_PRIMARY_DOMAIN="staging.suberes.com"
  SSL_ALT_DOMAIN="www.staging.suberes.com"
else
  ENV_FILE="${APP_ROOT}/.env.production"
  DB_HOST_DEFAULT_VAR="PROD_HOST"
  DB_PORT_DEFAULT_VAR="PROD_PORT"
  DB_NAME_VAR="PROD_DATABASE"
  DB_USER_VAR="PROD_USERNAME"
  DB_PASS_VAR="PROD_PASSWORD"
  DB_ADMIN_USER_DEFAULT="${PROD_ADMIN_USERNAME:-postgres}"
  DB_ADMIN_PASS_DEFAULT="${PROD_ADMIN_PASSWORD:-}"
  SSL_PRIMARY_DOMAIN="suberes.com"
  SSL_ALT_DOMAIN="www.suberes.com"
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "[provision] Skip: env file not found: $ENV_FILE"
  exit 0
fi

set -a
source "$ENV_FILE"
set +a

DB_HOST="${DB_HOST:-${!DB_HOST_DEFAULT_VAR:-}}"
DB_PORT="${DB_PORT:-${!DB_PORT_DEFAULT_VAR:-5432}}"
DB_NAME="${DB_NAME:-${!DB_NAME_VAR:-}}"
DB_USER="${DB_USER:-${!DB_USER_VAR:-}}"
DB_PASS="${DB_PASS:-${!DB_PASS_VAR:-}}"
DB_ADMIN_USER="${DB_ADMIN_USER:-$DB_ADMIN_USER_DEFAULT}"
DB_ADMIN_PASS="${DB_ADMIN_PASS:-$DB_ADMIN_PASS_DEFAULT}"

if [[ -z "${DB_HOST}" || -z "${DB_NAME}" || -z "${DB_USER}" || -z "${DB_PASS}" ]]; then
  echo "[provision] Skip DB bootstrap: DB_HOST/DB_NAME/DB_USER/DB_PASS incomplete for ${DEPLOY_ENV}"
else
  if [[ -z "${DB_ADMIN_PASS}" ]]; then
    echo "[provision] Skip DB bootstrap: missing admin password for ${DEPLOY_ENV}"
  else
    echo "[provision] Ensuring role and database exist for ${DEPLOY_ENV}"

    run_psql postgres "
DO
\$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${DB_USER}') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', '${DB_USER}', '${DB_PASS}');
  ELSE
    EXECUTE format('ALTER ROLE %I WITH LOGIN PASSWORD %L', '${DB_USER}', '${DB_PASS}');
  END IF;
END
\$\$;
"

    run_psql postgres "
SELECT format('CREATE DATABASE %I OWNER %I', '${DB_NAME}', '${DB_USER}')
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_NAME}')
\gexec
"

    run_psql postgres "
GRANT ALL PRIVILEGES ON DATABASE "${DB_NAME}" TO "${DB_USER}";
"

    run_psql "${DB_NAME}" "
GRANT ALL ON SCHEMA public TO "${DB_USER}";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "${DB_USER}";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "${DB_USER}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO "${DB_USER}";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO "${DB_USER}";
"

    echo "[provision] DB bootstrap completed for ${DEPLOY_ENV}"
  fi
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