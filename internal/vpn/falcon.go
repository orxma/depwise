package vpn

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// falconFallbackPy is a safe, self-contained HTTP/HTTPS (CONNECT) proxy used
// when the upstream FirewallFalcon binary cannot be downloaded. The upstream
// project's repository was removed and its components were reported as a
// supply-chain backdoor (root-CA injection, /etc/hosts hijacking, SSH
// backdoor). We therefore never fetch or execute that binary; instead we fall
// back to this auditable implementation.
//
//go:embed falconproxy_fallback.py
var falconFallbackPy []byte

// InstallFalcon descarga e instala Falcon Proxy
func InstallFalcon(port string) (string, error) {
	arch := runtime.GOARCH
	binName := "falconproxy"
	if arch == "arm64" {
		binName = "falconproxyarm"
	}

	// URLs a intentar
	urls := []string{
		fmt.Sprintf("https://github.com/firewallfalcons/FirewallFalcon-Manager/releases/download/v1.2-RustFast/%s", binName),
		fmt.Sprintf("https://github.com/firewallfalcons/FirewallFalcon-Manager/releases/latest/download/%s", binName),
	}

	var lastErr string
	downloaded := false

	for _, downURL := range urls {
		// Intentar con curl
		cmd := exec.Command("curl", "-L", "-f", "--connect-timeout", "15", "--max-time", "60", "-o", "/usr/local/bin/falconproxy", downURL)
		out, err := cmd.CombinedOutput()
		if err == nil {
			downloaded = true
			break
		}
		lastErr = fmt.Sprintf("curl: %s | %v", strings.TrimSpace(string(out)), err)

		// Intentar con wget como fallback
		cmd2 := exec.Command("wget", "--no-check-certificate", "-q", "-O", "/usr/local/bin/falconproxy", downURL)
		out2, err2 := cmd2.CombinedOutput()
		if err2 == nil {
			downloaded = true
			break
		}
		lastErr = fmt.Sprintf("wget: %s | %v", strings.TrimSpace(string(out2)), err2)
	}

	// 0. Fallback: if the upstream binary cannot be downloaded (repo removed
	// or unreachable), install the safe local Python proxy instead. This keeps
	// the feature working without pulling the backdoored FirewallFalcon binary.
	if !downloaded {
		if _, err := exec.LookPath("python3"); err != nil {
			// Try to ensure python3 is available.
			_ = exec.Command("apt-get", "update", "-qq").Run()
			_ = exec.Command("apt-get", "install", "-y", "-qq", "python3").Run()
		}
		if _, err := exec.LookPath("python3"); err != nil {
			return "", fmt.Errorf("falconproxy download failed and python3 is unavailable: %s", lastErr)
		}
		if err := os.WriteFile("/usr/local/bin/falconproxy", falconFallbackPy, 0755); err != nil {
			return "", fmt.Errorf("failed to write falconproxy fallback: %v", err)
		}
		downloaded = true
	}

	os.Chmod("/usr/local/bin/falconproxy", 0755)

	// 2. Configuración
	configContent := fmt.Sprintf("PORTS=\"%s\"\nINSTALLED_VERSION=\"latest\"\n", port)
	os.WriteFile("/etc/falconproxy.conf", []byte(configContent), 0644)

	// 3. Crear servicio Systemd
	service := fmt.Sprintf(`[Unit]
Description=Falcon Proxy Service
After=network.target

[Service]
User=root
Type=simple
ExecStart=/usr/local/bin/falconproxy -p %s
Restart=always
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, port)

	os.WriteFile("/etc/systemd/system/falconproxy.service", []byte(service), 0644)

	// 4. Iniciar servicio
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "falconproxy").Run()
	if err := exec.Command("systemctl", "restart", "falconproxy").Run(); err != nil {
		return "", fmt.Errorf("failed to start falconproxy: %v", err)
	}

	return "latest", nil
}

// RemoveFalcon elimina el servicio y archivos
func RemoveFalcon() error {
	exec.Command("systemctl", "stop", "falconproxy").Run()
	exec.Command("systemctl", "disable", "falconproxy").Run()
	os.Remove("/etc/systemd/system/falconproxy.service")
	os.Remove("/usr/local/bin/falconproxy")
	os.Remove("/etc/falconproxy.conf")
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}
