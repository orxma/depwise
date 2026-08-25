package sys

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
)

const (
	bannerDir       = "/etc/ssh_banners"
	sshdConfig      = "/etc/ssh/sshd_config"
	bannerMarkerStart = "# >>> ORX TUNNEL_USER_BANNERS_START <<<"
	bannerMarkerEnd   = "# >>> ORX TUNNEL_USER_BANNERS_END <<<"
)

// GenerateUserBanner genera el contenido HTML del banner para un usuario SSH
// Compatible con HTTP Injector, HTTP Custom, HA Tunnel y apps VPN
func GenerateUserBanner(username, title string, limit int, expireDate string, data *db.ConfigData) string {
	if title == "" {
		title = "INTERNET ILIMITADO"
	}

	promoText := "🔥 ¡SERVIDORES PREMIUM! 🔥"
	if data != nil && data.BannerPromoText != "" {
		promoText = data.BannerPromoText
	}

	promoChannel := "@orxtunnel"
	if data != nil && data.BannerPromoChannel != "" {
		promoChannel = data.BannerPromoChannel
	}

	promoSupport := "@KTTOFICIAL"
	if data != nil && data.BannerPromoSupport != "" {
		promoSupport = data.BannerPromoSupport
	}

	promoBotName := "@orxtunnel_bot"
	if data != nil && data.BannerPromoBotName != "" {
		promoBotName = data.BannerPromoBotName
	}

	// Calcular días restantes
	daysLeft := 0
	parsed, err := time.Parse("2006-01-02", expireDate)
	if err == nil {
		daysLeft = int(math.Ceil(time.Until(parsed).Hours() / 24))
		if daysLeft < 0 {
			daysLeft = 0
		}
	}

	limitStr := fmt.Sprintf("%d", limit)
	if limit <= 0 {
		limitStr = "∞ Ilimitado"
	}

	var b strings.Builder

	b.WriteString("<html>\n")

	// Separador superior
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	// Logo braille ORX TUNNEL (probado y funcional en HTTP Injector)
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font face=\"monospace\" color=\"#00ff00\">")
	b.WriteString("⠀⠀⢀⣶⡆kevin tech tutorials⢰⣶⡀⠀⠀<br>")
	b.WriteString("</font>")
	b.WriteString("</h5>\n")

	// Texto ORX TUNNEL
	b.WriteString("<h1 style=\"text-align:center;\">")
	b.WriteString("<font face=\"monospace\" color=\"#00ff00\"><b>ORX TUNNEL</b></font>")
	b.WriteString("</h1>\n")

	// Separador
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	// Título personalizado
	b.WriteString("<h3 style=\"text-align:center;\">")
	b.WriteString(fmt.Sprintf("<font color='#FF00FF'><b>⚡ %s ⚡</b></font>", title))
	b.WriteString("</h3>\n")

	// Separador
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	// Datos de la cuenta — cada dato en su propia línea con <br>
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>👤 Usuario: </font><font color='#f1c40f'><b>%s</b></font><br>", username))
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>📅 Vence: </font><font color='#f1c40f'><b>%s</b></font><br>", expireDate))
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>⏳ Días Restant.: </font><font color='#f1c40f'><b>%d</b></font><br>", daysLeft))
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>💻 Límite: </font><font color='#f1c40f'><b>%s</b></font>", limitStr))
	b.WriteString("</h5>\n")

	// Separador
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	// Promoción
	b.WriteString("<h4 style=\"text-align:center;\">")
	b.WriteString(fmt.Sprintf("<font color='#FF00FF'><b>%s</b></font>", promoText))
	b.WriteString("</h4>\n")

	// Contacto — cada uno en su línea
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>📢 Canal: </font><a href=\"https://t.me/%s\"><font color='#f1c40f'>%s</font></a><br>", strings.TrimPrefix(promoChannel, "@"), promoChannel))
	b.WriteString(fmt.Sprintf("<font color='#ffffff'>👤 Soporte: </font><a href=\"https://t.me/%s\"><font color='#f1c40f'>%s</font></a>", strings.TrimPrefix(promoSupport, "@"), promoSupport))
	b.WriteString("</h5>\n")

	// Separador
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	// Crédito
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString(fmt.Sprintf("<font color='#00e676'><b>✅ CREADO EN : %s</b></font>", promoBotName))
	b.WriteString("</h5>\n")

	// Línea inferior
	b.WriteString("<h5 style=\"text-align:center;\">")
	b.WriteString("<font color='#29b6f6'>══════════════════════</font>")
	b.WriteString("</h5>\n")

	b.WriteString("</html>\n")

	return b.String()
}

// WriteUserBanner genera y escribe el banner de un usuario en /etc/ssh_banners/
func WriteUserBanner(username, title string, limit int, expireDate string, data *db.ConfigData) error {
	if err := os.MkdirAll(bannerDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de banners: %v", err)
	}

	content := GenerateUserBanner(username, title, limit, expireDate, data)
	path := filepath.Join(bannerDir, username+".banner")
	return os.WriteFile(path, []byte(content), 0644)
}

// RemoveUserBanner elimina el banner de un usuario
func RemoveUserBanner(username string) error {
	path := filepath.Join(bannerDir, username+".banner")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// EnsureBannerSystem configura sshd_config con Match User blocks para cada usuario SSH
func EnsureBannerSystem() error {
	if err := os.MkdirAll(bannerDir, 0755); err != nil {
		return err
	}
	return SyncSSHDBanners()
}

// SyncSSHDBanners actualiza los bloques Match User en sshd_config para apuntar
// al banner individual de cada usuario SSH
func SyncSSHDBanners() error {
	data, err := db.Load()
	if err != nil {
		return err
	}

	// Leer sshd_config actual
	raw, err := os.ReadFile(sshdConfig)
	if err != nil {
		return fmt.Errorf("no se pudo leer sshd_config: %v", err)
	}

	content := string(raw)

	// Eliminar bloque anterior de ORX TUNNEL si existe
	if idx := strings.Index(content, bannerMarkerStart); idx >= 0 {
		endIdx := strings.Index(content, bannerMarkerEnd)
		if endIdx >= 0 {
			content = content[:idx] + content[endIdx+len(bannerMarkerEnd):]
		}
	}

	// Limpiar líneas vacías al final
	content = strings.TrimRight(content, "\n\t ") + "\n\n"

	// Construir nuevos bloques Match User
	var matchBlocks strings.Builder
	matchBlocks.WriteString(bannerMarkerStart + "\n")

	for user := range data.SSHTimeUsers {
		bannerFile := filepath.Join(bannerDir, user+".banner")
		if _, err := os.Stat(bannerFile); err == nil {
			matchBlocks.WriteString(fmt.Sprintf("Match User %s\n", user))
			matchBlocks.WriteString(fmt.Sprintf("    Banner %s\n\n", bannerFile))
		}
	}

	matchBlocks.WriteString(bannerMarkerEnd + "\n")

	// Escribir sshd_config actualizado
	newContent := content + matchBlocks.String()
	if err := os.WriteFile(sshdConfig, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error escribiendo sshd_config: %v", err)
	}

	// Recargar SSH para aplicar
	exec.Command("systemctl", "reload", "ssh").Run()
	exec.Command("systemctl", "reload", "sshd").Run()

	return nil
}

// GetAllUserMaxLogins lee todos los límites de una sola vez
func GetAllUserMaxLogins() map[string]int {
	limits := make(map[string]int)
	configData, err := os.ReadFile("/etc/security/limits.conf")
	if err == nil {
		lines := strings.Split(string(configData), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 4 && fields[1] == "hard" && fields[2] == "maxlogins" {
				lim, _ := strconv.Atoi(fields[3])
				limits[fields[0]] = lim
			}
		}
	}
	return limits
}

// RefreshAllBanners regenera los banners de todos los usuarios SSH activos
func RefreshAllBanners() {
	data, err := db.Load()
	if err != nil {
		return
	}

	// Solo regenerar si hay usuarios SSH
	if len(data.SSHTimeUsers) == 0 {
		return
	}

	// Asegurar que existe el directorio
	os.MkdirAll(bannerDir, 0755)
	
	// Leer todos los límites de una vez (Optimización CPU)
	limits := GetAllUserMaxLogins()

	for user, expire := range data.SSHTimeUsers {
		title := ""
		if data.SSHBannerTitles != nil {
			title = data.SSHBannerTitles[user]
		}
		limit := limits[user]
		WriteUserBanner(user, title, limit, expire, data)
	}

	// NOTA: No llamamos a SyncSSHDBanners() aquí para evitar 
	// recargas innecesarias (systemctl reload ssh) cada minuto,
	// lo cual causaba el uso alto de CPU en la VPS.
}
