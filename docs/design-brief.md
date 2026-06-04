# Unic-Tunnel — Design Brief

> Self-contained brief to hand to a designer (incl. a "Claude design" session). Designs the
> **two surfaces**: the **Panel** (admin web dashboard on a VPS) and the **App** (Windows
> desktop client). Repo for both: https://github.com/UnicTunnel

---

## 1. Product in one paragraph

**Unic-Tunnel** is a self-hosted SSH-tunnel product for people in Iran (and similar
filtered regions) who want to route their whole device's internet through their own VPS — without
the technical clutter of v2ray/xray panels. The admin provisions hardened *tunnel-only* SSH
users on the panel and shares a single branded **`unic://…`** link with each person; the
person pastes it into the app, hits **On**, and the whole laptop tunnels through the server.
That's the entire interaction.

**The differentiator is simplicity.** Not more protocols, not more knobs — fewer.

---

## 2. Audience & context

- **Users (the App)** — non-technical Iranians. May not speak English well. Mostly Windows
  on the laptop, Android on the phone (mobile is a later phase). Connecting under stress
  (filtering tightens unexpectedly). Light on patience; "one big button" is the win.
- **Admins (the Panel)** — slightly more technical: someone who already bought a VPS and
  shares access with friends/family/customers. May resell. Wants a clean dashboard, not a
  CLI.
- **Region:** Iran primarily. Slow connections, unreliable networks, RTL (right-to-left)
  language readers.

## 3. Tone & feel

- **Calm, trustworthy, quiet.** This is a tool people use during anxious moments — not a
  product to be loud or playful. No "rocket" iconography, no celebratory shouting.
- **Modern, minimal, generous whitespace.** Think the polished side of fintech / system
  utilities: Tailscale, 1Password, Cloudflare Zero Trust, Hetzner Cloud. *Not* a "gamer VPN".
- **Sober colors.** Dark mode first (Iranian users often work at night, low light); a clean
  light mode too. Avoid alarm-red as the primary; reserve red strictly for danger/error.
- **Branding is light.** A logotype + one symbol. No badges, no skins, no "lifetime deal"
  energy.

## 4. Brand basics (assumed — adjust if you have other intent)

- **Name:** Unic-Tunnel.
- **Tagline (placeholder):** "Your internet, through your server."
- **Vibe words:** quiet, dependable, fast, mine.
- **Symbol idea (open):** the letter U made of two parallel lines suggesting a tunnel; or a
  single bracket `{ }` enclosure suggesting "your own space." Designer's call — propose.

## 5. Hard constraints

- **Bilingual: English + Persian (Farsi).** Full **RTL** layout for fa-IR. All copy must
  flow without distortion when mirrored.
- **Accessibility:** WCAG AA contrast minimum; keyboard navigation in the panel; large hit
  targets in the app.
- **Light AND dark themes** required. Dark mode is the default.
- **No animations on the connect button longer than ~400ms.** It must feel responsive even on
  weak machines.
- **No third-party fonts that need download at runtime.** Bundle them.
- **No tracking pixels, no analytics SDKs in the App.** This is a tool people use to escape
  surveillance — design must respect that.
- **Stay within the data the panel actually has.** Don't design screens for "user location",
  "ad revenue", "AI recommendations" or other invented features.

---

## 6. The Panel (web dashboard)

### Surface
A web app served by a single Go binary running on the VPS. Self-hosted. The only thing the
admin does is **manage tunnel users and see who's using how much**.

### Screens to design

1. **Login** — admin only. Single password field. Quiet, no marketing. Show "this is your
   panel at `panel.yourdomain.tld`" so the admin knows where they are.

2. **Dashboard / Users list** (the main screen)
   - Table of tunnel users: name, status (active / revoked), data used, last connected,
     created date. Per row: copy-link button, revoke button.
   - Big primary action: **Create user**.
   - Tiny header strip: server name/IP, sshd port, panel version, "Connected: N now."
   - Empty state ("No users yet — create one") must feel intentional, not blank.

3. **Create user** — modal or side-panel.
   - Inputs: friendly *name* (e.g. "Sara — laptop"), optional *traffic cap*, optional *expiry
     date*.
   - On submit, the result view shows the generated **`unic://…`** link, a **QR code**, and
     a **"Copy link"** button. That link is the only artifact — make it the hero.
   - Include a "share safely" hint (the link contains the password — share over a private
     channel).

4. **User detail** — name, status, traffic chart over time (last 7 / 30 days), recent
   sessions (anonymized: timestamp + duration; **do not show user IPs**), rotate-password,
   revoke, delete.

5. **Settings** — server info (read-only host/port), change admin password, panel
   appearance (light/dark/system), language (English / فارسی).

### Microcopy direction
- Buttons are verbs: "Create user", "Revoke", "Copy link".
- Errors are short and actionable: "Couldn't reach the server. Check sshd is running."
- No marketing voice. The admin already bought in.

---

## 7. The App (Windows desktop client)

### Surface
A small Flutter app. **Single window**, ~420×640 (resize-able but not designed to grow).
Optional system-tray icon. Runs sing-box quietly in the background.

### Screens / states to design

1. **First run — paste a link.** Big input field, single primary button: **"Connect"**.
   That's it. A short helper line explaining what the link is. No accounts, no signup, no
   email.

2. **Idle (link saved, not connected)** — shows the saved server name, region/flag if known,
   and one big **On** button. Status pill: *Disconnected*.

3. **Connecting** — the On button transitions to a loading state. Show what's happening in
   one human line ("Securing your connection…"). Don't expose `dial tcp` errors verbatim.

4. **Connected** — the button is **On**, button color shifts to the "active" accent. Show:
   - your masked exit IP ("xxx.xxx.•.•") + region;
   - upload / download counters;
   - a small **Off** action;
   - a subtle "your traffic is now going through your server" reassurance.

5. **Error states** — server unreachable, auth failed, admin permission denied (TUN needs
   admin), kill-switch tripped. Each: one cause line + one action button. Never a stack
   trace.

6. **Settings** (gear icon) — language, theme, start-on-boot, **kill-switch** toggle
   (default ON), "forget this link", "about" (version + the bundled sing-box version).

7. **System tray menu** (Windows) — Connect / Disconnect, current state, Open app, Quit.

### Anti-patterns to avoid
- A "country picker" — there's only one server per link.
- Speed-test bars, ping graphs, "boost" buttons.
- Ads, news feeds, recommendations.
- Modal nag screens about updates while connected.

---

## 8. Deliverables we need

1. **Logotype** + 1 symbol. SVG. Dark + light variants. Favicons / app icons.
2. **Color tokens** (CSS variables / Flutter `ThemeData`): primary, accent, success, warning,
   danger, surface (light + dark), text (primary/secondary/disabled).
3. **Typography scale** — bundled fonts (suggest: Inter for English, Vazirmatn for Persian).
4. **Components** — button, input, modal, table row, status pill, toast.
5. **Screens** (high-fidelity, light + dark + RTL where relevant):
   - Panel: Login, Users list (with empty state), Create-user, User detail, Settings.
   - App: First run, Idle, Connecting, Connected, an Error variant, Settings, Tray menu.
6. **Microcopy sheet** — every visible string in English + Persian.

Format: Figma is fine. PNG/PDF exports + a tokens JSON also accepted. Hand off via the repo
under `design/` (we'll create the folder), or a shared Figma link in this doc.

---

## 9. Out of scope (deliberately)

- Marketing site, pricing page, app-store listings.
- iOS / macOS / Linux clients.
- Any UI for VLESS / Reality / Hysteria / xray — v1 is **SSH only**.
- Reseller / sub-admin roles.
- In-app payments.

---

## 10. References (vibes, not "copy this")

- Tailscale (clarity, dark mode, tray app)
- 1Password (calm, trustworthy, generous spacing)
- Cloudflare Zero Trust (admin dashboard density done right)
- Hetzner Cloud Console (admin UX without bloat)
- Linear (microcopy + empty states)

---

## 11. Open questions for the designer

1. Should the App support a **list of multiple links** (laptop + travel server) in v1, or
   strictly one at a time? (Engineer leans: one at a time for v1, list later.)
2. Should the Panel show user **passwords** anywhere after creation, or only the link?
   (Engineer leans: link only — make it copy-once with a "rotate password" action.)
3. Persian-first or English-first default? (Engineer leans: detect from OS, default Persian
   if locale is `fa-IR`.)
