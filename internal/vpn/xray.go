package vpn

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const xrayConfigPath = "/usr/local/etc/xray/config.json"
const xrayAccessLog = "/var/log/xray/access.log"

// InstallXray instala el núcleo de Xray y configura el archivo JSON inicial
func InstallXray() error {
	// 1. Descargar e instalar Xray desde el script oficial de GitHub
	cmd := exec.Command("bash", "-c", "bash -c \"$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)\" @ install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xray core installation failed: %v", err)
	}

	// 2. Crear configuración base de VMess WS
	// Asegurar que el directorio de logs exista
	os.MkdirAll(filepath.Dir(xrayAccessLog), 0755)

	baseConfig := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
			"access":   xrayAccessLog,
		},
		"inbounds": []map[string]interface{}{
			{
				"port":     10002, // Puerto local fijo para enlazar con HAProxy
				"listen":   "127.0.0.1",
				"protocol": "vmess",
				"settings": map[string]interface{}{
					"clients": []map[string]interface{}{},
				},
				"streamSettings": map[string]interface{}{
					"network": "ws",
					"wsSettings": map[string]interface{}{
						"path": "/vmess",
					},
				},
				"sniffing": map[string]interface{}{
					"enabled": true,
					"destOverride": []string{"http", "tls"},
				},
			},
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

	// Aplicar resiliencia del servicio (auto-restart y fix de OOM/Reboot)
	if err := EnsureXrayServiceResilience(); err != nil {
		return fmt.Errorf("error aplicando resiliencia a xray: %v", err)
	}

	return nil
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
// existente. Si no lo está, lo agrega y reinicia Xray. Útil para instalaciones
// anteriores a esta funcionalidad.
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

	// Si ya tiene el access log configurado, no hacer nada
	if existing, ok := logSection["access"].(string); ok && existing != "" {
		return nil
	}

	// Asegurar directorio de logs
	os.MkdirAll(filepath.Dir(xrayAccessLog), 0755)

	logSection["access"] = xrayAccessLog
	cfg["log"] = logSection

	return saveXrayConfig(cfg)
}

// EnsureXrayServiceResilience asegura que el demonio de Xray se reinicie automáticamente
// en caso de fallo (ej. OOM kill o saturación) y que espere a la red al reiniciar el VPS.
func EnsureXrayServiceResilience() error {
	dir := "/etc/systemd/system/xray.service.d"
	overridePath := filepath.Join(dir, "10-resilience.conf")

	// Si ya existe, asumimos que está configurado
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
	// Reinicio silencioso
	return exec.Command("systemctl", "restart", "xray").Run()
}

// AddXrayUser inyecta el nuevo usuario VMess al archivo y reinicia el core
func AddXrayUser(uuid, email string) error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	inbounds, ok := cfg["inbounds"].([]interface{})
	if !ok || len(inbounds) == 0 {
		return fmt.Errorf("invalid inbounds format in config.json")
	}

	inbound0 := inbounds[0].(map[string]interface{})
	settings := inbound0["settings"].(map[string]interface{})
	
	var clients []interface{}
	if settings["clients"] != nil {
		clients = settings["clients"].([]interface{})
	}

	newUser := map[string]interface{}{
		"id":    uuid,
		"level": 0,
		"email": email, // Guardamos el alias o chatid para identificarlo
	}
	clients = append(clients, newUser)
	settings["clients"] = clients

	return saveXrayConfig(cfg)
}

// RemoveXrayUser busca el UUID y lo elimina de la lista de clientes.
func RemoveXrayUser(uuid string) error {
	cfg, err := loadXrayConfig()
	if err != nil {
		return err
	}

	inbounds, ok := cfg["inbounds"].([]interface{})
	if !ok || len(inbounds) == 0 {
		return fmt.Errorf("invalid inbounds format in config.json")
	}

	inbound0 := inbounds[0].(map[string]interface{})
	settings := inbound0["settings"].(map[string]interface{})
	
	if settings["clients"] == nil {
		return nil // no hay clientes
	}
	clients := settings["clients"].([]interface{})

	var newClients []interface{}
	for _, c := range clients {
		clientMap := c.(map[string]interface{})
		if clientMap["id"] != uuid {
			newClients = append(newClients, c)
		}
	}
	settings["clients"] = newClients

	return saveXrayConfig(cfg)
}

// GenerateVmessLink crea el texto base64 para importar el perfil en v2rayNG / HTTP Custom
func GenerateVmessLink(alias, uuid, domain string) string {
	vmessObj := map[string]interface{}{
		"v":    "2",
		"ps":   alias,
		"add":  domain,
		"port": "443", // Puerto SSL Tunnel (HAProxy)
		"id":   uuid,
		"aid":  "0",
		"scy":  "auto",
		"net":  "ws",
		"type": "none",
		"host": domain,
		"path": "/vmess",
		"tls":  "tls",
		"sni":  domain,
		"alpn": "",
	}
	
	raw, _ := json.Marshal(vmessObj)
	encoded := base64.StdEncoding.EncodeToString(raw)
	return "vmess://" + encoded
}

// GetXrayOnlineUsers retorna los emails de usuarios VMess activos en los últimos 60 segundos
// leyendo el access log de Xray.
// Formato de log: 2026/04/06 18:30:00 1.2.3.4:12345 accepted tcp:8.8.8.8:443 [vmess >> direct] email: user@alias
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

		// Buscar "email:" en la línea
		emailIdx := strings.Index(line, "email: ")
		if emailIdx == -1 {
			continue
		}

		// Parsear timestamp al inicio de la línea (formato: 2026/04/06 18:30:00)
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
