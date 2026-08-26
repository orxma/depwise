#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "=========================================="
echo "  Depwise / ORX TUNNEL Bot Installer"
echo "=========================================="
echo

read -p "Enter Bot Token: " BOT_TOKEN
if [[ -z "$BOT_TOKEN" ]]; then
    log_error "Bot token is required"
    exit 1
fi

read -p "Enter Super Admin ID (Telegram numeric ID): " SUPER_ADMIN
if [[ -z "$SUPER_ADMIN" ]] || ! [[ "$SUPER_ADMIN" =~ ^[0-9]+$ ]]; then
    log_error "Valid numeric Super Admin ID is required"
    exit 1
fi

log_info "Updating system..."
apt-get update -qq
apt-get install -y -qq git curl golang-go 2>/dev/null || {
    log_info "Installing Go manually..."
    GO_VERSION="1.22.5"
    ARCH=$(dpkg --print-architecture)
    if [[ "$ARCH" == "amd64" ]]; then
        GO_ARCH="amd64"
    elif [[ "$ARCH" == "arm64" ]]; then
        GO_ARCH="arm64"
    else
        log_error "Unsupported architecture: $ARCH"
        exit 1
    fi
    curl -L -s "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" | tar -xz -C /usr/local
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
}

export PATH=$PATH:/usr/local/go/bin

log_info "Cloning repository..."
cd /root
if [[ -d "depwise" ]]; then
    log_warn "Directory exists, pulling latest..."
    cd depwise && git pull
else
    git clone https://github.com/orxma/depwise.git
    cd depwise
fi

log_info "Building bot..."
go build -o /usr/local/bin/depwise-bot ./cmd/orxtunnel

log_info "Creating config directory..."
mkdir -p /opt/depwise_bot
cat > /opt/depwise_bot/.env <<EOF
BOT_TOKEN=$BOT_TOKEN
SUPER_ADMIN=$SUPER_ADMIN
EOF
chmod 600 /opt/depwise_bot/.env

log_info "Creating systemd service..."
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

log_info "Starting bot..."
systemctl daemon-reload
systemctl enable depwise
systemctl restart depwise

sleep 3

if systemctl is-active --quiet depwise; then
    log_info "Bot installed and running successfully!"
    echo
    echo "Service: systemctl status depwise"
    echo "Logs:    journalctl -u depwise -f"
    echo "Config:  /opt/depwise_bot/.env"
    echo "Binary:  /usr/local/bin/depwise-bot"
else
    log_error "Bot failed to start. Check logs: journalctl -u depwise -n 50"
    exit 1
fi