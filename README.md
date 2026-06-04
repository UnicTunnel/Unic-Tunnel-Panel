# Unic-Tunnel-Panel

Light **Go web dashboard** that runs on a VPS, provisions *tunnel-only* SSH users, and generates
`unic://` connection links for the Unic-Tunnel desktop app.

**Status:** foundation (design + docs). No code yet.

## Docs
- Architecture & how it works: [`docs/architecture.md`](docs/architecture.md)
- Link format: [`docs/unic-link-spec.md`](docs/unic-link-spec.md)
- VPS setup runbook: [`docs/vps-setup-runbook.md`](docs/vps-setup-runbook.md)
- Agent memory / conventions: [`CLAUDE.md`](CLAUDE.md)

## Stack (planned)
Go single binary · embedded web UI · SQLite. Builds for linux/amd64 and runs on the VPS.

## Companion
Client app: **[Unic-Tunnel-App](https://github.com/Unic-Tunnel/Unic-Tunnel-App)**
