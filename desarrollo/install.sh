#!/bin/bash
set -euo pipefail

# =========================================================
# INSTALADOR UNIVERSAL V8.0.5: BOT TELEGRAM ORX TUNNEL SSH 💎 (BINARY EDITION)
# =========================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

if [ "$EUID" -ne 0 ]; then
  log_error "Por favor, ejecuta este script como root"
  exit 1
fi

PROJECT_DIR="/opt/orxtunnel_bot"
ENV_FILE="$PROJECT_DIR/.env"

# --- CONFIGURACION PRIVADA ---
FIREBASE_URL_B64="aHR0cHM6Ly9rZXlnZW5icHQtZGVmYXVsdC1ydGRiLmZpcmViYXNlaW8uY29t"
FIREBASE_URL=$(echo "$FIREBASE_URL_B64" | base64 -d)

# ----------------------------------------------------------

install_bot() {
    echo -e "${GREEN}=================================================="
    echo -e "       CONFIGURACION BOT ORX TUNNEL V8.0 (BINARY)"
    echo -e "==================================================${NC}"

    # Validación de Key de Instalación
    if [ -z "${INSTALL_KEY:-}" ]; then
        read -p "Introduce tu Key de Instalación: " INSTALL_KEY
    fi
    if [ -z "$INSTALL_KEY" ]; then
        log_error "La Key no puede estar vacía."
        exit 1
    fi
    
    # Limpiar posibles caracteres ocultos (CRLF, espacios) de copiar y pegar
    INSTALL_KEY=$(echo "$INSTALL_KEY" | tr -d '\r' | tr -d '\n' | tr -d ' ')
    
    log_info "Instalando y actualizando dependencias de red..."
    apt update -y && apt install -y curl wget ca-certificates || { log_error "Error al instalar dependencias base."; exit 1; }
    update-ca-certificates || true

    log_info "Verificando Key en la base de datos..."
    if ! KEY_RESPONSE=$(curl -k -4 -s -m 10 "${FIREBASE_URL}/keys/${INSTALL_KEY}.json" || wget --no-check-certificate -qO- --timeout=10 "${FIREBASE_URL}/keys/${INSTALL_KEY}.json"); then
        log_error "Error de conexión con Firebase. Revisa tu internet o DNS."
        exit 1
    fi
    if [ "$KEY_RESPONSE" == "null" ] || [ -z "$KEY_RESPONSE" ]; then
        log_error "Key inválida o ya ha sido usada."
        exit 1
    fi
    
    log_info "Key válida. Quemando Key..."
    curl -4 -s -X DELETE "${FIREBASE_URL}/keys/${INSTALL_KEY}.json" > /dev/null || true
    
    if [ -f "$ENV_FILE" ]; then
        log_info "Cargando credenciales existentes desde $ENV_FILE..."
        BOT_TOKEN=$(grep -E "^BOT_TOKEN=" "$ENV_FILE" | cut -d'=' -f2-)
        ADMIN_ID=$(grep -E "^SUPER_ADMIN=" "$ENV_FILE" | cut -d'=' -f2-)
    fi

    if [ -z "${BOT_TOKEN:-}" ]; then
        read -p "Introduce el TOKEN: " BOT_TOKEN
    fi
    
    if [ -z "${ADMIN_ID:-}" ]; then
        if [ -n "${SUPER_ADMIN:-}" ]; then
            ADMIN_ID="${SUPER_ADMIN}"
        else
            read -p "Introduce tu Chat ID de Telegram: " ADMIN_ID
        fi
    fi

    if [ -z "$BOT_TOKEN" ] || [ -z "$ADMIN_ID" ]; then
        log_error "Error: Datos incompletos."
        exit 1
    fi

    # 1. Preparar Entorno
    mkdir -p "$PROJECT_DIR"
    echo "BOT_TOKEN=$BOT_TOKEN" > "$ENV_FILE"
    echo "SUPER_ADMIN=$ADMIN_ID" >> "$ENV_FILE"
    chmod 600 "$ENV_FILE"

    log_info "Instalando dependencias base..."
    apt update -y && apt install -y curl wget make git || true

    # 2. Descargar Binario del Bot
    log_info "Descargando el Bot en Go (Binario Precompilado Oficial)..."
    systemctl stop depwise 2>/dev/null || true
    
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        BIN_URL="https://github.com/Depwisescript/Depwise-Installers/releases/latest/download/orxtunnel-bot-amd64"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        BIN_URL="https://github.com/Depwisescript/Depwise-Installers/releases/latest/download/orxtunnel-bot-arm64"
    else
        log_error "Arquitectura no soportada: $ARCH"
        exit 1
    fi

    wget -qO /usr/local/bin/orxtunnel-bot "${BIN_URL}?t=$(date +%s)" || { log_error "Error al descargar el bot."; exit 1; }
    chmod +x /usr/local/bin/orxtunnel-bot

    # 3. Compilar BadVPN nativamente (Repositorio Público, sin tokens)
    if [ ! -f "/usr/bin/badvpn-udpgw" ]; then
        log_info "Compilando motor BadVPN (puede tomar 1 minuto)..."
        apt install -y cmake build-essential
        cd /tmp
        rm -rf badvpn
        git clone https://github.com/ambrop72/badvpn.git
        cd badvpn
        cmake -DBUILD_NOTHING_BY_DEFAULT=1 -DBUILD_UDPGW=1 .
        make
        cp udpgw/badvpn-udpgw /usr/bin/badvpn-udpgw
        chmod +x /usr/bin/badvpn-udpgw
        cd ~
        rm -rf /tmp/badvpn
    fi

    # 4. Servicio Systemd
    log_info "Generando sistema daemon SystemD..."
    cat << EOF > /etc/systemd/system/orxtunnel.service
[Unit]
Description=Depwise Telegram Bot (Go Edition)
After=network.target

[Service]
Type=simple
User=root
EnvironmentFile=$ENV_FILE
Environment="GOMEMLIMIT=40MiB" "GOGC=20"
ExecStart=/usr/local/bin/orxtunnel-bot
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable orxtunnel.service
    systemctl restart orxtunnel.service

    echo -e "${GREEN}=================================================="
    echo -e "       INSTALACION V8.0 COMPLETADA 💎"
    echo -e "=================================================="
    echo -e "El bot de Go está escuchando. Puedes enviar /start en Telegram.${NC}"
}

uninstall_all() {
    echo -e "${RED}=================================================="
    echo -e "       ⚠️ ADVERTENCIA: DESINSTALACIÓN TOTAL ⚠️"
    echo -e "==================================================${NC}"
    echo -e "Esto eliminará:"
    echo -e "- El Bot de Telegram y sus configuraciones"
    echo -e "- Todos los servicios VPN instalados por el bot (SlowDNS, ProxyDT, SSL, etc.)"
    echo -e "- Los binarios descargados"
    echo -e "- La base de datos de usuarios (bot_data.json)"
    
    read -p "¿Estás completamente seguro de continuar? (escribe 'si' para confirmar): " confirm
    if [ "$confirm" != "si" ]; then
        log_info "Desinstalación cancelada."
        return
    fi

    log_info "1/4 Deteniendo servicios..."
    systemctl stop orxtunnel.service 2>/dev/null || true
    systemctl disable orxtunnel.service 2>/dev/null || true
    
    # Detener proxies y vpns
    local services=("badvpn" "proxydt" "stunnel4" "dropbear" "falconproxy" "udpcustom" "zivpn" "nsd")
    for svc in "${services[@]}"; do
        systemctl stop "$svc" 2>/dev/null || true
        systemctl disable "$svc" 2>/dev/null || true
        rm -f "/etc/systemd/system/${svc}.service"
    done

    log_info "2/4 Eliminando archivos y binarios..."
    rm -f /usr/local/bin/orxtunnel-bot
    rm -f /etc/systemd/system/orxtunnel.service
    rm -rf "$PROJECT_DIR"
    rm -f /root/bot_data.json
    
    # VPN Binaries & Configs
    rm -f /usr/local/bin/badvpn-udpgw
    rm -f /usr/bin/badvpn-udpgw
    rm -f /usr/bin/badvpn
    rm -f /usr/local/bin/proxydt
    rm -f /usr/local/bin/falconproxy
    rm -f /usr/local/bin/udpcustom
    rm -rf /etc/zivpn
    rm -f /usr/local/bin/zivpn
    rm -f /etc/falconproxy.conf
    rm -rf /etc/slowdns

    log_info "3/4 Recargando demonios de sistema..."
    systemctl daemon-reload

    echo -e "${GREEN}=================================================="
    echo -e "   ✅ DESINSTALACIÓN COMPLETADA EXITOSAMENTE  "
    echo -e "==================================================${NC}"
}

enable_root() {
    echo -e "${CYAN}=================================================="
    echo -e "       HABILITANDO ACCESO ROOT POR SSH"
    echo -e "==================================================${NC}"
    read -p "Introduce una nueva contraseña para el usuario root: " ROOT_PASS
    if [ -z "$ROOT_PASS" ]; then
        log_error "La contraseña no puede estar vacía."
        sleep 2
        return
    fi
    echo "root:$ROOT_PASS" | chpasswd
    
    # Habilitar PermitRootLogin
    sed -i 's/^#*PermitRootLogin.*/PermitRootLogin yes/g' /etc/ssh/sshd_config
    if ! grep -q "^PermitRootLogin yes" /etc/ssh/sshd_config; then
        echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
    fi
    
    # Habilitar PasswordAuthentication
    sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication yes/g' /etc/ssh/sshd_config
    if ! grep -q "^PasswordAuthentication yes" /etc/ssh/sshd_config; then
        echo "PasswordAuthentication yes" >> /etc/ssh/sshd_config
    fi
    
    # Solucionar archivos de sobreescritura de AWS (Ubuntu 22.04+)
    if [ -d "/etc/ssh/sshd_config.d" ]; then
        sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication yes/g' /etc/ssh/sshd_config.d/*.conf 2>/dev/null || true
        sed -i 's/^#*PermitRootLogin.*/PermitRootLogin yes/g' /etc/ssh/sshd_config.d/*.conf 2>/dev/null || true
    fi
    
    # Reiniciar servicio SSH (soportado en Ubuntu 22.04 y 24.04)
    log_info "Reiniciando servicio SSH..."
    systemctl restart ssh 2>/dev/null || systemctl restart sshd 2>/dev/null
    
    log_info "Acceso root habilitado correctamente con contraseña."
    sleep 2
}

show_menu() {
    clear
    echo -e "${CYAN}=================================================="
    echo -e "       ORX TUNNEL BOT INSTALLER (GO EDITION)"
    echo -e "==================================================${NC}"
    echo -e "  1. ${GREEN}Instalar / Actualizar Bot${NC}"
    echo -e "  2. ${RED}Desinstalar Todo (Bot + VPNs)${NC}"
    echo -e "  3. ${YELLOW}Habilitar Acceso Root SSH (AWS/VPS)${NC}"
    echo -e "  4. Salir"
    echo -e "${CYAN}==================================================${NC}"
    read -p "Selecciona una opción [1-4]: " opt

    case $opt in
        1) install_bot ;;
        2) uninstall_all ;;
        3) enable_root ; show_menu ;;
        4) exit 0 ;;
        *) log_error "Opción inválida"; sleep 2; show_menu ;;
    esac
}

if [ -n "${INSTALL_KEY:-}" ] && [ -n "${BOT_TOKEN:-}" ]; then
    install_bot
else
    show_menu
fi
