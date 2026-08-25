package sys

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// CurrentVersion indica la versión actual en ejecución
	CurrentVersion = "7.9"
	// RemoteVersionURL es el archivo en GitHub que dice la última versión disponible
	RemoteVersionURL = "https://raw.githubusercontent.com/orxma/depwise/main/version.txt"
)

// CheckForUpdate verifica si hay una actualización disponible comparando la versión local con la remota.
// Retorna (hayActualizacion, nuevaVersion, error)
func CheckForUpdate() (bool, string, error) {
	// Agregamos un timestamp para evitar la caché agresiva de 5 minutos de raw.githubusercontent.com
	urlWithNoCache := fmt.Sprintf("%s?t=%d", RemoteVersionURL, time.Now().Unix())
	resp, err := http.Get(urlWithNoCache)
	if err != nil {
		return false, "", fmt.Errorf("error conectando con GitHub: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("código HTTP inesperado: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("error leyendo versión remota: %v", err)
	}

	remoteVerStr := strings.TrimSpace(string(body))
	if remoteVerStr == "" {
		return false, "", fmt.Errorf("archivo de versión remoto vacío")
	}

	// Comparación muy simple asumiendo formato "7.4", "7.5", etc.
	localVer, errL := strconv.ParseFloat(CurrentVersion, 64)
	remoteVer, errR := strconv.ParseFloat(remoteVerStr, 64)

	if errL != nil || errR != nil {
		// Si no se puede parsear como flotante (ej: 7.4.1), comparamos como strings básicos.
		if remoteVerStr != CurrentVersion {
			return true, remoteVerStr, nil
		}
		return false, remoteVerStr, nil
	}

	if remoteVer > localVer {
		return true, remoteVerStr, nil
	}

	return false, remoteVerStr, nil
}

// RunUpdate lanza el proceso de actualización en segundo plano.
// Desvincula el proceso del bot para que sobreviva al reinicio del servicio.
func RunUpdate() error {
	updateScript := `#!/bin/bash
sleep 2
cd /tmp
rm -rf depwise
git clone https://github.com/orxma/depwise.git
cd depwise
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o /usr/local/bin/orxtunnel-bot cmd/orxtunnel/main.go
systemctl restart orxtunnel
`
	err := os.WriteFile("/tmp/orxtunnel_update.sh", []byte(updateScript), 0755)
	if err != nil {
		return fmt.Errorf("error creando script de actualización: %v", err)
	}

	unitName := fmt.Sprintf("orxtunnel-updater-%d", time.Now().Unix())
	cmd := exec.Command("systemd-run", "--unit="+unitName, "/tmp/orxtunnel_update.sh")
	err = cmd.Start()
	if err != nil {
		// Fallback por si systemd-run falla
		cmdFallback := exec.Command("sh", "-c", `nohup /tmp/orxtunnel_update.sh > /dev/null 2>&1 &`)
		return cmdFallback.Start()
	}
	
	return nil
}
