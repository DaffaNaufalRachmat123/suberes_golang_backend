#!/usr/bin/env bash
# deploy/vps-hardening.sh
# ─────────────────────────────────────────────────────────────────────────────
# Run ONCE on a freshly-provisioned Ubuntu 22.04/24.04 VPS as root.
# Hardens SSH, installs fail2ban, enables automatic security updates.
#
# Usage:
#   sudo bash deploy/vps-hardening.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

# ── Configuration — edit before running ──────────────────────────────────────
SSH_PORT="${SSH_PORT:-22}"              # SSH port (default 22, ganti jika perlu)
DEPLOY_USER="${DEPLOY_USER:-deployer}"  # non-root user for deployments
# ─────────────────────────────────────────────────────────────────────────────

echo "==> [1/6] System update"
# Set DEBIAN_FRONTEND agar apt tidak stuck nunggu input interaktif (tzdata dll)
export DEBIAN_FRONTEND=noninteractive

# Pre-set timezone sebelum apt jalan — cegah dpkg --configure tzdata hang
ln -sf /usr/share/zoneinfo/Asia/Jakarta /etc/localtime
echo "Asia/Jakarta" > /etc/timezone

# Selesaikan configure yang mungkin stuck dari sesi sebelumnya
dpkg --configure -a
apt-get install -f -y

apt-get update -y
apt-get upgrade -y
apt-get install -y tzdata ufw fail2ban unattended-upgrades curl wget gnupg2 ca-certificates

echo "==> [2/6] Create deployment user: ${DEPLOY_USER}"
if id "$DEPLOY_USER" &>/dev/null; then
    echo "User ${DEPLOY_USER} sudah ada — skip pembuatan user"
else
    useradd -m -s /bin/bash "$DEPLOY_USER"
    echo "User ${DEPLOY_USER} berhasil dibuat"
fi

# Pastikan user masuk group docker (idempotent — aman jika sudah ada)
usermod -aG docker "$DEPLOY_USER"
echo "User ${DEPLOY_USER} ditambahkan ke group docker"

# Copy SSH authorized keys dari root agar deployer bisa login dengan key yang sama
# Hanya copy jika belum ada authorized_keys milik deployer
if [ -f /root/.ssh/authorized_keys ]; then
    mkdir -p "/home/${DEPLOY_USER}/.ssh"
    if [ ! -f "/home/${DEPLOY_USER}/.ssh/authorized_keys" ]; then
        cp /root/.ssh/authorized_keys "/home/${DEPLOY_USER}/.ssh/authorized_keys"
        echo "SSH authorized_keys disalin dari root ke ${DEPLOY_USER}"
    else
        echo "SSH authorized_keys ${DEPLOY_USER} sudah ada — skip copy"
    fi
    chown -R "${DEPLOY_USER}:${DEPLOY_USER}" "/home/${DEPLOY_USER}/.ssh"
    chmod 700 "/home/${DEPLOY_USER}/.ssh"
    chmod 600 "/home/${DEPLOY_USER}/.ssh/authorized_keys"
fi

echo "==> [3/6] Harden SSH (port=${SSH_PORT}, key-only, PermitRootLogin prohibit-password)"
SSHD_CONF=/etc/ssh/sshd_config

# Back up original config
cp "$SSHD_CONF" "${SSHD_CONF}.bak.$(date +%s)"

sed -i "s/^#\?Port .*/Port ${SSH_PORT}/"                   "$SSHD_CONF"
# prohibit-password: root login hanya via key (tidak via password), lebih aman dari 'no' untuk emergency
sed -i "s/^#\?PermitRootLogin .*/PermitRootLogin prohibit-password/" "$SSHD_CONF"
sed -i "s/^#\?PasswordAuthentication .*/PasswordAuthentication no/"  "$SSHD_CONF"
sed -i "s/^#\?PubkeyAuthentication .*/PubkeyAuthentication yes/"     "$SSHD_CONF"
sed -i "s/^#\?ChallengeResponseAuthentication .*/ChallengeResponseAuthentication no/" "$SSHD_CONF"
sed -i "s/^#\?X11Forwarding .*/X11Forwarding no/"          "$SSHD_CONF"
sed -i "s/^#\?MaxAuthTries .*/MaxAuthTries 3/"              "$SSHD_CONF"
sed -i "s/^#\?LoginGraceTime .*/LoginGraceTime 20/"         "$SSHD_CONF"

# Ensure AllowUsers exists — restrict to deployer + root (key-only)
grep -q "^AllowUsers" "$SSHD_CONF" \
    && sed -i "s/^AllowUsers .*/AllowUsers ${DEPLOY_USER} root/" "$SSHD_CONF" \
    || echo "AllowUsers ${DEPLOY_USER} root" >> "$SSHD_CONF"

systemctl restart ssh
echo "SSH hardened. Port: ${SSH_PORT} | PermitRootLogin: prohibit-password | PasswordAuthentication: no"

echo "==> [4/6] Configure fail2ban"
cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime  = 3600
findtime = 600
maxretry = 5
backend  = systemd

[sshd]
enabled  = true
port     = ${SSH_PORT}
logpath  = %(sshd_log)s
maxretry = 3
bantime  = 86400
EOF

systemctl enable fail2ban
systemctl restart fail2ban
echo "fail2ban configured (SSH ban after 3 failures for 24 h)"

echo "==> [5/6] Enable automatic security updates"
cat > /etc/apt/apt.conf.d/50unattended-upgrades << 'EOF'
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF

cat > /etc/apt/apt.conf.d/20auto-upgrades << 'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
EOF

systemctl enable unattended-upgrades
echo "Automatic security updates enabled"

echo "==> [6/6] Set deployment directory permissions"
DEPLOY_DIR="/opt/suberes"
mkdir -p "$DEPLOY_DIR"
chown -R "${DEPLOY_USER}:${DEPLOY_USER}" "$DEPLOY_DIR"
chmod 750 "$DEPLOY_DIR"

echo ""
echo ""
echo "════════════════════════════════════════════════════════════════"
echo " VPS hardening complete."
echo " Langkah selanjutnya:"
echo "   1. Jalankan firewall:"
echo "      sudo SSH_PORT=${SSH_PORT} bash deploy/setup-firewall.sh"
echo "════════════════════════════════════════════════════════════════"
