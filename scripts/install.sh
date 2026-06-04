#!/usr/bin/env bash
# Unic-Tunnel Panel installer.
# Usage (as root on the VPS):
#   bash <(curl -Ls https://raw.githubusercontent.com/UnicTunnel/Unic-Tunnel-Panel/main/scripts/install.sh)

set -euo pipefail

REPO="UnicTunnel/Unic-Tunnel-Panel"
BIN_NAME="unic-panel"
INSTALL_PREFIX="/usr/local/bin"
SBIN_PREFIX="/usr/local/sbin"
DATA_DIR="/var/lib/unic-panel"
ENV_FILE="/etc/unic-panel.env"
SUDOERS_FILE="/etc/sudoers.d/unicpanel"
SCRIPT_PATH="$SBIN_PREFIX/tunnel-user.sh"
SERVICE_FILE="/etc/systemd/system/unic-panel.service"
SERVICE_USER="unicpanel"

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
blue()  { printf '\033[34m%s\033[0m\n' "$*"; }
die()   { red "ERROR: $*"; exit 1; }

[ "$(id -u)" -eq 0 ] || die "run as root"
[ "$(uname -s)" = "Linux" ] || die "Linux only"

case "$(uname -m)" in
    x86_64|amd64) ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *) die "unsupported architecture: $(uname -m)" ;;
esac

# Latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
         | grep -oP '"tag_name": "\K[^"]+' || true)
[ -n "$LATEST" ] || die "no releases yet on $REPO. Create one first (push a v* tag)."
blue "→ Installing $REPO $LATEST (linux/$ARCH)"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Download binary + scripts
BIN_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BIN_NAME}-linux-${ARCH}"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/${LATEST}/scripts"
curl -fsSL -o "$TMPDIR/$BIN_NAME"           "$BIN_URL"               || die "binary download failed"
curl -fsSL -o "$TMPDIR/tunnel-user.sh"      "$RAW_BASE/tunnel-user.sh"     || die "tunnel-user.sh"
curl -fsSL -o "$TMPDIR/sudoers.unicpanel"   "$RAW_BASE/sudoers.unicpanel"  || die "sudoers"
curl -fsSL -o "$TMPDIR/unic-panel.service"  "$RAW_BASE/unic-panel.service" || die "service file"

# tunnelers group + sshd hardening reminder
getent group tunnelers >/dev/null 2>&1 || groupadd tunnelers
if ! grep -rq 'Match Group tunnelers' /etc/ssh/sshd_config /etc/ssh/sshd_config.d/ 2>/dev/null; then
    red "⚠ sshd 'Match Group tunnelers' block not found."
    red "  Apply the runbook BEFORE creating tunnel users via the panel:"
    red "  https://github.com/${REPO}/blob/${LATEST}/docs/vps-setup-runbook.md"
    echo
fi

# Service user
id "$SERVICE_USER" >/dev/null 2>&1 || \
    useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"

# Install binary + scripts
install -m 0755 "$TMPDIR/$BIN_NAME"          "$INSTALL_PREFIX/$BIN_NAME"
install -m 0750 -o root -g "$SERVICE_USER" \
        "$TMPDIR/tunnel-user.sh"             "$SCRIPT_PATH"
install -m 0440 -o root -g root \
        "$TMPDIR/sudoers.unicpanel"          "$SUDOERS_FILE"
visudo -cf "$SUDOERS_FILE" >/dev/null || { rm -f "$SUDOERS_FILE"; die "sudoers validation failed"; }

install -d -m 0750 -o "$SERVICE_USER" -g "$SERVICE_USER" "$DATA_DIR"

# First-time interactive config
if [ ! -f "$ENV_FILE" ]; then
    blue "→ First-time config (write $ENV_FILE)"
    DEFAULT_HOST=$(hostname -I 2>/dev/null | awk '{print $1}')
    read -r -p "VPS public IP or hostname [${DEFAULT_HOST}]: " SSH_HOST
    SSH_HOST="${SSH_HOST:-$DEFAULT_HOST}"
    [ -n "$SSH_HOST" ] || die "SSH host is required"

    read -r -p "SSH port [22]: " SSH_PORT
    SSH_PORT="${SSH_PORT:-22}"

    read -r -p "Panel bind address [0.0.0.0:8080]: " BIND
    BIND="${BIND:-0.0.0.0:8080}"

    read -r -p "Admin password (empty = auto-generate): " ADMIN_PW
    if [ -z "$ADMIN_PW" ]; then
        ADMIN_PW=$(openssl rand -base64 18 2>/dev/null \
                   || head -c 18 /dev/urandom | base64)
        ADMIN_PW=$(printf '%s' "$ADMIN_PW" | tr -d '\n')
        GENERATED=1
    fi

    umask 077
    cat > "$ENV_FILE" <<EOF
UNIC_BIND=$BIND
UNIC_DB=$DATA_DIR/unic-panel.db
UNIC_ADMIN_PASSWORD=$ADMIN_PW
UNIC_SSH_HOST=$SSH_HOST
UNIC_SSH_PORT=$SSH_PORT
UNIC_PROVISIONER=sudo
UNIC_SUDO_SCRIPT=$SCRIPT_PATH
EOF
    chmod 0640 "$ENV_FILE"
    chown root:"$SERVICE_USER" "$ENV_FILE"
else
    blue "→ $ENV_FILE exists, keeping it"
fi

install -m 0644 "$TMPDIR/unic-panel.service" "$SERVICE_FILE"
systemctl daemon-reload
systemctl enable --now unic-panel.service
sleep 1

if systemctl is-active --quiet unic-panel.service; then
    BIND=$(grep ^UNIC_BIND= "$ENV_FILE" | cut -d= -f2-)
    HOST=$(grep ^UNIC_SSH_HOST= "$ENV_FILE" | cut -d= -f2-)
    DISPLAY="${BIND/0.0.0.0/$HOST}"
    echo
    green "✓ unic-panel started"
    green "  URL:      http://$DISPLAY"
    green "  Username: admin"
    if [ "${GENERATED:-0}" = "1" ]; then
        green "  Password: $ADMIN_PW"
        red   "  → SAVE THIS PASSWORD; it won't be shown again."
    else
        green "  Password: (the one you set)"
    fi
    echo
    blue "Manage:  systemctl {status|restart|stop} unic-panel"
    blue "Logs:    journalctl -u unic-panel -f"
else
    red "✗ unic-panel failed to start."
    red "  Check:  journalctl -u unic-panel -n 50 --no-pager"
    exit 1
fi
