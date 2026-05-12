#!/usr/bin/env bash
# deploy/deploy.sh — Zero-downtime deployment script for single VPS
# ─────────────────────────────────────────────────────────────────────────────
# Pulls the new image, replaces the running container, verifies health,
# and rolls back automatically if the new container fails to become healthy.
#
# Prerequisites on VPS:
#   • Docker + Docker Compose plugin installed
#   • deployer user has docker group membership
#   • /opt/suberes/.env.production exists with correct secrets
#   • GHCR (or Docker Hub) credentials configured: docker login ghcr.io
#
# Usage:
#   IMAGE_TAG=abc123 bash deploy/deploy.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-/opt/suberes/docker-compose.production.yml}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
SERVICE="${SERVICE:-app}"
HEALTH_RETRIES="${HEALTH_RETRIES:-10}"
HEALTH_WAIT="${HEALTH_WAIT:-5}"

ensure_path_access() {
  local path="$1"
  local mode="$2"

  if [[ -e "$path" ]]; then
    chmod "$mode" "$path" 2>/dev/null || sudo chmod "$mode" "$path"
  fi
}

ensure_env_permissions() {
  local env_path="$1"
  local runner_user
  runner_user="$(id -un)"
  local runner_group
  runner_group="$(id -gn)"

  local env_dir
  env_dir="$(dirname "$env_path")"

  # Compose must traverse parent dirs to read env_file.
  ensure_path_access "/opt" 755
  ensure_path_access "/opt/suberes" 755

  if [[ ! -d "$env_dir" ]]; then
    mkdir -p "$env_dir" 2>/dev/null || sudo mkdir -p "$env_dir"
  fi

  chown "$runner_user:$runner_group" "$env_dir" 2>/dev/null || sudo chown "$runner_user:$runner_group" "$env_dir"
  ensure_path_access "$env_dir" 750

  if [[ -f "$env_path" ]]; then
    chown "$runner_user:$runner_group" "$env_path" 2>/dev/null || sudo chown "$runner_user:$runner_group" "$env_path"
    ensure_path_access "$env_path" 640
  else
    echo "==> [deploy] WARNING: env file '${env_path}' not found (compose may fail if service uses env_file)"
  fi
}

# Derive env file and default health URL from compose file name
case "$COMPOSE_FILE" in
  *staging*)  ENV_FILE="${ENV_FILE:-$(dirname "$COMPOSE_FILE")/.env.staging}"
              HEALTH_URL="${HEALTH_URL:-http://localhost:8081/health}" ;;
  *)          ENV_FILE="${ENV_FILE:-$(dirname "$COMPOSE_FILE")/.env.production}"
              HEALTH_URL="${HEALTH_URL:-http://localhost:8080/health}" ;;
esac

# Runtime env_file mounted by compose service definitions
case "$COMPOSE_FILE" in
  *staging*)  SERVICE_ENV_FILE="/opt/suberes/staging/.env" ;;
  *)          SERVICE_ENV_FILE="/opt/suberes/production/.env" ;;
esac

ensure_env_permissions "$SERVICE_ENV_FILE"

echo "==> [deploy] Tag: ${IMAGE_TAG}"

cd "$(dirname "$COMPOSE_FILE")"

# Pull new image before stopping anything (minimises downtime window)
IMAGE_TAG="$IMAGE_TAG" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull "$SERVICE"

# Capture current container ID for rollback
OLD_CONTAINER=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q "$SERVICE" 2>/dev/null || true)

echo "==> [deploy] Restarting ${SERVICE} with new image"
IMAGE_TAG="$IMAGE_TAG" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --no-deps "$SERVICE"

# Derive container name from compose file
case "$COMPOSE_FILE" in
  *staging*) APP_CONTAINER="suberes_app_stag" ;;
  *)         APP_CONTAINER="suberes_app" ;;
esac

echo "==> [deploy] Waiting for health check..."
HEALTHY=0
for i in $(seq 1 "$HEALTH_RETRIES"); do
    sleep "$HEALTH_WAIT"
    STATUS=$(docker inspect --format='{{.State.Health.Status}}' "${APP_CONTAINER}" 2>/dev/null || echo "unknown")
    if [[ "$STATUS" == "healthy" ]]; then
        HEALTHY=1
        echo "==> [deploy] Health check passed after $((i * HEALTH_WAIT))s"
        break
    fi
    echo "==> [deploy] Attempt $i/${HEALTH_RETRIES} — status: ${STATUS}"
done

if [ "$HEALTHY" -ne 1 ]; then
    echo "==> [ROLLBACK] New container unhealthy, rolling back to previous image"
    if [ -n "$OLD_CONTAINER" ]; then
        docker start "$OLD_CONTAINER" || true
    fi
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs --tail=50 "$SERVICE"
    exit 1
fi

echo "==> [deploy] Removing dangling images"
docker image prune -f

# ── Run seed after app is healthy ────────────────────────────────────────────
# Tables are created by GORM AutoMigrate on first startup, so seed must run
# AFTER the app is confirmed healthy (not during provision).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SEED_FILE="${SCRIPT_DIR}/seed.sql"

if [[ -f "${SEED_FILE}" ]]; then
  # Derive container name and db vars from env file
  set -a; source "$ENV_FILE"; set +a

  case "$COMPOSE_FILE" in
    *staging*)
      PG_CONTAINER="suberes_postgres_stag"
      PG_USER="${STAG_USERNAME:-}"
      PG_DB="${STAG_DATABASE:-}"
      ;;
    *)
      PG_CONTAINER="suberes_postgres"
      PG_USER="${PROD_USERNAME:-}"
      PG_DB="${PROD_DATABASE:-}"
      ;;
  esac

  if [[ -n "${PG_USER}" && -n "${PG_DB}" ]] && \
     docker ps --format '{{.Names}}' | grep -qx "${PG_CONTAINER}"; then
    echo "==> [deploy] Running seed on '${PG_DB}' in '${PG_CONTAINER}'"
    docker exec -i "${PG_CONTAINER}" \
      psql -U "${PG_USER}" -d "${PG_DB}" -v ON_ERROR_STOP=1 < "${SEED_FILE}"
    echo "==> [deploy] Seed completed"
  else
    echo "==> [deploy] Skip seed: postgres container not running or vars missing"
  fi
else
  echo "==> [deploy] Skip seed: ${SEED_FILE} not found"
fi

echo "==> [deploy] Done. Deployment successful."
