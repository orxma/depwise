# 💎 ORX TUNNEL BOT — GO EDITION

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/OS-Ubuntu%2024.04-E95420?style=for-the-badge&logo=ubuntu" alt="Ubuntu 24.04">
  <img src="https://img.shields.io/badge/Platform-Linux-FCC624?style=for-the-badge&logo=linux" alt="Linux">
  <img src="https://img.shields.io/badge/Arch-amd64%20%7C%20arm64-lightgrey?style=for-the-badge" alt="Architecture">
  <img src="https://img.shields.io/badge/Status-Stable-success?style=for-the-badge" alt="Status">
</p>

---

## 🚀 What is ORX TUNNEL Bot?

**ORX TUNNEL Bot (Go Edition)** is an all-in-one, high-performance manager for VPN servers and SSH
accounts, driven entirely from Telegram. Written in **Go** for speed, stability and a low memory
footprint, it turns a plain VPS into an automated control panel.

The bot interface is **100% English**.

---

## 📥 Installation

> [!NOTE]
> **OS support:** developed and tested on **Ubuntu 24.04**. Use that release (or a derivative) so all
> dependencies behave (Go, systemd, SSH, Xray, SlowDNS, VayDNS, Slipstream, dnsdist).
>
> **Architecture:** both `amd64` and `arm64` (aarch64) are supported. The installer detects the host
> architecture and fetches the matching Go toolchain.

The bot is installed from the **ORX TUNNEL panel**, not from this repository directly:

```bash
bash /etc/orx-tunnel/tools/orxtunnel.sh
```

Choose option **1. Install / Update Bot**. You will be asked for:

| Prompt | Value |
| :--- | :--- |
| `TOKEN` | The bot token from [@BotFather](https://t.me/BotFather) |
| `Chat ID` | Your numeric Telegram ID (the SuperAdmin) |

Credentials are stored in `/opt/orxtunnel_bot/.env` (mode `600`) and reused on later runs.

### Building manually

```bash
git clone https://github.com/orxma/depwise.git
cd depwise
go mod tidy
go build -o /usr/local/bin/orxtunnel-bot cmd/orxtunnel/main.go
```

---

## 🔄 Updating

Re-run the installer and pick option **1**. Your database and configuration are preserved.

```bash
bash /etc/orx-tunnel/tools/orxtunnel.sh
```

The bot can also update itself from **⚙️ Pro Settings → 🔄 Update System**, which pulls this
repository, rebuilds, and restarts `orxtunnel.service`.

---

## 🧩 Service layout

| Item | Path |
| :--- | :--- |
| Binary | `/usr/local/bin/orxtunnel-bot` |
| Service | `orxtunnel.service` |
| Config / database | `/opt/orxtunnel_bot/` |
| SSH banners | `/etc/ssh_banners/` |
| Dropbear banner | `/etc/orxtunnel-banner.txt` |

---

## ✨ Features

### 🛠️ Protocol management (all-in-one)
- **Advanced DNS multiplexing:** using `dnsdist` plus kernel-level `U32` (Netfilter) filters, the bot
  runs **SlowDNS, VayDNS and Slipstream** concurrently on the same UDP port 53 without conflicts.
  Native support for mobile networks (IPv6 / NAT64).
- **SlowDNS / Noiz DNS:** highly stable, block-resistant DNS tunnels.
- **Slipstream:** advanced hybrid protocol with a native QUIC path on port 443.
- **SSH / Dropbear / WS TLS HTTP:** full account management with connection limits and HTML banners
  generated on the fly.
- **Xray (VMess):** VMess over WebSocket, compatible with Cloudflare and HAProxy.
- **ZiVPN & UDP Custom:** UDP gaming protocols and robust bypass (range `6000:19999`).
- **Falcon Proxy & ProxyDT:** optimized HTTP proxies.

### 🛡️ Administration
- **Root SSH auto-configuration:** permanently enables root SSH access on cloud VPS images.
- **Reboot resilience:** rebuilds networking, iptables, u32 rules, IPv6 and dnsdist on every boot so
  no protocol silently dies.
- **Broadcast:** send announcements to every registered user.
- **Live monitoring:** VPS metrics (cores, RAM, disk, uptime) and active protocols.
- **Bans and quotas:** per-role limits on account count, duration and devices.
- **Expiry alerts:** admins are notified 1 day and 1 hour before an SSH, ZiVPN or Xray account expires.

### 🧹 Maintenance
- **Durable state:** traffic counters and settings survive reboots.
- **Service self-healing:** automatic recovery of HAProxy, Xray and DNS.
- **HAProxy auto-recovery:** verifies HAProxy is running, kills port squatters, restarts when needed.

---

## 💸 Monetization (Monetag + Vercel)

The bot ships with a native monetization flow. When enabled, public users must watch a Rewarded
Interstitial ad before **creating** or **renewing** an account. Setup is driven by an interactive
wizard inside the bot.

### Step-by-step

**1. Create a Monetag account**
- Register as a publisher at [Monetag.com](https://monetag.com).
- Add a new application and select the **Telegram Mini App** format.
- Provide your bot username (e.g. `@orxtunnel_bot`).
- Create a **Rewarded Interstitial** ad block.

**2. Collect the two code snippets**
- **SDK script:** a `<script>` tag, e.g.
  `<script src='//libtl.com/sdk.js' data-zone='1234567' data-sdk='show_1234567'></script>`
- **Rewarded block:** a JavaScript block, e.g. `show_1234567().then(() => { ... })`

**3. Run the in-bot wizard**
- Open **⚙️ Pro Settings → ⚙️ Configure MiniApp Ads**.
- Paste the `<script>` tag when asked.
- Paste the **Rewarded** block when asked.
- The bot builds the HTML for you and returns a `monetag_miniapp.zip`. Download it, extract it, and
  use the `.html` file inside.

**4. Host the Mini App**
- Deploy the extracted `monetag_miniapp.html` to [Vercel](https://vercel.com/),
  [GitHub Pages](https://pages.github.com/) or any static host.
- Copy the public direct link to the HTML file
  (e.g. `https://my-miniapp.vercel.app/monetag_miniapp.html`).

**5. Activate**
- Send that link to the bot as the final wizard step. The ad wall is enabled automatically.

---

## ☁️ Native backups (Telegram)

Backups are delivered straight to the admin chat — **on demand** from the panel, or **automatically**
every 1, 3, 7 or 30 days. No external service required. Restoring is done by sending the `.json`
backup file back to the bot.

---

## 🛠️ Troubleshooting

| Symptom | Likely cause | Fix |
| :--- | :--- | :--- |
| **No connection on mobile data** | NS domain has no IP, or IPv6 is failing | Reinstall the protocols from the Protocols menu |
| **Bot not responding** | Process hung | `systemctl restart orxtunnel` |
| **Xray / VMess not connecting** | HAProxy or Xray did not start | `systemctl status haproxy xray` |
| **Automatic backup failing** | Invalid chat ID | Reconfigure the backup interval in the menu |
| **`Exec format error`** | Binary built for the wrong CPU architecture | Reinstall; the installer selects the correct `amd64` / `arm64` build |

Useful commands:

```bash
systemctl status orxtunnel      # service state
journalctl -u orxtunnel -f      # live logs
```

---

## 💎 Credits and support

- **📢 Channel:** [@orxtunnel](https://t.me/orxtunnel)
- **🤖 Bot:** [@orxtunnel_bot](https://t.me/orxtunnel_bot)

Built on the Go rewrite of the original Telegram VPN manager, rebranded and maintained as
**ORX TUNNEL**.

---

<p align="center">
  <i>Powering your VPS with the speed of Go and the reach of the kernel.</i><br>
  <b>© 2026 ORX TUNNEL Project</b>
</p>
