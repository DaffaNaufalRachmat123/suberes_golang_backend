#!/usr/bin/env bash
# deploy/provision_ssl.sh
# Only handles SSL certbot + nginx reload. Run as root!
set -euo pipefail

# ENV: DEPLOY_ENV (staging/production), SSL_EMAIL (required)

if [[ -z "${DEPLOY_ENV:-}" ]]; then
  echo "[ssl] DEPLOY_ENV is not set (staging/production)" >&2
  exit 1
fi
if [[ -z "${SSL_EMAIL:-}" ]]; then
  echo "[ssl] SSL_EMAIL is not set" >&2
  exit 1
fi

if [[ "${DEPLOY_ENV}" == "staging" ]]; then
  SSL_PRIMARY_DOMAIN="staging.suberes.com"
  SSL_ALT_DOMAIN="www.staging.suberes.com"
  COMPOSE_FILE="/opt/suberes/docker-compose.staging.yml"
elif [[ "${DEPLOY_ENV}" == "production" ]]; then
  SSL_PRIMARY_DOMAIN="suberes.com"
  SSL_ALT_DOMAIN="www.suberes.com"
  COMPOSE_FILE="/opt/suberes/docker-compose.production.yml"
else
  echo "[ssl] Unknown DEPLOY_ENV: $DEPLOY_ENV" >&2
  exit 1
fi

# 1. Create required host directories (as root)
sudo mkdir -p /var/www/certbot
sudo mkdir -p /etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}
sudo chmod -R 755 /var/www/certbot
sudo chmod -R 700 /etc/letsencrypt

# 2. Generate dummy cert if needed
CERT_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/fullchain.pem"
KEY_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/privkey.pem"
CHAIN_FILE="/etc/letsencrypt/live/${SSL_PRIMARY_DOMAIN}/chain.pem"
if [[ ! -f "${CERT_FILE}" ]]; then
  echo "[ssl] No SSL cert found — generating temporary self-signed cert for nginx bootstrap..."
  if ! command -v openssl &>/dev/null; then
    echo "[ssl] Installing openssl..."
    apt-get update -qq && apt-get install -y -qq openssl
  fi
  sudo openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
    -keyout "${KEY_FILE}" \
    -out "${CERT_FILE}" \
    -subj "/CN=${SSL_PRIMARY_DOMAIN}" \
    2>/dev/null
  sudo cp "${CERT_FILE}" "${CHAIN_FILE}"
  echo "[ssl] Temporary self-signed cert created at ${CERT_FILE}"
else
  echo "[ssl] Existing SSL cert found at ${CERT_FILE} — skipping dummy cert generation"
fi

# 3. Start nginx container (so certbot can do challenge)
echo "[ssl] Starting nginx container..."
docker compose --env-file /opt/suberes/${DEPLOY_ENV}/.env -f "${COMPOSE_FILE}" up -d nginx

# 4. Run certbot to issue/renew real cert
echo "[ssl] Running certbot to issue/renew Let's Encrypt cert for ${SSL_PRIMARY_DOMAIN}..."
docker compose --env-file /opt/suberes/${DEPLOY_ENV}/.env -f "${COMPOSE_FILE}" run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "${SSL_PRIMARY_DOMAIN}" \
  -d "${SSL_ALT_DOMAIN}" \
  --email "${SSL_EMAIL}" \
  --agree-tos --no-eff-email --keep-until-expiring

echo "[ssl] Let's Encrypt cert issued/renewed successfully"

# 5. Reload nginx to pick up the real cert
echo "[ssl] Reloading nginx with the real Let's Encrypt cert..."
docker compose --env-file /opt/suberes/${DEPLOY_ENV}/.env -f "${COMPOSE_FILE}" \
  exec -T nginx nginx -s reload \
  || docker compose --env-file /opt/suberes/${DEPLOY_ENV}/.env -f "${COMPOSE_FILE}" restart nginx

echo "[ssl] ✓ SSL bootstrap completed for ${DEPLOY_ENV}"
echo "[ssl] ✓ ${SSL_PRIMARY_DOMAIN} is now served over HTTPS"
echo "[ssl] ✓ ${SSL_ALT_DOMAIN} is now served over HTTPS"
