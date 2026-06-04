# Unic-Tunnel-Panel — repo memory

> The workspace `CLAUDE.md` one level up loads automatically and holds the **shared
> rules** (role, security, scope, token efficiency). This file is repo-specific.

## Stack

Go (single static binary) · `net/http` stdlib · `modernc.org/sqlite` (pure-Go,
no CGO) · embedded web UI via Go `embed`. Targets linux/amd64 for the VPS.

## Planned layout (NOT all scaffolded — confirm scope before generating)

- `cmd/panel/` — entrypoint
- `internal/store/` — SQLite + queries (admins, sessions, tunnel_users)
- `internal/accounts/` — `Provisioner` interface + `Stub` (Windows dev) + `Sudo` (VPS)
- `internal/links/` — `unic://` builder + parser
- `internal/server/` — HTTP routes, handlers, session middleware
- `web/embed.go` + `web/templates/*.html` + `web/static/*`
- `scripts/tunnel-user.sh`, `scripts/sudoers.unicpanel`, `scripts/unic-panel.service`
- `docs/` — architecture, unic-link-spec, vps-setup-runbook, design-brief

Full diagram: `../docs/structure.md`.

## Build / run / test (will activate once scaffolded)

- Local with Stub provisioner:
  `UNIC_PROVISIONER=stub UNIC_SSH_HOST=198.105.115.89 UNIC_SSH_PORT=2222 \
   UNIC_ADMIN_PASSWORD=changeme rtk go run ./cmd/panel`
- VPS build (cross-compile from Windows):
  `$env:GOOS='linux'; $env:GOARCH='amd64'; rtk go build -o unic-panel ./cmd/panel`
- Tests: `rtk go test ./...`

## Provisioner contract

Two implementations of `accounts.Provisioner`:
- **`Stub`** — logs the call only. Used for local Windows dev where there's no sshd.
- **`Sudo`** — `sudo -n /usr/local/sbin/tunnel-user.sh {create|lock|delete} ...`.
  The script + sudoers entry are set up by `docs/vps-setup-runbook.md`.

The panel itself runs as a **non-root** service user (e.g. `unicpanel`); the scoped
sudoers entry whitelists ONLY the one script.

## Gotchas specific to this repo

- The VPS sshd runs on **port 2222**, not 22. Default `UNIC_SSH_PORT=22` is wrong
  for this server — always set `UNIC_SSH_PORT=2222` in env.
- `modernc.org/sqlite` registers driver name `"sqlite"` (not `"sqlite3"`).
- Generated tunnel-user passwords are stored **in plaintext in SQLite** because the
  panel must include them in the `unic://` link. The DB file MUST be `chmod 600` on
  the VPS. Documented limitation; encrypted-at-rest is a Later item.

## Scope reminder

v1 = SSH-only. No VLESS / Reality / Hysteria / xray. See workspace `CLAUDE.md`.
