package vpn

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func installLibSSL11() {
	// Check if already exists
	if _, err := os.Stat("/usr/lib/x86_64-linux-gnu/libssl.so.1.1"); err == nil {
		return
	}
	if _, err := os.Stat("/usr/lib/aarch64-linux-gnu/libssl.so.1.1"); err == nil {
		return
	}

	arch := runtime.GOARCH
	var url string
	if arch == "amd64" {
		url = "http://nz2.archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb"
	} else if arch == "arm64" || arch == "aarch64" {
		url = "http://ports.ubuntu.com/ubuntu-ports/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_arm64.deb"
	}

	if url != "" {
		_ = exec.Command("wget", "-q", "-O", "/tmp/libssl1.1.deb", url).Run()
		_ = exec.Command("dpkg", "-i", "/tmp/libssl1.1.deb").Run()
		_ = os.Remove("/tmp/libssl1.1.deb")
	}
}

// GetSystemReport returns a diagnostic string about network and services
func GetSystemReport() string {
	report := "🛡️ <b>NETWORK TECHNICAL REPORT</b>\n\n"

	// 1. IPTables NAT (Prerouting)
	iptNat, _ := exec.Command("bash", "-c", "iptables -t nat -L PREROUTING -n -v | head -15").Output()
	report += "🔌 <b>Redirecciones (NAT):</b>\n<pre>" + string(iptNat) + "</pre>\n"

	// 2. Status Servicios
	svcs := []string{
		"badvpn.service",
		"udp-custom.service",
		"ssh-ws.service",
		"ssh-ws-pro.service",
		"haproxy.service",
		"dropbear_custom.service",
		"zivpn.service",
		"falconproxy.service",
	}
	report += "⚙️ <b>Service Status:</b>\n"
	for _, s := range svcs {
		active, _ := exec.Command("systemctl", "is-active", s).Output()
		status := strings.TrimSpace(string(active))
		if status == "" {
			status = "no encontrado"
		}
		report += fmt.Sprintf("• %s: <code>%s</code>\n", s, status)
	}

	// 3. RAM
	free, _ := exec.Command("free", "-m").Output()
	report += "\n💾 <b>Memoria RAM (MB):</b>\n<pre>" + string(free) + "</pre>"

	return report
}

// RestoreIptablesRules re-applies the NAT configuration required for some protocols.
// Since IPTables rules are lost on reboot on most unmanaged systems, this function is called
// by the bot on startup to ensure VPN routing (like UDP redirects for ZiVPN/SlowDNS) works.
func RestoreIptablesRules() {
	// 1. DNS Protocols Multiplexing (SlowDNS, VayDNS, Slipstream, Dnsdist)
	// Sincronizar y generar las reglas U32 e IPv6 correctas para todos los protocolos DNS.
	SyncDNSDist()

	// 1b. Slipstream (Alternativa nativa QUIC en UDP 443)
	if _, err := os.Stat("/etc/slipstream/cert.pem"); err == nil {
		exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("ip6tables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("ip6tables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
	}

	// 2. ZiVPN (Redirect UDP range 6000-19999)
	if _, err := os.Stat("/etc/zivpn/config.json"); err == nil {
		data, err := os.ReadFile("/etc/zivpn/config.json")
		if err == nil {
			var config ZivpnConfig
			if err := json.Unmarshal(data, &config); err == nil {
				// Listen is usually ":5667", so we extract the port
				parts := strings.Split(config.Listen, ":")
				if len(parts) >= 2 {
					port := parts[len(parts)-1]
					devOut, _ := exec.Command("bash", "-c", "ip -4 route show default | awk '{print $5}' | head -1").Output()
					dev := strings.TrimSpace(string(devOut))
					if dev == "" {
						devOut, _ = exec.Command("bash", "-c", "ip link show up | grep -v loopback | grep -v 'lo:' | head -1 | awk '{print $2}' | cut -d':' -f1").Output()
						dev = strings.TrimSpace(string(devOut))
					}
					if dev != "" {
						// CLEAN: Wipe old rules to avoid duplication
						exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables -t nat $line; done").Run()
						exec.Command("bash", "-c", "iptables -S INPUT | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables $line; done").Run()
						exec.Command("bash", "-c", "iptables -S INPUT | grep -w '"+port+"' | sed 's/-A/-D/' | while read line; do iptables $line; done").Run()

						// APPLY new rules
						_ = exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "1", "-i", dev, "-p", "udp", "--dport", "6000:19999", "-j", "REDIRECT", "--to-port", port).Run()
						_ = exec.Command("iptables", "-I", "INPUT", "1", "-p", "udp", "--dport", port, "-j", "ACCEPT").Run()
						_ = exec.Command("iptables", "-I", "INPUT", "1", "-p", "udp", "--dport", "6000:19999", "-j", "ACCEPT").Run()
						_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run()
						_ = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run()
					}
				}
			}
		}
	}
}
