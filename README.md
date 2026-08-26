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

The bot **IS the control panel** — no external panel required. Install the bot, then use Telegram
to install/manage all protocols (Xray, SSH WebSocket, HAProxy, DNS tunnels, etc.).

The bot interface is **100% English**.

---

## 📥 Installation

> [!NOTE]
> **OS support:** developed and tested on **Ubuntu 24.04**. Use that release (or a derivative) so all
> dependencies behave (Go, systemd, SSH, Xray, SlowDNS, VayDNS, Slipstream, dnsdist).
>
> **Architecture:** both `amd64` and `arm64` (aarch64) are supported.

### Option 1: Automated Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/orxma/depwise/main/install_depwise.sh | sudo bash
```

Or download and run manually:
```bash
wget https://raw.githubusercontent.com/orxma/depwise/main/install_depwise.sh
chmod +x install_depwise.sh
sudo ./install_depwise.sh
```

The script will prompt for:
| Prompt | Value |
| :--- | :--- |
| `BOT_TOKEN` | Token from [@BotFather](https://t.me/BotFather) |
| `SUPER_ADMIN` | Your numeric Telegram ID |

### Option 2: Manual Build

```bash
git clone https://github.com/orxma/depwise.git
cd depwise
export PATH=$PATH:/usr/local/go/bin
go build -o /usr/local/bin/depwise-bot ./cmd/orxtunnel

mkdir -p /opt/depwise_bot
cat > /opt/depwise_bot/.env <<EOF
BOT_TOKEN=your_bot_token_here
SUPER_ADMIN=your_telegram_id_here
EOF
chmod 600 /opt/depwise_bot/.env

# Create systemd service
cat > /etc/systemd/system/depwise.service <<EOF
[Unit]
Description=Depwise Telegram Bot (Go Edition)
After=network.target

[Service]
Type=simple
User=root
EnvironmentFile=/opt/depwise_bot/.env
Environment="GOMEMLIMIT=40MiB" "GOGC=20"
ExecStart=/usr/local/bin/depwise-bot
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable depwise
systemctl start depwise
```

---

## 🔄 Updating

```bash
cd /root/depwise && git pull
export PATH=$PATH:/usr/local/go/bin
go build -o /usr/local/bin/depwise-bot ./cmd/orxtunnel
systemctl restart depwise
```

Or use the in-bot updater: **⚙️ Pro Settings → 🔄 Update System**

---

## 🧩 Service Layout

| Item | Path |
| :--- | :--- |
| Binary | `/usr/local/bin/depwise-bot` |
| Service | `depwise.service` |
| Config / Database | `/opt/depwise_bot/` |
| SSH Banners | `/etc/ssh_banners/` |
| Dropbear Banner | `/etc/orxtunnel-banner.txt` |

---

## ✨ Features

### 🛠️ Protocol Management (All-in-One)
- **Xray (VMess / VLESS / Trojan):** Over WebSocket with HAProxy TLS termination on port 443
- **SSH / Dropbear / WebSocket:** Full account management with connection limits, HTML banners
- **SSH WebSocket SSL/Non-SSL:** HAProxy on 80/443/8080 → internal WS proxy (10015) → SSH:22
- **Advanced DNS Multiplexing:** `dnsdist` + kernel `U32` filters run **SlowDNS, VayDNS, Slipstream** concurrently on UDP 53
- **ZiVPN & UDP Custom:** UDP gaming protocols (ports 6000-19999)
- **Falcon Proxy & ProxyDT:** Optimized HTTP proxies

### 🛡️ Administration
- **Root SSH auto-configuration** on cloud VPS images
- **Reboot resilience:** rebuilds networking, iptables, u32 rules, IPv6, dnsdist on every boot
- **Broadcast:** send announcements to all registered users
- **Live monitoring:** VPS metrics (cores, RAM, disk, uptime) + active protocols
- **Bans and quotas:** per-role limits on accounts, duration, devices
- **Expiry alerts:** admins notified 1 day and 1 hour before account expiry

### 🧹 Maintenance
- **Durable state:** traffic counters and settings survive reboots
- **Service self-healing:** auto-recovery of HAProxy, Xray, DNS
- **HAProxy auto-recovery:** verifies running, kills port squatters, restarts when needed

---

## 💸 Monetization (Monetag + Vercel)

The bot ships with a native monetization flow. When enabled, public users must watch a Rewarded
Interstitial ad before **creating** or **renewing** an account. Setup is driven by an interactive
wizard inside the bot.

### Step-by-step

**1. Create a Monetag account**
- Register as a publisher at [Monetag.com](https://monetag.com)
- Add a new application → select **Telegram Mini App** format
- Provide your bot username (e.g. `@orxtunnel_bot`)
- Create a **Rewarded Interstitial** ad block

**2. Collect the two code snippets**
- **SDK script:** `<script>` tag
- **Rewarded block:** JavaScript function

**3. Run the in-bot wizard**
- Open **⚙️ Pro Settings → ⚙️ Configure MiniApp Ads**
- Paste the SDK script and Rewarded block
- Bot builds `monetag_miniapp.zip` — download, extract, use the `.html`

**4. Host the Mini App**
- Deploy to [Vercel](https://vercel.com/), [GitHub Pages](https://pages.github.com/), or any static host
- Copy the direct link to the HTML file

**5. Activate**
- Send that link to the bot as the final wizard step. Ad wall enabled automatically.

---

## ☁️ Native Backups (Telegram)

Backups delivered to admin chat — **on demand** from the menu, or **automatically**
every 1, 3, 7, or 30 days. No external service required. Restore by sending the `.json`
backup file back to the bot.

---

## 🛠️ Troubleshooting

| Symptom | Likely Cause | Fix |
| :--- | :--- | :--- |
| No connection on mobile data | NS domain has no IP, or IPv6 failing | Reinstall protocols from Protocols menu |
| Bot not responding | Process hung | `systemctl restart depwise` |
| Xray / VMess not connecting | HAProxy or Xray not started | `systemctl status haproxy xray` |
| Automatic backup failing | Invalid chat ID | Reconfigure backup interval in menu |
| `Exec format error` | Binary built for wrong CPU arch | Reinstall; installer selects correct `amd64` / `arm64` |

Useful commands:
```bash
systemctl status depwise         # service state
journalctl -u depwise -f         # live logs
```

---

## 💎 Credits and Support

- **📢 Channel:** [@orxtunnel](https://t.me/orxtunnel)
- **🤖 Bot:** [@orxtunnel_bot](https://t.me/orxtunnel_bot)

Built on the Go rewrite of the original Telegram VPN manager, rebranded and maintained as
**ORX TUNNEL**.

---

<p align="center">
  <i>Powering your VPS with the speed of Go and the reach of the kernel.</i><br>
  <b>© 2026 ORX TUNNEL Project</b>
</p>