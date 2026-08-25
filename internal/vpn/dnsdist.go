package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/orxma/depwise/internal/db"
)

// EnsureDNSDistInstalled verifica y en caso necesario instala dnsdist
func EnsureDNSDistInstalled() error {
	if _, err := exec.LookPath("dnsdist"); err != nil {
		cmd := exec.Command("bash", "-c", "DEBIAN_FRONTEND=noninteractive apt-get install -yq dnsdist")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error instalando dnsdist: %v", err)
		}
	}
	return nil
}

// SyncDNSDist actualiza la configuración de dnsdist según los protocolos activos.
// Si no hay ninguno, detiene dnsdist.
func SyncDNSDist() error {
	data, err := db.Load()
	if err != nil {
		return err
	}

	hasSlowDNS := data.SlowDNS.NS != ""
	hasVayDNS := data.VayDNS.NS != ""
	hasSlipstream := data.Slipstream.NS != ""

	// Si no hay protocolos DNS, eliminar iptables, desactivar y salir
	if !hasSlowDNS && !hasVayDNS && !hasSlipstream {
		// Stop and disable
		exec.Command("systemctl", "stop", "dnsdist").Run()
		exec.Command("systemctl", "disable", "dnsdist").Run()
		
		// Remove iptables rules
		for {
			err := exec.Command("iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5380").Run()
			if err != nil {
				break
			}
		}
		return nil
	}

	if err := EnsureDNSDistInstalled(); err != nil {
		return err
	}

	// Generate config
	var sb strings.Builder
	sb.WriteString("-- Auto-generado por ORX TUNNEL Bot\n")
	sb.WriteString("setLocal(\"0.0.0.0:5380\")\n")
	sb.WriteString("addLocal(\"[::]:5380\")\n")
	sb.WriteString("addACL('0.0.0.0/0')\n")
	sb.WriteString("addACL('::/0')\n")

	if hasSlowDNS {
		sb.WriteString("newServer({address=\"127.0.0.1:5300\", name=\"slowdns\", pool=\"slowdns\"})\n")
		if data.SlowDNS.NS != "" {
			slowRegex := strings.ReplaceAll(data.SlowDNS.NS, ".", "\\\\.")
			sb.WriteString(fmt.Sprintf("addAction(RegexRule(\"%s\"), PoolAction(\"slowdns\"))\n", slowRegex))
		} else {
			sb.WriteString("addAction(AllRule(), PoolAction(\"slowdns\"))\n")
		}
	}
	
	if hasVayDNS {
		sb.WriteString("newServer({address=\"127.0.0.1:5301\", name=\"vaydns\", pool=\"vaydns\"})\n")
		if data.VayDNS.NS != "" {
			vayRegex := strings.ReplaceAll(data.VayDNS.NS, ".", "\\\\.")
			sb.WriteString(fmt.Sprintf("addAction(RegexRule(\"%s\"), PoolAction(\"vaydns\"))\n", vayRegex))
		} else {
			// If NS is empty but we have VayDNS, fallback
			if !hasSlowDNS || data.SlowDNS.NS != "" {
				sb.WriteString("addAction(AllRule(), PoolAction(\"vaydns\"))\n")
			}
		}
	}

	if hasSlipstream {
		sb.WriteString("newServer({address=\"127.0.0.1:5302\", name=\"slipstream\", pool=\"slipstream\"})\n")
		if data.Slipstream.NS != "" {
			slipRegex := strings.ReplaceAll(data.Slipstream.NS, ".", "\\\\.")
			sb.WriteString(fmt.Sprintf("addAction(RegexRule(\"%s\"), PoolAction(\"slipstream\"))\n", slipRegex))
		} else {
			if (!hasSlowDNS || data.SlowDNS.NS != "") && (!hasVayDNS || data.VayDNS.NS != "") {
				sb.WriteString("addAction(AllRule(), PoolAction(\"slipstream\"))\n")
			}
		}
	}

	err = os.WriteFile("/etc/dnsdist/dnsdist.conf", []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("error escribiendo dnsdist.conf: %v", err)
	}

	// Limpiar reglas de PREROUTING antiguas en el puerto 53 para evitar conflictos de multiplexacion
	exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep -e '--dport 53 ' | sed 's/-A /-D /' | while read rule; do iptables -t nat $rule; done").Run()
	exec.Command("bash", "-c", "ip6tables -t nat -S PREROUTING | grep -e '--dport 53 ' | sed 's/-A /-D /' | while read rule; do ip6tables -t nat $rule; done").Run()
	
	// Limpiar la de u32 también si existe
	exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep 'u32' | grep '5380' | sed 's/-A /-D /' | while read rule; do iptables -t nat $rule; done").Run()
	exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep 'u32' | grep '5353' | sed 's/-A /-D /' | while read rule; do iptables -t nat $rule; done").Run()

	if hasSlowDNS || hasVayDNS || hasSlipstream {
		// U32 match (IPv4): Solo captura el tráfico que sea estrictamente DNS (Preguntas=1, Respuestas=0) en el byte 12.
		exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "1", "-p", "udp", "--dport", "53", "-m", "u32", "--u32", "0>>22&0x3C@12=0x00010000", "-j", "REDIRECT", "--to-ports", "5380").Run()
		// Para IPv6 simplemente redirigimos todo, ya que las operadoras móviles usan NAT64/IPv6 y QUIC puro casi no se usa sobre IPv6 en estas VPNs.
		exec.Command("ip6tables", "-t", "nat", "-I", "PREROUTING", "1", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", "5380").Run()
	}
	// Restart service
	exec.Command("systemctl", "enable", "dnsdist").Run()
	exec.Command("systemctl", "restart", "dnsdist").Run()

	return nil
}
