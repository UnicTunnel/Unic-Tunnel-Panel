# Unic-Tunnel-Panel — Claude Code memory

> Server-side **web dashboard** for Unic-Tunnel. Mints locked-down, *tunnel-only* SSH users on a
> VPS and emits a `unic://` connection link for the desktop app to decode. Companion repo:
> **Unic-Tunnel-App** (Flutter client). Canonical design docs live in `docs/`.

## What this is (one paragraph)
A small, light, good-looking admin panel that runs **on the VPS**. The admin creates a user →
the panel provisions one hardened Linux SSH account (username + password) that can **only
forward TCP** (no shell, no files) and returns a `unic://` link to share. "Revoke" disables that
account. v1 transport is a plain **SSH tunnel**; the heavy lifting on the client is done by
sing-box. See `docs/architecture.md`.

## Stack & why
- **Go** (single static binary) — lightest possible deploy on a cheap VPS: one file, tiny RAM.
- **Embedded web UI** — SPA built ahead of time and embedded via Go `embed`; served by the same
  binary (no separate web server).
- **SQLite** — zero-ops local DB (servers, users, links, admin session).
- Keep dependencies minimal: stdlib + a small router; avoid heavy frameworks.

## Layout (target — not scaffolded yet)
- `cmd/panel/` — main entrypoint (HTTP server, embeds the built `web/`).
- `internal/server/` — VPS records (host, the admin creds used for provisioning).
- `internal/accounts/` — SSH user provisioning: create / set-password / lock / delete, via a
  **scoped sudoers script** (v1 = local OS on the same VPS). See Security.
- `internal/links/` — builds the `unic://` link from an account (`docs/unic-link-spec.md`).
- `internal/store/` — SQLite access.
- `web/` — frontend source + embedded build output.
- `scripts/tunnel-user.sh` — the provisioning helper invoked through sudoers.

## Build / run / test (fill in once scaffolded)
- Build (local): `go build ./cmd/panel`   ·   Test: `go test ./...`
- Run (dev): `go run ./cmd/panel`
- Ship to VPS: `GOOS=linux GOARCH=amd64 go build -o unic-panel ./cmd/panel`
- Build the frontend before `go build` so `embed` picks up the latest assets.

## Security rules (hard constraints)
- **Never log secrets** — SSH passwords, admin VPS creds. Redact in logs and error messages.
- Provision users through a **tightly-scoped `sudoers`** entry that allows ONLY
  `useradd`/`chpasswd`/`usermod`/`userdel` for the `tunnelers` group. **Never run the panel as
  full root.** Details in `docs/vps-setup-runbook.md`.
- Generated SSH passwords must be **strong + random** (≥20 chars from a CSPRNG).
- Every tunnel account is **tunnel-only** (no shell, no SFTP, forward-only); the hardening lives
  in `sshd_config`. The panel only creates accounts — do not weaken the hardening.
- Panel is **admin-only**: auth on every route. The only optional unauthenticated route is a
  link-fetch endpoint gated by an unguessable token (consider for later, not v1).
- Build shell commands with **arg arrays, never string interpolation**; validate every value
  that reaches the OS.

## The unic:// contract
`unic://<base64url(json)>` where JSON = `{ v, name, host, port, user, password }`. Canonical:
`docs/unic-link-spec.md`. Keep the `v` version field — panel and app must agree on it.

## Dangerous areas (take care / escalate)
- Anything touching `sshd_config`, the `tunnelers` group, or the sudoers script: a mistake can
  lock the admin out of the VPS or open a hole. Change deliberately and update the runbook.
- Password generation/handling and the provisioning shell calls.

## Out of scope for v1 (see the plan's "Later options")
Reality/VLESS fallback, link encryption, `vless://` import, multi-server rotation. Do **not**
build these into v1 — keep the panel simple.

## Notes
Repo is currently empty (foundation pass). Go 1.26 is installed locally; the panel targets
linux/amd64 for the VPS. Keep this file lean — let auto-memory accumulate build/debug specifics.
