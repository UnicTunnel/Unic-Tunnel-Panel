#!/usr/bin/env bash
# Unic-Tunnel: provisions locked-down tunnel-only SSH users.
# Invoked by the panel via sudoers (see /etc/sudoers.d/unicpanel).
#
# Usage: tunnel-user.sh {create|lock|delete} <username>
#   - create: password is read from stdin (NOT a command-line arg, so it doesn't
#             leak through `ps`).
#   - lock:   usermod -L (account disabled, can be re-enabled)
#   - delete: userdel -r (account removed; home dir too if it exists)

set -euo pipefail

TUNNEL_GROUP="tunnelers"

die() { echo "tunnel-user: $*" >&2; exit 1; }

ensure_group() {
    getent group "$TUNNEL_GROUP" >/dev/null 2>&1 || groupadd "$TUNNEL_GROUP"
}

action="${1:-}"
username="${2:-}"

[ -n "$action" ]   || die "missing action (create|lock|delete)"
[ -n "$username" ] || die "missing username"

# Defensive: only allow safe username characters.
case "$username" in
    *[!a-zA-Z0-9_]*) die "username contains forbidden characters: $username" ;;
esac

case "$action" in
    create)
        ensure_group
        useradd -m -g "$TUNNEL_GROUP" -s /usr/sbin/nologin "$username"
        # If anything below fails, undo the useradd so we don't orphan an
        # account without a password.
        trap "userdel -r '$username' >/dev/null 2>&1 || true" ERR
        # Use cat (not `read`) so we don't care whether the caller terminates
        # the password with a newline. $() strips trailing newlines.
        pw=$(cat)
        [ -n "$pw" ] || die "empty password on stdin"
        printf '%s:%s\n' "$username" "$pw" | chpasswd
        trap - ERR
        ;;
    lock)
        usermod -L "$username"
        ;;
    delete)
        userdel -r "$username" 2>/dev/null || userdel "$username"
        ;;
    *)
        die "unknown action: $action (want create|lock|delete)"
        ;;
esac
