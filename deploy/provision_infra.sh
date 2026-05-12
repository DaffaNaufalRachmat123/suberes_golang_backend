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
  mkdir -p "$PRODUCTION_DIR"

  chown deployer:deployer "$PRODUCTION_DIR" || true
  chmod 750 "$PRODUCTION_DIR"
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
  cp "${APP_ROOT}/.env.production" "${PRODUCTION_DIR}/.env"

  chown deployer:deployer "${PRODUCTION_DIR}/.env" || true
  chmod 640 "${PRODUCTION_DIR}/.env"

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

  echo "[provision] Waiting for postgres to become healthy (max 90s)..."
  POSTGRES_HEALTHY=0
  for i in $(seq 1 18); do
    sleep 5
    STATUS=$(docker inspect --format='{{.State.Health.Status}}' "${_PG_CONTAINER}" 2>/dev/null || echo "missing")
    if [[ "${STATUS}" == "healthy" ]]; then
      POSTGRES_HEALTHY=1
      echo "[provision] Postgres is ready"
      break
    fi
    echo "[provision] Waiting for postgres... attempt ${i}/18 (status: ${STATUS})"
  done

  if [[ "${POSTGRES_HEALTHY}" -ne 1 ]]; then
    echo "[provision] ERROR: postgres did not become healthy in time, aborting"
    echo "[provision] Last health status: $(docker inspect --format='{{json .State.Health}}' "${_PG_CONTAINER}" 2>/dev/null || echo 'N/A')"
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

echo ""
echo "[provision] ✓ DB and environment provisioning completed for ${DEPLOY_ENV}"