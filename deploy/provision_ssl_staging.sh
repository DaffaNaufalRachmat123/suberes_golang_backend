#!/usr/bin/env bash
# deploy/provision_ssl_staging.sh
# ─────────────────────────────────────────────────────────────────────────────
# Issues a real Let's Encrypt certificate for staging.suberes.com using
# the host-level Nginx (webroot challenge).
#
# Prerequisites:
#   - setup_host_nginx.sh has been run (host Nginx is running with dummy cert)
#   - Docker staging containers are running
#   - DNS for staging.suberes.com points to this server
#
# Usage:
#   sudo bash /opt/suberes/deploy/provision_ssl_staging.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SSL_EMAIL="support@suberes.com"
DOMAIN="staging.suberes.com"
ALT_DOMAIN="www.staging.suberes.com"

# ── Remove dummy self-signed cert so certbot can create a fresh one ──────────
if [[ -d "/etc/letsencrypt/live/${DOMAIN}" ]]; then
  echo "[ssl-staging] Removing dummy cert for ${DOMAIN}..."
  rm -rf "/etc/letsencrypt/live/${DOMAIN}"
  rm -rf "/etc/letsencrypt/archive/${DOMAIN}"
  rm -f  "/etc/letsencrypt/renewal/${DOMAIN}.conf"
fi

echo "[ssl-staging] Requesting Let's Encrypt certificate for ${DOMAIN}..."

certbot certonly \
  --webroot \
  --webroot-path /var/www/certbot \
  -d "${DOMAIN}" \
  -d "${ALT_DOMAIN}" \
  --email "${SSL_EMAIL}" \
  --agree-tos \
  --no-eff-email \
  --keep-until-expiring \
  --expand \
  --non-interactive

echo "[ssl-staging] Certificate issued successfully."
echo "[ssl-staging] Reloading host Nginx to activate real certificate..."
nginx -t && systemctl reload nginx

echo ""
echo "[ssl-staging] ✓ Real SSL cert is now active for ${DOMAIN}"
echo "[ssl-staging] ✓ Auto-renewal is handled by certbot timer (systemd) or cron."
echo ""

# ── Setup auto-renewal cron (if not already present) ──────────────────────
CRON_JOB="0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'"
if ! crontab -l 2>/dev/null | grep -qF "certbot renew"; then
  echo "[ssl-staging] Adding certbot auto-renewal cron job..."
  (crontab -l 2>/dev/null; echo "${CRON_JOB}") | crontab -
  echo "[ssl-staging] ✓ Cron job added: ${CRON_JOB}"
else
  echo "[ssl-staging] ✓ Certbot auto-renewal cron already exists — skipping."
fi
