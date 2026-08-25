package sys

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsScannerToolInstalled verifica si una herramienta específica del escáner está instalada
func IsScannerToolInstalled(name string) bool {
	return findGoBinary(name) != ""
}

// GetScannerStatus devuelve el estado de las herramientas del escáner
func GetScannerStatus() (assetfinderOK bool, httpxOK bool) {
	return IsScannerToolInstalled("assetfinder"), IsScannerToolInstalled("httpx")
}

// InstallScannerTool instala una herramienta específica del escáner con progreso
// Retorna un canal que emite mensajes de progreso y un canal de error final
func InstallScannerTool(name string) error {
	goPath := findGoPath()
	if goPath == "" {
		return fmt.Errorf("Go is not installed on this server")
	}

	var pkg string
	switch name {
	case "assetfinder":
		pkg = "github.com/tomnomnom/assetfinder@latest"
	case "httpx":
		pkg = "github.com/projectdiscovery/httpx/cmd/httpx@latest"
	default:
		return fmt.Errorf("herramienta desconocida: %s", name)
	}

	// Usar -v para output verboso (muestra paquetes conforme se descargan)
	cmd := exec.Command(goPath, "install", pkg)
	cmd.Env = append(os.Environ(),
		"GOPATH="+getGoPathDir(),
		"PATH="+os.Getenv("PATH")+":/usr/local/go/bin",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error installing %s: %v\nOutput: %s", name, err, string(out))
	}

	// Verificar que realmente se instaló
	if !IsScannerToolInstalled(name) {
		// Intentar vincular al PATH
		binPath := getGoPathDir() + "/bin/" + name
		if _, err := os.Stat(binPath); err == nil {
			exec.Command("ln", "-sf", binPath, "/usr/local/bin/"+name).Run()
		}
	}

	return nil
}

// UninstallScannerTool desinstala una herramienta del escáner
func UninstallScannerTool(name string) error {
	path := findGoBinary(name)
	if path == "" {
		return nil // Ya no existe
	}

	// Eliminar el binario
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("error eliminando %s: %v", name, err)
	}

	// También eliminar del /usr/local/bin si hay un symlink
	localBin := "/usr/local/bin/" + name
	if p, _ := os.Readlink(localBin); p != "" {
		os.Remove(localBin)
	} else if _, err := os.Stat(localBin); err == nil {
		os.Remove(localBin)
	}

	return nil
}

// UninstallAllScannerTools desinstala todas las herramientas del escáner
func UninstallAllScannerTools() error {
	var errs []string
	if err := UninstallScannerTool("assetfinder"); err != nil {
		errs = append(errs, err.Error())
	}
	if err := UninstallScannerTool("httpx"); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// EnsureScannerDeps checks and installs assetfinder and httpx if missing
func EnsureScannerDeps() error {
	goPath := findGoPath()
	if goPath == "" {
		return fmt.Errorf("Go is not installed. Please install Go first")
	}

	tools := map[string]string{
		"assetfinder": "github.com/tomnomnom/assetfinder@latest",
		"httpx":       "github.com/projectdiscovery/httpx/cmd/httpx@latest",
	}

	for name, pkg := range tools {
		if findGoBinary(name) == "" {
			out, err := exec.Command(goPath, "install", "-v", pkg).CombinedOutput()
			if err != nil {
				return fmt.Errorf("error installing %s: %v\nOutput: %s", name, err, string(out))
			}
		}
	}

	return nil
}

func findGoPath() string {
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	commonPaths := []string{"/usr/local/go/bin/go", "/usr/bin/go"}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getGoPathDir() string {
	goCmd := findGoPath()
	if goCmd == "" {
		return "/root/go"
	}
	if out, err := exec.Command(goCmd, "env", "GOPATH").Output(); err == nil {
		p := strings.TrimSpace(string(out))
		if p != "" {
			return p
		}
	}
	return "/root/go"
}

func findGoBinary(name string) string {
	// 1. Try PATH
	if p, err := exec.LookPath(name); err == nil {
		return p
	}

	// 2. Try to get GOPATH bin manually using a likely go binary location if not in PATH
	goCmd := "go"
	if _, err := exec.LookPath("go"); err != nil {
		if err := exec.Command("ls", "/usr/local/go/bin/go").Run(); err == nil {
			goCmd = "/usr/local/go/bin/go"
		}
	}

	if out, err := exec.Command(goCmd, "env", "GOPATH").Output(); err == nil {
		gopath := strings.TrimSpace(string(out))
		if gopath != "" {
			p := fmt.Sprintf("%s/bin/%s", gopath, name)
			if err := exec.Command("ls", p).Run(); err == nil {
				return p
			}
		}
	}

	// 3. Try common VPS Go bin paths
	paths := []string{
		"/usr/local/bin/" + name,
		"/usr/bin/" + name,
		"/bin/" + name,
		"/root/go/bin/" + name,
		"/usr/local/go/bin/" + name,
		"/home/ubuntu/go/bin/" + name,
		"/home/debian/go/bin/" + name,
	}

	for _, p := range paths {
		if err := exec.Command("ls", p).Run(); err == nil {
			return p
		}
	}
	return ""
}

// RunScanner runs assetfinder and httpx on a domain
func RunScanner(domain string) (string, error) {
	// Resolve paths
	assetPath := findGoBinary("assetfinder")
	if assetPath == "" {
		return "", fmt.Errorf("assetfinder not found. Install it from the Protocols > Scanner menu.")
	}

	httpxPath := findGoBinary("httpx")
	if httpxPath == "" {
		return "", fmt.Errorf("httpx not found. Install it from the Protocols > Scanner menu.")
	}

	// 1. Assetfinder
	cmdAsset := exec.Command(assetPath, "--subs-only", domain)
	outAsset, err := cmdAsset.Output()
	if err != nil {
		return "", fmt.Errorf("error en assetfinder (%s): %v", assetPath, err)
	}

	subs := strings.TrimSpace(string(outAsset))
	if subs == "" {
		return "❌ No subdomains found.", nil
	}

	// 2. HTTPX (using stdin)
	cmdHttpx := exec.Command(httpxPath, "-silent", "-status-code", "-title", "-tech-detect", "-ip")
	cmdHttpx.Stdin = strings.NewReader(subs)
	outHttpx, err := cmdHttpx.Output()
	if err != nil {
		return "", fmt.Errorf("error en httpx (%s): %v", httpxPath, err)
	}

	result := string(outHttpx)
	if result == "" {
		return "🔍 Subdomains found, but none responded on HTTP/HTTPS.", nil
	}

	return result, nil
}
