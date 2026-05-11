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

# -----------------------------------------------------------------------------
# pre_migrate_db <container_name> <db_user> <db_name>
#   Prepares existing varchar latitude/longitude columns for GORM AutoMigrate's
#   ALTER TYPE to double precision.
#
#   GORM will generate:
#     ALTER TABLE users ALTER COLUMN latitude TYPE double precision USING latitude::double precision
#   This fails if the column has DEFAULT '' or rows containing ''.
#
#   This function (idempotent):
#     1. Drops the '' default from latitude/longitude (if column is still varchar)
#     2. Replaces '' values with NULL so the USING clause succeeds
# -----------------------------------------------------------------------------
pre_migrate_db() {
  local container="$1"
  local db_user="$2"
  local db_name="$3"

  if ! docker ps --format '{{.Names}}' | grep -qx "${container}"; then
    echo "[provision] Skip pre-migrate for '${container}': container not running"
    return 0
  fi

  echo "[provision] Running pre-migrate on '${db_name}' in '${container}'..."

  docker exec -i "${container}" psql -U "${db_user}" -d "${db_name}" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
BEGIN
  -- Only act when the column is still varchar (before GORM migrates it to double precision).
  -- Running this on an already-migrated DB is a safe no-op.
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'users' AND column_name = 'latitude'
      AND data_type IN ('character varying', 'text', 'character')
  ) THEN
    -- Drop the default '' so PostgreSQL can ALTER TYPE without the "default cannot be cast" error
    ALTER TABLE users ALTER COLUMN latitude DROP DEFAULT;
    -- Replace '' with NULL; '' cannot be cast to double precision
    UPDATE users SET latitude = NULL WHERE latitude = '';

    -- Drop geom generated column (it depends on latitude/longitude, blocks ALTER TYPE)
    ALTER TABLE users DROP COLUMN IF EXISTS geom;

    RAISE NOTICE 'pre_migrate: latitude default dropped, empty strings set to NULL, geom dropped';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'users' AND column_name = 'longitude'
      AND data_type IN ('character varying', 'text', 'character')
  ) THEN
    ALTER TABLE users ALTER COLUMN longitude DROP DEFAULT;
    UPDATE users SET longitude = NULL WHERE longitude = '';
    RAISE NOTICE 'pre_migrate: longitude default dropped and empty strings set to NULL';
  END IF;
END $$;
SQL

  echo "[provision] Pre-migrate done for '${db_name}' in '${container}'"
}


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
    _PG_CONTAINER="suberes_postgres_stag"
    _PG_VOLUME="suberes_staging_postgres_data_stag"
  else
    _PG_USER="${PROD_USERNAME:-postgres}"
    _PG_CONTAINER="suberes_postgres"
    _PG_VOLUME="suberes_production_postgres_data"
  fi

  # If the running container is NOT using the postgis image, recreate it.
  # This handles migration from postgres:16-alpine → postgis/postgis:16-alpine.
  CURRENT_IMAGE=$(docker inspect --format='{{.Config.Image}}' "${_PG_CONTAINER}" 2>/dev/null || echo "")
  if [[ -n "${CURRENT_IMAGE}" && "${CURRENT_IMAGE}" != postgis/postgis* ]]; then
    echo "[provision] Detected non-PostGIS postgres image ('${CURRENT_IMAGE}'). Recreating container with postgis/postgis image..."
    docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" stop postgres || true
    docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" rm -f postgres || true
    # Drop the old volume so initdb (which installs extensions) runs fresh
    docker volume rm "${_PG_VOLUME}" 2>/dev/null || true
    echo "[provision] Old postgres volume removed — data will be re-provisioned"
  fi

  echo "[provision] Starting postgres for ${DEPLOY_ENV}..."
  docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" up -d postgres

  echo "[provision] Waiting for postgres to become healthy (max 60s)..."
  POSTGRES_HEALTHY=0
  for i in $(seq 1 12); do
    sleep 5
    if docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
        pg_isready -U "${_PG_USER}" > /dev/null 2>&1; then      POSTGRES_HEALTHY=1
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
    pre_migrate_db "suberes_postgres_stag" "${STAG_USERNAME}" "${STAG_DATABASE}"
  else
    echo "[provision] Skip staging DB bootstrap: STAG_DATABASE/STAG_USERNAME/STAG_PASSWORD incomplete"
  fi
elif [[ "${DEPLOY_ENV}" == "production" ]]; then
  if [[ -n "${PROD_DATABASE:-}" && -n "${PROD_USERNAME:-}" && -n "${PROD_PASSWORD:-}" ]]; then
    bootstrap_db "suberes_postgres" "${PROD_USERNAME}" "${PROD_PASSWORD}" "${PROD_DATABASE}"
    pre_migrate_db "suberes_postgres" "${PROD_USERNAME}" "${PROD_DATABASE}"
  else
    echo "[provision] Skip production DB bootstrap: PROD_DATABASE/PROD_USERNAME/PROD_PASSWORD incomplete"
  fi
fi

# =============================================================================
# Nginx + SSL Bootstrap
# =============================================================================
# NOTE: Nginx runs as a Docker container (image: nginx:1.27-alpine).
#       /etc/nginx does NOT exist on the host — it lives inside the container.
#       The container mounts /etc/letsencrypt and /var/www/certbot from the HOST.
#
# Chicken-and-egg strategy:
#   1. Ensure required host directories exist.
#   2. If no real Let's Encrypt cert exists yet, generate a temporary self-signed
#      cert so nginx can start (it needs the cert files to load the HTTPS block).
#   3. Start the nginx container with the temporary cert.
#   4. Run certbot (via Docker) to issue/renew the real cert using webroot.
#   5. Reload nginx — it now serves with the real Let's Encrypt cert.
# =============================================================================

# SSL domain depends on DEPLOY_ENV
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

echo "[provision] === Nginx + SSL bootstrap for ${DEPLOY_ENV} (${SSL_PRIMARY_DOMAIN}) ==="

# ── 1. Create required host directories ──────────────────────────────────────
echo "[provision] Creating host directories for certbot and letsencrypt..."
mkdir -p /var/www/certbot
mkdir -p /etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}
chmod -R 755 /var/www/certbot
chmod -R 700 /etc/letsencrypt

# ── 2. Generate a temporary self-signed cert if no real cert exists yet ───────
# This lets nginx start for the first time so certbot can do the ACME challenge.
CERT_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/fullchain.pem"
KEY_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/privkey.pem"
CHAIN_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/chain.pem"

if [[ ! -f "${CERT_FILE}" ]]; then
  echo "[provision] No SSL cert found — generating temporary self-signed cert for nginx bootstrap..."

  if ! command -v openssl &>/dev/null; then
    echo "[provision] Installing openssl..."
    apt-get update -qq && apt-get install -y -qq openssl
  fi

  openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
    -keyout "${KEY_FILE}" \
    -out "${CERT_FILE}" \
    -subj "/CN=${SSL_PRIMARY_DOMAIN}" \
    2>/dev/null

  # chain.pem must also exist (nginx ssl_trusted_certificate requires it)
  cp "${CERT_FILE}" "${CHAIN_FILE}"

  echo "[provision] Temporary self-signed cert created at ${CERT_FILE}"
else
  echo "[provision] Existing SSL cert found at ${CERT_FILE} — skipping dummy cert generation"
fi

# ── 3. Start nginx container (uses the cert that now exists on the host) ──────
echo "[provision] Starting nginx container..."
docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" up -d nginx

# Wait briefly for nginx to fully initialize
for i in $(seq 1 6); do
  sleep 3
  if docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" \
      exec -T nginx nginx -t &>/dev/null; then
    echo "[provision] Nginx is healthy"
    break
  fi
  echo "[provision] Waiting for nginx... attempt ${i}/6"
done

# ── 4. Issue / renew the real Let's Encrypt certificate ───────────────────────
echo "[provision] Running certbot to issue/renew Let's Encrypt cert for ${SSL_PRIMARY_DOMAIN}..."
docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "${SSL_PRIMARY_DOMAIN}" \
  -d "${SSL_ALT_DOMAIN}" \
  --email "${SSL_EMAIL}" \
  --agree-tos --no-eff-email --keep-until-expiring

echo "[provision] Let's Encrypt cert issued/renewed successfully"

# ── 5. Reload nginx to pick up the real cert ──────────────────────────────────
echo "[provision] Reloading nginx with the real Let's Encrypt cert..."
docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" \
  exec -T nginx nginx -s reload \
  || docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" restart nginx

# ── 6. Final verification ──────────────────────────────────────────────────────
echo "[provision] Verifying nginx config after reload..."
docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" \
  exec -T nginx nginx -T 2>&1 | grep -i "server_name" || true

echo ""
echo "[provision] ✓ SSL bootstrap completed for ${DEPLOY_ENV}"
echo "[provision] ✓ ${SSL_PRIMARY_DOMAIN} is now served over HTTPS"
echo "[provision] ✓ ${SSL_ALT_DOMAIN} is now served over HTTPS"