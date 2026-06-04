# VPS setup runbook (v1 — SSH tunnel)

Run this **once** when you buy the VPS. It produces a server that can host tunnel-only SSH
accounts (no shell, no files, forward-only). Assumes Ubuntu 22.04/24.04 and a root/sudo shell.
Replace `SERVER_IP` and other CAPS placeholders.

> ⚠️ Do step 2 (keep your admin key login working) **before** you change password settings, or
> you can lock yourself out.

## 0. Pick a VPS
- A provider with good international bandwidth; avoid IP ranges already blocked in Iran.
- A small box is fine (1 vCPU / 1 GB RAM). You need root SSH access.

## 1. First login + update
```bash
ssh root@SERVER_IP
apt update && apt upgrade -y
```

## 2. Keep YOUR admin login safe (key-based)
Ensure key-based admin access works **now**, because we'll disable passwords for everyone except
the tunnel group:
```bash
mkdir -p ~/.ssh && chmod 700 ~/.ssh
# paste your PUBLIC key into authorized_keys:
echo "ssh-ed25519 AAAA... you@host" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

## 3. Create the tunnel-only group
```bash
groupadd tunnelers
```

## 4. Harden sshd for the tunnel group
Create a drop-in `/etc/ssh/sshd_config.d/10-tunnelers.conf`:
```
# Admin stays key-only (global default)
PasswordAuthentication no

Match Group tunnelers
    PasswordAuthentication yes
    AllowTcpForwarding yes
    X11Forwarding no
    PermitTTY no
    AllowAgentForwarding no
    PermitTunnel no
    ForceCommand /usr/sbin/nologin
```
Why this is "tunnel-only":
- `ForceCommand /usr/sbin/nologin` + `PermitTTY no` → no shell, no commands.
- sing-box only needs `AllowTcpForwarding yes` (it opens direct-tcpip channels and never asks for
  a shell), so the account can tunnel and **nothing else**.
- No SFTP subsystem is granted to this group → **no file access**.
- *(Optional, stronger)* limit destinations with `PermitOpen`, and/or firewall the tunnel users
  away from `127.0.0.0/8` and private ranges so they can't reach the VPS's own services.

Apply:
```bash
sshd -t && systemctl reload ssh
```

## 5. Create a tunnel user (manual test / fallback)
The Panel automates this; to test by hand:
```bash
PW=$(openssl rand -base64 18)
useradd -m -g tunnelers -s /usr/sbin/nologin u_test
echo "u_test:$PW" | chpasswd
echo "username: u_test   password: $PW"
```
Verify it can tunnel but not shell:
```bash
ssh -N -L 9999:example.com:80 u_test@SERVER_IP   # should connect & forward
ssh u_test@SERVER_IP                              # should be refused a shell
```
Revoke later:
```bash
usermod -L u_test     # lock (disable)
# or
userdel -r u_test     # delete entirely
```

## 6. Brute-force protection
```bash
apt install -y fail2ban
systemctl enable --now fail2ban
```
*(Optionally move SSH off port 22 and/or rate-limit new connections.)*

## 7. Firewall
```bash
ufw allow OpenSSH        # or your custom SSH port
ufw enable
```

## 8. Let the Panel provision users (scoped sudo)
The Panel runs as a **non-root** service user (e.g. `unicpanel`) and manages tunnel accounts
through one small script — never as full root.

`/usr/local/sbin/tunnel-user.sh` (root-owned, `0755`):
```bash
#!/usr/bin/env bash
set -euo pipefail
case "$1" in
  create) useradd -m -g tunnelers -s /usr/sbin/nologin "$2"; echo "$2:$3" | chpasswd ;;
  passwd) echo "$2:$3" | chpasswd ;;
  lock)   usermod -L "$2" ;;
  delete) userdel -r "$2" ;;
  *) echo "usage: tunnel-user.sh {create|passwd|lock|delete} <user> [pass]" >&2; exit 2 ;;
esac
```
`/etc/sudoers.d/unicpanel` (check with `visudo -cf /etc/sudoers.d/unicpanel`):
```
unicpanel ALL=(root) NOPASSWD: /usr/local/sbin/tunnel-user.sh
```
Now the panel can create/lock/delete tunnel users but **cannot** run arbitrary root commands.

## 9. Make the `unic://` link
For the manual test, build the link from the user you created:
```
payload = {"v":1,"name":"Server-1","host":"SERVER_IP","port":22,"user":"u_test","password":"<PW>"}
link    = "unic://" + base64url(payload)
```
The Panel does this automatically. Format details: `docs/unic-link-spec.md`.

## Later (not v1)
If plain SSH gets throttled by DPI, install sing-box **on the server** and switch to
VLESS+Reality. The app already runs sing-box, so this is a new `proto` value in the link — not a
new app. See the plan's "Later options".
