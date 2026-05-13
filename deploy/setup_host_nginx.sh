#!/usr/bin/env bash
# deploy/setup_host_nginx.sh
# ─────────────────────────────────────────────────────────────────────────────
# One-time setup: installs host-level Nginx, generates self-signed dummy certs
# for both domains so Nginx can start, then activates the host.conf config.
#
# Run ONCE as root on the VM before running provision_ssl_staging.sh /
# provision_ssl_prod.sh to get real Let's Encrypt certs.
#
# Usage:
#   sudo bash /opt/suberes/deploy/setup_host_nginx.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

PROD_DOMAIN="suberes.com"
PROD_ALT="www.suberes.com"
STAG_DOMAIN="staging.suberes.com"
STAG_ALT="www.staging.suberes.com"

HOST_CONF_SRC="/opt/suberes/nginx/conf.d/host.conf"
HOST_CONF_DST="/etc/nginx/conf.d/suberes.conf"

# ── 1. Install Nginx & Certbot ─────────────────────────────────────────────
echo "[setup] Installing nginx and certbot..."
apt-get update -qq
apt-get install -y -qq nginx certbot

# ── 2. Disable default Nginx site ─────────────────────────────────────────
echo "[setup] Disabling default nginx site..."
rm -f /etc/nginx/sites-enabled/default
rm -f /etc/nginx/conf.d/default.conf

# ── 3. Create directories ─────────────────────────────────────────────────
echo "[setup] Creating required directories..."
mkdir -p /var/www/certbot
mkdir -p /etc/letsencrypt/live/${PROD_DOMAIN}
mkdir -p /etc/letsencrypt/live/${STAG_DOMAIN}
chmod 755 /var/www/certbot
chmod -R 700 /etc/letsencrypt

# ── 4. Generate dummy self-signed certs ────────────────────────────────────
echo "[setup] Generating dummy self-signed cert for ${PROD_DOMAIN}..."
openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
  -keyout /etc/letsencrypt/live/${PROD_DOMAIN}/privkey.pem \
  -out    /etc/letsencrypt/live/${PROD_DOMAIN}/fullchain.pem \
  -subj   "/CN=${PROD_DOMAIN}" 2>/dev/null
cp /etc/letsencrypt/live/${PROD_DOMAIN}/fullchain.pem \
   /etc/letsencrypt/live/${PROD_DOMAIN}/chain.pem

echo "[setup] Generating dummy self-signed cert for ${STAG_DOMAIN}..."
openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
  -keyout /etc/letsencrypt/live/${STAG_DOMAIN}/privkey.pem \
  -out    /etc/letsencrypt/live/${STAG_DOMAIN}/fullchain.pem \
  -subj   "/CN=${STAG_DOMAIN}" 2>/dev/null
cp /etc/letsencrypt/live/${STAG_DOMAIN}/fullchain.pem \
   /etc/letsencrypt/live/${STAG_DOMAIN}/chain.pem

# ── 5. Install host Nginx config ───────────────────────────────────────────
echo "[setup] Installing host nginx config from ${HOST_CONF_SRC}..."
cp "${HOST_CONF_SRC}" "${HOST_CONF_DST}"

# ── 6. Test and start Nginx ────────────────────────────────────────────────
echo "[setup] Testing nginx config..."
nginx -t

echo "[setup] Enabling and starting nginx..."
systemctl enable nginx
systemctl restart nginx

echo ""
echo "[setup] ✓ Host Nginx is running with dummy SSL certs."
echo "[setup] ✓ Dummy cert for: ${PROD_DOMAIN}, ${PROD_ALT}"
echo "[setup] ✓ Dummy cert for: ${STAG_DOMAIN}, ${STAG_ALT}"
echo ""
echo "[setup] Next steps:"
echo "  1. Make sure Docker containers are running (staging & production)"
echo "  2. Run: sudo bash /opt/suberes/deploy/provision_ssl_staging.sh"
echo "  3. Run: sudo bash /opt/suberes/deploy/provision_ssl_prod.sh"
