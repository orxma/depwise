package sys

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

// CountZivpnActive returns true if any UDP session exists for zivpn
func CountZivpnActive() bool {
	out, err := exec.Command("sh", "-c", "ss -u -n -p | grep 'zivpn' | wc -l").Output()
	if err != nil {
		return false
	}
	count := strings.TrimSpace(string(out))
	return count != "" && count != "0"
}

// AutoCleanupLoop corre en un hilo separado ejecutando la limpieza de Iptables
// y usuarios excedidos cada cierto tiempo.
func AutoCleanupLoop(b *tele.Bot) {

	tick := 0
	lastUpdateCheck := time.Now()
	for {
		// Revisar límites de conexión activa cada 14 segundos (2 ticks)
		if tick%2 == 0 {
			EnforceConnectionLimits()
		}

		// 1. Limpieza de usuarios vencidos y AutoReboot de forma periódica
		if tick >= 9 { // Cada 60-70 segundos aprox
			// Guardar el tráfico en DB para que persista tras reiniciar la VPS
			GetGlobalTraffic()

			// Listas para recopilar expirados DENTRO del lock,
			// y ejecutar las operaciones pesadas FUERA del lock.
			var expiredSSH []string
			var expiredZivpn []string
			var expiredXray []string
			var expiredAdmins []int64 // Telegram IDs para notificar
			var shouldReboot bool

			db.Update(func(data *db.ConfigData) error {
				now := time.Now()
				nowStr := now.Format("2006-01-02")

				// REBOOT AUTOMÁTICO POR UPTIME (24 HORAS)
				if data.AutoReboot {
					outUptime, err := exec.Command("awk", "{print $1}", "/proc/uptime").Output()
					if err == nil {
						uptimeSecStr := strings.TrimSpace(string(outUptime))
						var uptimeSecFloat float64
						fmt.Sscanf(uptimeSecStr, "%f", &uptimeSecFloat)
						
						if uptimeSecFloat >= 86400 {
							shouldReboot = true
						}
					}
				}

				// Detectar SSH expirados (solo borrar de DB, NO ejecutar comandos)
				for user, expire := range data.SSHTimeUsers {
					if nowStr >= expire {
						expiredSSH = append(expiredSSH, user)
						delete(data.SSHTimeUsers, user)
						delete(data.SSHOwners, user)
						delete(data.SSHLastActive, user)
						delete(data.SSHBannerTitles, user)
						delete(data.SSHHandles, user)
						delete(data.Alerts1DaySent, "SSH:"+user)
						delete(data.Alerts1HourSent, "SSH:"+user)
					}
				}

				// Detectar ZiVPN expirados (solo borrar de DB)
				for pass, expire := range data.ZivpnUsers {
					if nowStr >= expire {
						expiredZivpn = append(expiredZivpn, pass)
						delete(data.ZivpnUsers, pass)
						delete(data.ZivpnOwners, pass)
						delete(data.ZivpnLastActive, pass)
						delete(data.ZivpnHandles, pass)
						delete(data.Alerts1DaySent, "ZiVPN:"+pass)
						delete(data.Alerts1HourSent, "ZiVPN:"+pass)
					}
				}

				// Detectar Xray expirados (solo borrar de DB)
				for uid, user := range data.XrayUsers {
					if nowStr >= user.Expire {
						expiredXray = append(expiredXray, uid)
						delete(data.XrayUsers, uid)
						delete(data.Alerts1DaySent, "Xray/V2Ray:"+uid)
						delete(data.Alerts1HourSent, "Xray/V2Ray:"+uid)
					}
				}

				// Detectar Admins expirados (solo borrar de DB)
				for adminID, adminInfo := range data.Admins {
					if adminInfo.Expire != "" && nowStr >= adminInfo.Expire {
						delete(data.Admins, adminID)
						id, errParse := strconv.ParseInt(adminID, 10, 64)
						if errParse == nil {
							expiredAdmins = append(expiredAdmins, id)
						}
					}
				}

				return nil
			})

			// === EJECUTAR OPERACIONES PESADAS FUERA DEL LOCK ===

			// Borrar usuarios SSH del sistema operativo
			for _, user := range expiredSSH {
				DeleteSSHUser(user)
			}

			// Borrar usuarios ZiVPN del servicio
			for _, pass := range expiredZivpn {
				vpn.RemoveZivpnUser(pass)
			}

			// Borrar usuarios Xray del servicio
			for _, uid := range expiredXray {
				vpn.RemoveXrayUser(uid)
			}

			// Notificar admins expirados (llamada de red fuera del lock)
			for _, id := range expiredAdmins {
				b.Send(&tele.User{ID: id}, "⏳ <b>Administrator Subscription Expired</b>\n\nYour administrator access time has ended. Your account has reverted to a normal user.\n\nThank you for using our service.", tele.ModeHTML)
			}

			if len(expiredSSH) > 0 {
				SyncSSHDBanners()
			}

			if shouldReboot {
				go func() {
					time.Sleep(2 * time.Second)
					exec.Command("reboot").Run()
				}()
			}

			// === AUTO-UPDATE CHECK (CADA 12 HORAS) ===
			if time.Since(lastUpdateCheck) > 12*time.Hour {
				lastUpdateCheck = time.Now()
				
				var autoUpdateEnabled bool
				db.Update(func(d *db.ConfigData) error {
					autoUpdateEnabled = d.AutoUpdate
					return nil
				})

				if autoUpdateEnabled {
					hasUpdate, _, err := CheckForUpdate()
					if err == nil && hasUpdate {
						// Notificar a los admins y ejecutar actualización
						data, _ := db.Load()
						for adminID := range data.Admins {
							id, _ := strconv.ParseInt(adminID, 10, 64)
							b.Send(&tele.User{ID: id}, "🔄 <b>AUTOMATIC UPDATE</b>\nA new version of the bot was detected on GitHub. Applying the update in the background...")
						}
						RunUpdate()
					}
				}
			}
			// ==========================================

			// Liberar memoria RAM inactiva al Sistema Operativo
			debug.FreeOSMemory()

			// Regenerar banners de usuarios SSH para actualizar días restantes
			RefreshAllBanners()

			// Nueva Ejecución: Limpieza cada 60s terminada
			tick = 0
		}

		tick++
		time.Sleep(7 * time.Second)
	}
}
