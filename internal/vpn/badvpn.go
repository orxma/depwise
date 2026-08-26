package vpn

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// badvpnPorts son los puertos donde escucha BadVPN (como en producción)
var badvpnPorts = []string{"7100", "7200", "7300"}

// badvpnBin es la ruta del binario custom de BadVPN (soporta multi-listen-addr y RakNet/Minecraft)
const badvpnBin = "/usr/bin/badvpn"

// InstallBadVPN compila e instala BadVPN desde fuente
// en múltiples puertos (7100, 7200, 7300) con un solo servicio.
// Este binario soporta multi-listen-addr y maneja mejor juegos como Minecraft Bedrock.
func InstallBadVPN(port string) error {
	// 1. Dependencias de compilación
	_ = exec.Command("apt-get", "update").Run()
	_ = exec.Command("apt-get", "install", "-y", "build-essential", "cmake", "git", "gcc", "g++").Run()

	// 2. Compilar badvpn desde fuente
	if _, err := os.Stat(badvpnBin); os.IsNotExist(err) {
		if err := compileBadVPN(); err != nil {
			// Fallback: intentar instalar badvpn-udpgw del paquete del sistema
			return installBadVPNFallback()
		}
	}

	// 3. Verificar que el binario es ejecutable
	if err := exec.Command(badvpnBin, "--help").Run(); err != nil {
		os.Remove(badvpnBin)
		return installBadVPNFallback()
	}

	// 4. Limpiar servicios viejos (por-puerto)
	for _, p := range badvpnPorts {
		exec.Command("systemctl", "stop", "badvpn-"+p+".service").Run()
		exec.Command("systemctl", "disable", "badvpn-"+p+".service").Run()
		os.Remove("/etc/systemd/system/badvpn-" + p + ".service")
	}

	// 5. Crear servicio único con multi-listen-addr (como el servidor de producción)
	svc := `[Unit]
Description=BadVPN UDP Gateway (Multi-Port)
Documentation=https://t.me/orxtunnel
After=syslog.target network-online.target

[Service]
User=root
NoNewPrivileges=true
ExecStart=/usr/bin/badvpn --listen-addr 127.0.0.1:7100 --listen-addr 127.0.0.1:7200 --listen-addr 127.0.0.1:7300 --max-clients 500
Restart=on-failure
RestartPreventExitStatus=23
LimitNPROC=10000
LimitNOFILE=1000000

[Install]
WantedBy=multi-user.target`

	svcFile := "/etc/systemd/system/badvpn.service"
	if err := os.WriteFile(svcFile, []byte(svc), 0644); err != nil {
		return fmt.Errorf("failed to write badvpn.service: %v", err)
	}

	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "badvpn.service").Run()
	if err := exec.Command("systemctl", "restart", "badvpn.service").Run(); err != nil {
		return fmt.Errorf("failed to restart badvpn.service: %v", err)
	}

	// 6. Verificación
	time.Sleep(2 * time.Second)
	if err := exec.Command("systemctl", "is-active", "--quiet", "badvpn.service").Run(); err != nil {
		logCmd, _ := exec.Command("journalctl", "-u", "badvpn.service", "--no-pager", "-n", "10").Output()
		logs := string(logCmd)
		if logs == "" {
			logs = "Could not retrieve logs."
		}

		_ = exec.Command("systemctl", "stop", "badvpn.service").Run()
		_ = os.Remove(svcFile)
		_ = exec.Command("systemctl", "daemon-reload").Run()
		return fmt.Errorf("badvpn could not stay active.\n\n📝 <b>LOGS:</b>\n<pre>%s</pre>", logs)
	}

	return nil
}

// compileBadVPN compila badvpn-udpgw desde fuente
func compileBadVPN() error {
	tmpDir := "/tmp/badvpn-build"
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)

	// Clonar repositorio de badvpn
	cmd := exec.Command("git", "clone", "--depth=1", "https://github.com/ambrop72/badvpn.git", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone badvpn: %v", err)
	}

	// Compilar solo badvpn-udpgw
	buildDir := filepath.Join(tmpDir, "build")
	os.MkdirAll(buildDir, 0755)

	cmd = exec.Command("cmake", "..", "-DBUILD_CLIENT=OFF", "-DBUILD_SERVER=OFF", "-DBUILD_TUN2SOCKS=OFF")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmake failed: %v", err)
	}

	cmd = exec.Command("make", "-j$(nproc)", "badvpn-udpgw")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make failed: %v", err)
	}

	// Copiar binario a /usr/bin/badvpn
	src := filepath.Join(buildDir, "udpgw", "badvpn-udpgw")
	if err := os.WriteFile(badvpnBin, mustReadFile(src), 0755); err != nil {
		return fmt.Errorf("failed to install binary: %v", err)
	}

	// Limpiar
	os.RemoveAll(tmpDir)
	return nil
}

func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

// RemoveBadVPN detiene y elimina todos los servicios badvpn
func RemoveBadVPN() error {
	// Servicio único (custom binary)
	exec.Command("systemctl", "stop", "badvpn.service").Run()
	exec.Command("systemctl", "disable", "badvpn.service").Run()
	os.Remove("/etc/systemd/system/badvpn.service")

	// Servicios por-puerto (fallback)
	for _, p := range badvpnPorts {
		svcName := "badvpn-" + p
		exec.Command("systemctl", "stop", svcName+".service").Run()
		exec.Command("systemctl", "disable", svcName+".service").Run()
		os.Remove("/etc/systemd/system/" + svcName + ".service")
	}

	os.Remove("/usr/bin/badvpn")
	os.Remove("/usr/bin/badvpn-udpgw")
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}
