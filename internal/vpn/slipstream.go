package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// InstallSlipstream downloads the slipstream server, generates TLS certs, and configures iptables + systemd.
func InstallSlipstream(domain, port string) error {
	arch := runtime.GOARCH
	var url string
	if arch == "amd64" {
		url = "https://github.com/Mygod/slipstream-rust/releases/download/v0.1.1/slipstream-linux-x86_64.tar.gz"
	} else if arch == "arm64" {
		url = "https://github.com/Mygod/slipstream-rust/releases/download/v0.1.1/slipstream-linux-arm64.tar.gz"
	} else {
		return fmt.Errorf("unsupported architecture for slipstream: %s", arch)
	}

	// Download and extract slipstream-server
	exec.Command("rm", "-rf", "/tmp/slipstream").Run()
	os.MkdirAll("/tmp/slipstream", 0755)

	cmd := exec.Command("curl", "-L", "-s", "-o", "/tmp/slipstream/slip.tar.gz", url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error downloading slipstream: %v", err)
	}

	cmd = exec.Command("tar", "-xzf", "/tmp/slipstream/slip.tar.gz", "-C", "/tmp/slipstream")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extrayendo slipstream: %v", err)
	}

	// The tar contains a folder slipstream-linux-<arch>/slipstream-server
	folderArch := "x86_64"
	if arch == "arm64" {
		folderArch = "arm64"
	}
	srcPath := fmt.Sprintf("/tmp/slipstream/slipstream-linux-%s/slipstream-server", folderArch)
	
	cmd = exec.Command("mv", srcPath, "/usr/bin/slipstream-server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error moviendo binario: %v", err)
	}
	os.Chmod("/usr/bin/slipstream-server", 0755)

	// Clean up /tmp
	exec.Command("rm", "-rf", "/tmp/slipstream").Run()

	// Generate TLS Certs
	os.MkdirAll("/etc/slipstream", 0755)
	if _, err := os.Stat("/etc/slipstream/cert.pem"); os.IsNotExist(err) {
		cmdSSL := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:4096", "-nodes",
			"-keyout", "/etc/slipstream/key.pem",
			"-out", "/etc/slipstream/cert.pem",
			"-days", "3650",
			"-subj", "/CN="+domain)
		if err := cmdSSL.Run(); err != nil {
			return fmt.Errorf("error generando certificado TLS: %v", err)
		}
	}

	// iptables UDP 53 to 5302 (Todo lo que no sea capturado por el u32 de DNS, caerá aquí)
	exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("ip6tables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("ip6tables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
	
	// iptables UDP 443 to 5302 (Alternativa nativa para QUIC)
	exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("ip6tables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
	exec.Command("ip6tables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()

	// Systemd Service
	svc := `[Unit]
Description=Slipstream ORX TUNNEL Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/slipstream
ExecStart=/usr/bin/slipstream-server -l 5302 -c /etc/slipstream/cert.pem -k /etc/slipstream/key.pem -d ` + domain + ` -a 127.0.0.1:` + port + `
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target`

	os.WriteFile("/etc/systemd/system/slipstream.service", []byte(svc), 0644)
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "slipstream").Run()
	exec.Command("systemctl", "restart", "slipstream").Run()

	return nil
}

// RemoveSlipstream stops and removes the slipstream daemon.
func RemoveSlipstream() error {
	exec.Command("systemctl", "stop", "slipstream").Run()
	exec.Command("systemctl", "disable", "slipstream").Run()
	os.Remove("/etc/systemd/system/slipstream.service")
	exec.Command("systemctl", "daemon-reload").Run()

	// Loop removal iptables
	for {
		err := exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("ip6tables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5302").Run()
		exec.Command("ip6tables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "443", "-j", "REDIRECT", "--to-ports", "5302").Run()
		if err != nil {
			break
		}
	}

	os.RemoveAll("/etc/slipstream")
	os.Remove("/usr/bin/slipstream-server")
	return nil
}
