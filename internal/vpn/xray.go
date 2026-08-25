package vpn

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const xrayConfigPath = "/usr/local/etc/xray/config.json"
const xrayAccessLog = "/var/log/xray/access.log"

// Xray protocol definitions. Each protocol listens on its own local port and is
// exposed to clients through HAProxy (TLS termination) on port 443/80 using a
// distinct WebSocket path.
const (
	ProtoVMess = "vmess"
	ProtoVLESS = "vless"
	ProtoTrojan = "trojan"

	vmessPort = 10002
	vlessPort = 10012
	trojanPort = 10013

	vmessPath = "/vmess"
	vlessPath = "/vless"
	trojanPath = "/trojan-ws"
)

// protocolPort devuelve el puerto local y la ruta WS para un protocolo.
func protocolPort(proto string) (int, string) {
	switch proto {
	case ProtoVLESS:
		return vlessPort, vlessPath
	case ProtoTrojan:
		return trojanPort, trojanPath
	default: // vmess
		return vmessPort, vmessPath
	}
}

func protocolLabel(proto string) string {
	switch proto {
	case ProtoVLESS:
		return "VLESS"
	case ProtoTrojan:
		return "Trojan"
	default:
		return "VMess"
	}
}

// InstallXray instala el núcleo de Xray y configura el archivo JSON inicial
// con los tres protocolos (VMess, VLESS y Trojan) sobre WebSocket.
func InstallXray() error {
	// 1. Descargar e instalar Xray desde el script oficial de GitHub
	cmd := exec.Command("bash", "-c", "bash -c \"$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)\" @ install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xray core installation failed: %v", err)
	}

	// 2. Crear configuración base con los tres protocolos
	os.MkdirAll(filepath.Dir(xrayAccessLog), 0755)

	baseConfig := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
			"access":   xrayAccessLog,
		},
		"inbounds": []map[string]interface{}{
			makeInbound(ProtoVMess),
			makeInbound(ProtoVLESS),
			makeInbound(ProtoTrojan),
		},
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"tag":      "direct",
			},
			{
				"protocol": "blackhole",
				"tag":      "block",
			},
		},
	}

	raw, err := json.MarshalIndent(baseConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error generando JSON base: %v", err)
	}

	if err := os.WriteFile(xrayConfigPath, raw, 0644); err != nil {
		return fmt.Errorf("error escribiendo config.json de xray: %v", err)
	}

	// 3. Asegurar que los tres inbounds existan (idempotente)
	if err := EnsureXrayProtocols(); err != nil {
		return fmt.Errorf("error asegurando protocolos xray: %v", err)
	}

	// 4. Aplicar resiliencia del servicio (auto-restart y fix de OOM/Reboot)
	if err := EnsureXrayServiceResilience(); err != nil {
		return fmt.Errorf("error aplicando resiliencia a xray: %v", err)
	}

	return nil
}

// makeInbound construye la definición de un inbound Xray para el protocolo dado.
func makeInbound(proto string) map[string]interface{} {
	_, wsPath := protocolPort(proto)
	inbound := map[string]interface{}{
		"listen":   "127.0.0.1",
		"protocol": proto,
		"settings": map[string]interface{}{
			"clients": []map[string]interface{}{},
		},
		"streamSettings": map[string]interface{}{
			"network": "ws",
			"wsSettings": map[string]interface{}{
				"path": wsPath,
			},
		},
		"sniffing": map[string]interface{}{
			"enabled":      true,
			"destOverride": []string{"http", "tls"},
		},
	}
	switch proto {
	case ProtoVMess:
		inbound["port"] = vmessPort
	case ProtoVLESS:
		inbound["port"] = vlessPort
		inbound["settings"] = map[string]interface{}{
			"clients":    []map[string]interface{}{},
			"decryption": "none",
		}
	case ProtoTrojan:
		inbound["port"] = trojanPort
		inbound["settings"] = map[string]interface{}{
			"clients": []map[string]interface{}{},
		}
	}
	return inbound
}

// RemoveXray detiene y borra el núcleo
func RemoveXray() error {
	exec.Command("systemctl", "stop", "xray").Run()
	exec.Command("systemctl", "disable", "xray").Run()
	exec.Command("bash", "-c", "bash -c \"$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)\" @ remove").Run()
	os.RemoveAll("/usr/local/etc/xray")
	return nil
}

// EnsureXrayAccessLog verifica que el access log esté habilitado en la config
// existente. Si no lo está, lo agrega y reinicia Xray.
func EnsureXrayAccessLog() error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	logSection, ok := cfg["log"].(map[string]interface{})
	if !ok {
		logSection = make(map[string]interface{})
		cfg["log"] = logSection
	}

	if existing, ok := logSection["access"].(string); ok && existing != "" {
		return nil
	}

	os.MkdirAll(filepath.Dir(xrayAccessLog), 0755)

	logSection["access"] = xrayAccessLog
	cfg["log"] = logSection

	return saveXrayConfig(cfg)
}

// EnsureXrayProtocols asegura que los tres inbounds (VMess, VLESS, Trojan)
// existan en la configuración actual de Xray. Si falta alguno, lo agrega.
// Es idempotente y seguro de llamar en cada arranque del bot.
func EnsureXrayProtocols() error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	inboundsRaw, ok := cfg["inbounds"].([]interface{})
	if !ok {
		inboundsRaw = []interface{}{}
	}

	present := map[string]bool{}
	for _, inb := range inboundsRaw {
		if m, ok := inb.(map[string]interface{}); ok {
			if p, ok := m["protocol"].(string); ok {
				present[p] = true
			}
		}
	}

	changed := false
	for _, proto := range []string{ProtoVMess, ProtoVLESS, ProtoTrojan} {
		if !present[proto] {
			inboundsRaw = append(inboundsRaw, makeInbound(proto))
			changed = true
		}
	}

	if changed {
		cfg["inbounds"] = inboundsRaw
		return saveXrayConfig(cfg)
	}
	return nil
}

// EnsureXrayServiceResilience asegura que el demonio de Xray se reinicie automáticamente
// en caso de fallo (ej. OOM kill o saturación) y que espere a la red al reiniciar el VPS.
func EnsureXrayServiceResilience() error {
	dir := "/etc/systemd/system/xray.service.d"
	overridePath := filepath.Join(dir, "10-resilience.conf")

	if _, err := os.Stat(overridePath); err == nil {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := `[Unit]
After=network-online.target
Wants=network-online.target

[Service]
Restart=always
RestartSec=3
StartLimitIntervalSec=0
`
	if err := os.WriteFile(overridePath, []byte(content), 0644); err != nil {
		return err
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "xray").Run()
	exec.Command("systemctl", "restart", "xray").Run()

	return nil
}

// loadXrayConfig lee la config JSON existente
func loadXrayConfig() (map[string]interface{}, error) {
	raw, err := os.ReadFile(xrayConfigPath)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(raw, &data)
	return data, err
}

// saveXrayConfig escribe la config JSON al sistema y reinicia el demonio
func saveXrayConfig(data map[string]interface{}) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(xrayConfigPath, raw, 0644); err != nil {
		return err
	}
	// Reinicio silencioso. Si se alcanza el límite de arranques de systemd
	// (p.ej. cambios muy rápidos), reseteamos y forzamos el arranque.
	if err := exec.Command("systemctl", "restart", "xray").Run(); err != nil {
		exec.Command("systemctl", "reset-failed", "xray").Run()
		return exec.Command("systemctl", "start", "xray").Run()
	}
	return nil
}

// AddXrayUser inyecta un nuevo cliente al inbound correspondiente al protocolo.
// Para VMess/VLESS el "credential" es el UUID; para Trojan es la contraseña.
func AddXrayUser(protocol, credential, email string) error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	inboundsRaw, ok := cfg["inbounds"].([]interface{})
	if !ok || len(inboundsRaw) == 0 {
		return fmt.Errorf("invalid inbounds format in config.json")
	}

	targetProto := protocol
	if targetProto == "" {
		targetProto = ProtoVMess
	}

	for _, inb := range inboundsRaw {
		inbound, ok := inb.(map[string]interface{})
		if !ok {
			continue
		}
		if inbound["protocol"] != targetProto {
			continue
		}

		settings, ok := inbound["settings"].(map[string]interface{})
		if !ok {
			settings = map[string]interface{}{}
			inbound["settings"] = settings
		}

		var clients []interface{}
		if settings["clients"] != nil {
			clients, _ = settings["clients"].([]interface{})
		}

		newClient := map[string]interface{}{
			"level": 0,
			"email": email,
		}
		if targetProto == ProtoTrojan {
			newClient["password"] = credential
		} else {
			newClient["id"] = credential
		}
		clients = append(clients, newClient)
		settings["clients"] = clients

		return saveXrayConfig(cfg)
	}

	return fmt.Errorf("inbound for protocol %s not found", targetProto)
}

// RemoveXrayUser busca el credential (UUID o password) en todos los inbounds y
// lo elimina.
func RemoveXrayUser(credential string) error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	inboundsRaw, ok := cfg["inbounds"].([]interface{})
	if !ok || len(inboundsRaw) == 0 {
		return fmt.Errorf("invalid inbounds format in config.json")
	}

	changed := false
	for _, inb := range inboundsRaw {
		inbound, ok := inb.(map[string]interface{})
		if !ok {
			continue
		}
		settings, ok := inbound["settings"].(map[string]interface{})
		if !ok || settings["clients"] == nil {
			continue
		}
		clients, ok := settings["clients"].([]interface{})
		if !ok {
			continue
		}

		var newClients []interface{}
		for _, c := range clients {
			clientMap, ok := c.(map[string]interface{})
			if !ok {
				newClients = append(newClients, c)
				continue
			}
			id, _ := clientMap["id"].(string)
			pw, _ := clientMap["password"].(string)
			if id == credential || pw == credential {
				changed = true
				continue
			}
			newClients = append(newClients, c)
		}
		settings["clients"] = newClients
	}

	if changed {
		return saveXrayConfig(cfg)
	}
	return nil
}

// GenerateVmessLink crea el texto base64 para importar el perfil en v2rayNG / HTTP Custom
func GenerateVmessLink(alias, uuid, domain string) string {
	vmessObj := map[string]interface{}{
		"v":    "2",
		"ps":   alias,
		"add":  domain,
		"port": "443",
		"id":   uuid,
		"aid":  "0",
		"scy":  "auto",
		"net":  "ws",
		"type": "none",
		"host": domain,
		"path": vmessPath,
		"tls":  "tls",
		"sni":  domain,
		"alpn": "",
	}

	raw, _ := json.Marshal(vmessObj)
	encoded := base64.StdEncoding.EncodeToString(raw)
	return "vmess://" + encoded
}

// GenerateVlessLink crea el enlace vless:// para clientes V2RayNG / v2rayN.
func GenerateVlessLink(alias, uuid, domain string) string {
	q := url.Values{}
	q.Set("type", "ws")
	q.Set("security", "tls")
	q.Set("path", vlessPath)
	q.Set("sni", domain)
	return fmt.Sprintf("vless://%s@%s:443?%s#%s", uuid, domain, q.Encode(), url.QueryEscape(alias))
}

// GenerateTrojanLink crea el enlace trojan:// para clientes compatibles.
func GenerateTrojanLink(alias, password, domain string) string {
	q := url.Values{}
	q.Set("type", "ws")
	q.Set("security", "tls")
	q.Set("path", trojanPath)
	q.Set("sni", domain)
	return fmt.Sprintf("trojan://%s@%s:443?%s#%s", password, domain, q.Encode(), url.QueryEscape(alias))
}

// GenCredential genera el identificador adecuado para un protocolo:
// UUID para VMess/VLESS, contraseña aleatoria para Trojan.
func GenCredential(protocol string) string {
	if protocol == ProtoTrojan {
		return randHex(24)
	}
	return uuid.New().String()
}

func randHex(n int) string {
	b := make([]byte, n/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetXrayOnlineUsers retorna los emails de usuarios activos en los últimos 60 segundos
// leyendo el access log de Xray.
func GetXrayOnlineUsers() []string {
	file, err := os.Open(xrayAccessLog)
	if err != nil {
		return nil
	}
	defer file.Close()

	cutoff := time.Now().Add(-60 * time.Second)
	activeEmails := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		emailIdx := strings.Index(line, "email: ")
		if emailIdx == -1 {
			continue
		}

		if len(line) < 19 {
			continue
		}
		tsStr := line[:19]
		ts, err := time.ParseInLocation("2006/01/02 15:04:05", tsStr, time.Local)
		if err != nil {
			continue
		}

		if ts.Before(cutoff) {
			continue
		}

		email := strings.TrimSpace(line[emailIdx+7:])
		if email != "" {
			activeEmails[email] = true
		}
	}

	var result []string
	for email := range activeEmails {
		result = append(result, email)
	}
	return result
}
