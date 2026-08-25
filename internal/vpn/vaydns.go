package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// InstallVayDNS downloads the real vaydns-server binary, generating keys and services for VayDNS.
func InstallVayDNS(domain, port string) (string, error) {
	arch := runtime.GOARCH
	binName := "vaydns-server-linux-" + arch
	if arch == "arm" {
		binName = "vaydns-server-linux-armv7"
	}

	mirrors := []string{
		"https://github.com/net2share/vaydns/releases/download/v0.2.8/" + binName,
	}

	// Attempt Download
	success := false

	// Verify if the local file exists and is indeed VayDNS (not a copy of slowdns-server)
	if _, err := os.Stat("/usr/bin/vaydns-server"); err == nil {
		out, _ := exec.Command("/usr/bin/vaydns-server", "-h").CombinedOutput()
		if strings.Contains(string(out), "-domain") {
			success = true
		}
	}

	if !success {
		for _, url := range mirrors {
			cmd := exec.Command("curl", "-L", "-k", "-s", "-f", "-o", "/usr/bin/vaydns-server", url)
			if err := cmd.Run(); err == nil {
				success = true
				break
			}
		}
	}

	if !success {
		return "", fmt.Errorf("fallo al descargar binario para %s", arch)
	}
	os.Chmod("/usr/bin/vaydns-server", 0755)

	// Key Generation
	os.MkdirAll("/etc/vaydns", 0755)
	if _, err := os.Stat("/etc/vaydns/server.pub"); os.IsNotExist(err) {
		exec.Command("/usr/bin/vaydns-server", "-gen-key", "-privkey-file", "/etc/vaydns/server.key", "-pubkey-file", "/etc/vaydns/server.pub").Run()
	}

	// Se eliminó iptables de aquí, ahora lo maneja dnsdist

	// Service Creation (VayDNS uses port 5301)
	svc := `[Unit]
Description=VayDNS ORX TUNNEL Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/vaydns
ExecStart=/usr/bin/vaydns-server -udp :5301 -privkey-file /etc/vaydns/server.key -domain ` + domain + ` -upstream 127.0.0.1:` + port + `
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target`

	os.WriteFile("/etc/systemd/system/vaydns.service", []byte(svc), 0644)
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "vaydns").Run()
	exec.Command("systemctl", "restart", "vaydns").Run()

	pubBytes, _ := os.ReadFile("/etc/vaydns/server.pub")
	return strings.TrimSpace(string(pubBytes)), nil
}

// RemoveVayDNS stops and removes the vaydns daemon.
func RemoveVayDNS() error {
	exec.Command("systemctl", "stop", "vaydns").Run()
	exec.Command("systemctl", "disable", "vaydns").Run()
	os.Remove("/etc/systemd/system/vaydns.service")
	exec.Command("systemctl", "daemon-reload").Run()

	// Se eliminó iptables de aquí, ahora lo maneja dnsdist

	os.RemoveAll("/etc/vaydns")
	return nil
}
