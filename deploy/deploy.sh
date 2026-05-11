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

# Derive env file and default health URL from compose file name
case "$COMPOSE_FILE" in
  *staging*)  ENV_FILE="${ENV_FILE:-$(dirname "$COMPOSE_FILE")/.env.staging}"
              HEALTH_URL="${HEALTH_URL:-http://localhost:8081/health}" ;;
  *)          ENV_FILE="${ENV_FILE:-$(dirname "$COMPOSE_FILE")/.env.production}"
              HEALTH_URL="${HEALTH_URL:-http://localhost:8080/health}" ;;
esac

echo "==> [deploy] Tag: ${IMAGE_TAG}"

cd "$(dirname "$COMPOSE_FILE")"

# Pull new image before stopping anything (minimises downtime window)
IMAGE_TAG="$IMAGE_TAG" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull "$SERVICE"

# Capture current container ID for rollback
OLD_CONTAINER=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q "$SERVICE" 2>/dev/null || true)

echo "==> [deploy] Restarting ${SERVICE} with new image"
IMAGE_TAG="$IMAGE_TAG" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --no-deps "$SERVICE"

echo "==> [deploy] Waiting for health check..."
HEALTHY=0
for i in $(seq 1 "$HEALTH_RETRIES"); do
    sleep "$HEALTH_WAIT"
    if curl -sf "$HEALTH_URL" > /dev/null 2>&1; then
        HEALTHY=1
        echo "==> [deploy] Health check passed after $((i * HEALTH_WAIT))s"
        break
    fi
    echo "==> [deploy] Attempt $i/${HEALTH_RETRIES} — not healthy yet"
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

echo "==> [deploy] Done. Deployment successful."
