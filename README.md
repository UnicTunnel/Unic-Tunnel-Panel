# Unic-Tunnel-Panel

Light **Go web dashboard** that runs on a VPS, provisions *tunnel-only* SSH users, and generates
`unic://` connection links for the Unic-Tunnel desktop app.

## Install on a VPS (one-liner)

SSH into your VPS as root, then:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/UnicTunnel/Unic-Tunnel-Panel/main/scripts/install.sh)
```

It downloads the binary, sets up the `unicpanel` service user, the scoped sudoers entry, the
systemd unit, asks for your VPS host/port + admin password, and starts the service. The URL
and admin credentials are printed at the end.

**Prerequisite:** apply the VPS setup runbook *first* so the `tunnelers` group and sshd
hardening are in place. See [`docs/vps-setup-runbook.md`](docs/vps-setup-runbook.md).

## Manage the service

```bash
systemctl status unic-panel
journalctl -u unic-panel -f
systemctl restart unic-panel
```

## Local development (Windows or Linux)

```bash
git clone https://github.com/UnicTunnel/Unic-Tunnel-Panel.git
cd Unic-Tunnel-Panel
go test ./...

# Run with the Stub provisioner (no real user creation, for UI dev):
UNIC_SSH_HOST=198.105.115.89 UNIC_SSH_PORT=2222 \
UNIC_ADMIN_PASSWORD=changeme \
UNIC_PROVISIONER=stub \
go run ./cmd/panel
# → http://127.0.0.1:8080
```

## Docs
- Architecture & how it works: [`docs/architecture.md`](docs/architecture.md)
- Link format: [`docs/unic-link-spec.md`](docs/unic-link-spec.md)
- VPS setup runbook: [`docs/vps-setup-runbook.md`](docs/vps-setup-runbook.md)
- Agent memory / conventions: [`CLAUDE.md`](CLAUDE.md)

## Stack
Go single binary · embedded web UI via `go:embed` · SQLite (pure-Go, no CGO). Builds for
linux/amd64 + linux/arm64; runs on the VPS.

## Companion
Client app: **[Unic-Tunnel-App](https://github.com/Unic-Tunnel/Unic-Tunnel-App)**
