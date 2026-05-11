#!/usr/bin/env bash
# deploy/setup-firewall.sh
# ─────────────────────────────────────────────────────────────
# UFW firewall setup for production VPS
#
# Opens only:
#   - SSH
#   - HTTP  (80)
#   - HTTPS (443)
#
# Everything else blocked by default.
# ─────────────────────────────────────────────────────────────

set -euo pipefail

# SSH port (default 22)
SSH_PORT="${SSH_PORT:-22}"

echo "==> Configure UFW firewall"

# Reset all old rules
ufw --force reset

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH FIRST
# Important to avoid locking yourself out
ufw allow "${SSH_PORT}/tcp" comment 'SSH'

# Allow web traffic
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'

# Explicitly deny internal-only services
# (already blocked by default deny incoming,
# but added for security clarity)
ufw deny 5432/tcp comment 'PostgreSQL'
ufw deny 6379/tcp comment 'Redis'
ufw deny 8080/tcp comment 'Internal App Port'

# Enable firewall
ufw --force enable

# Show active rules
ufw status verbose

echo ""
echo "══════════════════════════════"
echo "Firewall configured"
echo "Allowed:"
echo "  SSH   : ${SSH_PORT}"
echo "  HTTP  : 80"
echo "  HTTPS : 443"
echo "══════════════════════════════"