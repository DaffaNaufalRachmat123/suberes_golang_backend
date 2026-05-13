#!/usr/bin/env bash
# deploy/provision_ssl_prod.sh
# ─────────────────────────────────────────────────────────────────────────────
# Issues a real Let's Encrypt certificate for suberes.com using
# the host-level Nginx (webroot challenge).
#
# Prerequisites:
#   - setup_host_nginx.sh has been run (host Nginx is running with dummy cert)
#   - Docker production containers are running
#   - DNS for suberes.com points to this server
#
# Usage:
#   sudo bash /opt/suberes/deploy/provision_ssl_prod.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SSL_EMAIL="support@suberes.com"
DOMAIN="suberes.com"
ALT_DOMAIN="www.suberes.com"

echo "[ssl-prod] Requesting Let's Encrypt certificate for ${DOMAIN}..."

certbot certonly \
  --webroot \
  --webroot-path /var/www/certbot \
  -d "${DOMAIN}" \
  -d "${ALT_DOMAIN}" \
  --email "${SSL_EMAIL}" \
  --agree-tos \
  --no-eff-email \
  --keep-until-expiring \
  --non-interactive

echo "[ssl-prod] Certificate issued successfully."
echo "[ssl-prod] Reloading host Nginx to activate real certificate..."
nginx -t && systemctl reload nginx

echo ""
echo "[ssl-prod] ✓ Real SSL cert is now active for ${DOMAIN}"
echo "[ssl-prod] ✓ Auto-renewal is handled by certbot timer (systemd) or cron."
echo ""

# ── Setup auto-renewal cron (if not already present) ──────────────────────
CRON_JOB="0 3 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'"
if ! crontab -l 2>/dev/null | grep -qF "certbot renew"; then
  echo "[ssl-prod] Adding certbot auto-renewal cron job..."
  (crontab -l 2>/dev/null; echo "${CRON_JOB}") | crontab -
  echo "[ssl-prod] ✓ Cron job added: ${CRON_JOB}"
else
  echo "[ssl-prod] ✓ Certbot auto-renewal cron already exists — skipping."
fi
