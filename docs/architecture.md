# Architecture — Unic-Tunnel

Plain-language design for the whole system. This is the canonical reference; both repos point
here.

## The problem
A user in Iran wants to route **all** of a device's traffic through a VPS they control — a clean
"install, turn on, whole device tunneled" experience — without configs that are easy to
recognize and filter. **v1 uses a plain SSH tunnel** (works while filtering is relaxed). A
stronger disguise (Reality) is deliberately parked for later so v1 stays simple.

## Three pieces, one job each
1. **VPS** — a normal Ubuntu server. We use only its built-in SSH. A `tunnelers` group holds
   **tunnel-only** accounts: they can forward TCP but get no shell and see no files.
2. **Panel** (`Unic-Tunnel-Panel`, Go) — a light web dashboard on the VPS. "Create user" →
   provisions one locked-down SSH account (username + password) → returns a `unic://` link.
   "Revoke" disables the account.
3. **App** (`Unic-Tunnel-App`, Flutter) — paste the `unic://` link, press **On**.

## Where sing-box fits (the "engine")
The App is the steering wheel; **sing-box is the engine under the hood** (we run it unmodified).
On press-On, the app hands the SSH login to sing-box, which (a) opens the SSH connection to the
VPS and (b) captures all device traffic + DNS via a virtual TUN adapter and pushes it through
that connection. Plain SSH alone only makes a per-app proxy; the whole-device capture is the one
job sing-box does. On Windows we run it as a background `sing-box.exe` **sidecar** process — no
Dart↔Go FFI in v1.

## Diagram (v1)
```
  ADMIN                          USER'S DEVICE (Windows first)          VPS
┌──────────────────┐          ┌──────────────────────────────┐     ┌──────────────┐
│ Unic-Tunnel-Panel│  unic:// │ Unic-Tunnel-App (Flutter UI)  │ SSH │ sshd         │
│ Go + web UI      │ ───────► │  • decode unic:// link        │ usr │ tunnel-only  │
│  • create SSH    │  link    │  • write sing-box config      │ +pw │ users:       │
│    user + pass   │ (shared  │  • run sing-box.exe (sidecar) │────►│  no shell,   │
│  • make unic://  │  to user)│  • TUN = whole-device tunnel  │ TUN │  no files,   │
│  • revoke user   │          │  • big On/Off + stats         │     │  fwd only    │
└──────────────────┘          └──────────────────────────────┘     └──────────────┘
```

## Data flow
1. Admin clicks "create user" in the Panel.
2. Panel runs the scoped provisioning script → a new `tunnelers` SSH account with a random
   password.
3. Panel encodes `{v,name,host,port,user,password}` as `unic://<base64url(json)>`.
4. User pastes the link into the App.
5. App decodes it → writes a sing-box `config.json` (`ssh` outbound + `tun` inbound).
6. App launches `sing-box.exe` (elevated) → the whole device is tunneled. Off = stop the process.

## Why build fresh (not fork)
Existing tools (Hiddify, Marzban, 3x-ui) are great references but are built around the
xray/v2ray subscription model and/or carry restrictive licenses (Hiddify-App adds extra
conditions on forks). Our needs are narrower (SSH-user provisioning + a custom link) and our
stack choices differ (Go panel, Flutter app), so we build fresh and merely **use** sing-box as
the engine.

## Out of scope for v1
See the plan's "Later options": Reality/VLESS disguise, encrypted links, `vless://` import,
multi-server + IP rotation, and the Android client.
