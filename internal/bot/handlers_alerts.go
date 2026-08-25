package bot

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	tele "gopkg.in/telebot.v3"
)

func autoExpirationAlertLoop(b *tele.Bot) {
	// Pequeño retraso al inicio para estabilización
	time.Sleep(2 * time.Minute)
	
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		data, err := db.Load()
		if err == nil {
			now := time.Now()
			tomorrowDate := now.AddDate(0, 0, 1).Format("2006-01-02")
			currentHour := now.Hour()
			
			// Función helper para procesar la cuenta
			processAlert := func(service, cbType, identifier, expireDate, ownerStr string) {
				if ownerStr == "" {
					return
				}
				
				ownerID, errParse := strconv.ParseInt(ownerStr, 10, 64)
				if errParse != nil {
					return
				}
				
				// Clave única para los diccionarios
				alertKey := fmt.Sprintf("%s:%s", cbType, identifier)
				
				// ¿Es mañana la fecha de vencimiento?
				if expireDate == tomorrowDate {
					markup := b.NewMarkup()
					btnRenew := markup.Data(i18n.T(ownerID, "btn.renew"), "renew_"+cbType, identifier)
					markup.Inline(markup.Row(btnRenew))

					// 1. Alerta de 1 Día (si no se ha enviado aún)
					if !data.Alerts1DaySent[alertKey] {
						msg := i18n.Tf(ownerID, "alert.expiry_1day", service, identifier)
						_, errSend := b.Send(&tele.Chat{ID: ownerID}, msg, tele.ModeHTML, markup)
						if errSend == nil {
							db.Update(func(d *db.ConfigData) error {
								if d.Alerts1DaySent == nil {
									d.Alerts1DaySent = make(map[string]bool)
								}
								d.Alerts1DaySent[alertKey] = true
								return nil
							})
							// Recargar los datos en memoria para los próximos loops (en la misma pasada)
							data.Alerts1DaySent[alertKey] = true
						}
					}
					
					// 2. Alerta de 1 Hora (si no se ha enviado y ya son las 23:00+)
					if currentHour >= 23 && !data.Alerts1HourSent[alertKey] {
						msg := i18n.Tf(ownerID, "alert.expiry_1hour", service, identifier)
						_, errSend := b.Send(&tele.Chat{ID: ownerID}, msg, tele.ModeHTML, markup)
						if errSend == nil {
							db.Update(func(d *db.ConfigData) error {
								if d.Alerts1HourSent == nil {
									d.Alerts1HourSent = make(map[string]bool)
								}
								d.Alerts1HourSent[alertKey] = true
								return nil
							})
							data.Alerts1HourSent[alertKey] = true
						}
					}
				}
			}
			
			// SSH
			for user, expire := range data.SSHTimeUsers {
				processAlert("SSH", "ssh", user, expire, data.SSHOwners[user])
			}
			
			// ZiVPN
			for pass, expire := range data.ZivpnUsers {
				processAlert("ZiVPN", "zi", pass, expire, data.ZivpnOwners[pass])
			}
			
			// Xray
			for uid, user := range data.XrayUsers {
				processAlert("Xray/V2Ray", "xray", uid, user.Expire, user.Owner)
			}
		} else {
			log.Printf("Error reading DB for alerts: %v", err)
		}
		
		<-ticker.C
	}
}
