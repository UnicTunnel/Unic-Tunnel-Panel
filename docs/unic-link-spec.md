# `unic://` link spec (v1)

The connection link the **Panel generates** and the **App decodes**. Canonical for both repos.

## Format
```
unic://<base64url(json)>
```
- The `unic://` scheme followed by a single base64url-encoded JSON object.
- Keep it on one line and copy-pasteable. Decoders should accept the payload with or without
  base64 `=` padding.

## v1 payload (SSH profile)
```json
{ "v": 1, "name": "Berlin-1", "host": "1.2.3.4", "port": 22,
  "user": "u_8f3a", "password": "<generated>" }
```

| field | type | meaning |
|---|---|---|
| `v` | int | format version. `1` = SSH profile. **Required.** |
| `name` | string | display label shown in the app |
| `host` | string | VPS IP or hostname |
| `port` | int | SSH port (default `22`) |
| `user` | string | the tunnel-only SSH username |
| `password` | string | the generated SSH password |

## How the app uses it
Decode base64url → parse JSON → check `v` → build a sing-box **`ssh` outbound** plus a **`tun`
inbound** (whole-device), then run sing-box. Sketch of the outbound:
```json
{ "type": "ssh", "tag": "proxy",
  "server": "<host>", "server_port": <port>,
  "user": "<user>", "password": "<password>" }
```

## Growth seam (NOT in v1 — keeps old links working)
The `v` field plus an optional `proto` field let the format grow without breaking existing links:
- `proto: "ssh"` — implicit in v1.
- `proto: "vless-reality"` — later, carrying `uuid/pbk/sni/sid/fp/flow` (see the plan's "Later
  options").
Apps must **reject an unknown `v`** with a clear "please update the app" message rather than
guessing.

## Encoding & safety notes
- Use **base64url** (`-`/`_`, not `+`/`/`) so the link is URL-safe.
- v1 is **not encrypted**: anyone holding the link can use (and identify) the tunnel. Optional
  link encryption is a "Later" item. Treat links as secrets and share them over a trusted
  channel.
