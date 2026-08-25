package bot

import (
	"fmt"
	"time"

	"github.com/orxma/depwise/internal/db"
	tele "gopkg.in/telebot.v3"
)

// handleBackupMenu muestra el menú de opciones de respaldo
func handleBackupMenu(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	
	btnNow := markup.Data("🚀 Un solo uso (Enviar Ahora)", "backup_now")
	btn1Day := markup.Data("🕐 Automático cada 1 día", "backup_auto_1")
	btn3Days := markup.Data("🕒 Automático cada 3 días", "backup_auto_3")
	btn7Days := markup.Data("🕖 Automático cada 7 días", "backup_auto_7")
	btn30Days := markup.Data("📅 Automático cada 30 días", "backup_auto_30")
	btnOff := markup.Data("❌ Desactivar Automático", "backup_auto_0")
	btnBack := markup.Data("🔙 Volver", "menu_admins")

	markup.Inline(
		markup.Row(btnNow),
		markup.Row(btn1Day, btn3Days),
		markup.Row(btn7Days, btn30Days),
		markup.Row(btnOff),
		markup.Row(btnBack),
	)

	data, _ := db.Load()
	status := "Desactivado"
	if data.BackupIntervalDays > 0 {
		status = fmt.Sprintf("Cada %d días", data.BackupIntervalDays)
	}

	text := fmt.Sprintf("🗄 <b>Configuración de Backups</b>\n\n"+
		"Estado Automático: <b>%s</b>\n\n"+
		"Selecciona una opción para el respaldo de tu base de datos.", status)

	return SafeEditCtx(c, b, text, markup)
}

// handleSetBackupInterval configura el intervalo de respaldo
func handleSetBackupInterval(c tele.Context, b *tele.Bot, days int) error {
	chatID := c.Chat().ID
	
	err := db.Update(func(d *db.ConfigData) error {
		d.BackupIntervalDays = days
		if days > 0 {
			d.BackupChatID = chatID
		} else {
			d.BackupChatID = 0
		}
		return nil
	})

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver a Backups", "menu_backup")))

	if err != nil {
		return SafeEditCtx(c, b, "❌ Error guardando la configuración.", markup)
	}

	if days == 0 {
		return SafeEditCtx(c, b, "✅ <b>Backups Automáticos Desactivados.</b>", markup)
	}
	
	return SafeEditCtx(c, b, fmt.Sprintf("✅ <b>Backups Automáticos Activados.</b>\nSe enviarán cada %d días a este chat.", days), markup)
}

// handleLocalBackup envía el respaldo inmediatamente al chat
func handleLocalBackup(c tele.Context, b *tele.Bot) error {
	SafeEditCtx(c, b, "⏳ <i>Preparando y enviando copia de seguridad...</i>", nil)
	
	doc := &tele.Document{File: tele.FromDisk(db.GetDataPath())}
	doc.FileName = fmt.Sprintf("bot_data_%s.json", time.Now().Format("2006-01-02"))
	doc.Caption = "✅ <b>Copia de Seguridad de tu Base de Datos</b>"
	
	_, err := b.Send(c.Chat(), doc, &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
	
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_backup")))

	if err != nil {
		return SafeEditCtx(c, b, fmt.Sprintf("❌ <b>Error al enviar backup:</b>\n%v", err), markup)
	}

	db.Update(func(d *db.ConfigData) error {
		d.LocalLastBackup = time.Now().Format(time.RFC3339)
		return nil
	})

	return SafeEditCtx(c, b, "✅ <b>Copia de Seguridad enviada.</b>\nPor favor, guarda el archivo en un lugar seguro.", markup)
}

// handleLocalRestoreReq pide el archivo para restaurar
func handleLocalRestoreReq(c tele.Context, b *tele.Bot) error {
	SetUserStep(c.Chat().ID, "awaiting_backup_restore")
	
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))
	
	return SafeEditCtx(c, b, "📥 <b>Restaurar Base de Datos</b>\n\nPor favor, <b>envía ahora mismo</b> el archivo `.json` de respaldo a este chat.\n\n⚠️ <i>Asegúrate de que el archivo sea un respaldo válido.</i>", markup)
}

// handleRestoreDocument processa el documento subido para restaurar
func handleRestoreDocument(c tele.Context, b *tele.Bot) error {
	doc := c.Message().Document
	if doc == nil || doc.MIME != "application/json" && doc.MIME != "text/plain" {
		return c.Send("❌ <b>Formato inválido.</b> Por favor envía un archivo .json válido.", &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	msg, _ := b.Send(c.Chat(), "⏳ <i>Descargando y aplicando copia de seguridad...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

	err := b.Download(&doc.File, db.GetDataPath())
	if err != nil {
		_, errEdit := b.Edit(msg, fmt.Sprintf("❌ <b>Error al descargar archivo:</b>\n%v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return errEdit
	}

	DeleteUserStep(c.Chat().ID)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver al Inicio", "menu_admins")))
	
	_, errEdit := b.Edit(msg, "✅ <b>Base de Datos Restaurada Exitosamente!</b>\nLos datos se han cargado.\n\n⚠️ <i>Te recomiendo presionar 'Reiniciar VPS' en Ajustes Pro para aplicar cambios por completo si restauraste configuraciones de red.</i>", &tele.SendOptions{
		ParseMode: tele.ModeHTML,
		ReplyMarkup: markup,
	})
	return errEdit
}

// autoBackupLoop es un demonio que corre en segundo plano y verifica si es hora de enviar un respaldo automático.
func autoBackupLoop(b *tele.Bot) {
	time.Sleep(1 * time.Minute) // Esperar 1 minuto tras iniciar
	
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		data, err := db.Load()
		if err != nil || data.BackupIntervalDays <= 0 || data.BackupChatID == 0 {
			continue
		}
		
		needsBackup := false
		if data.LocalLastBackup == "" {
			needsBackup = true
		} else {
			lastBackup, errParse := time.Parse(time.RFC3339, data.LocalLastBackup)
			if errParse == nil {
				// Han pasado más horas que las configuradas?
				if time.Since(lastBackup).Hours() >= float64(data.BackupIntervalDays*24) {
					needsBackup = true
				}
			} else {
				needsBackup = true
			}
		}

		if needsBackup {
			fmt.Println("Ejecutando ciclo de respaldo automático por Telegram...")
			doc := &tele.Document{File: tele.FromDisk(db.GetDataPath())}
			doc.FileName = fmt.Sprintf("bot_data_auto_%s.json", time.Now().Format("2006-01-02"))
			doc.Caption = fmt.Sprintf("🤖 <b>Respaldo Automático (%d días)</b>", data.BackupIntervalDays)
			
			_, errSend := b.Send(&tele.Chat{ID: data.BackupChatID}, doc, &tele.SendOptions{
				ParseMode: tele.ModeHTML,
			})
			
			if errSend == nil {
				fmt.Println("✅ Backup automático enviado por Telegram.")
				db.Update(func(d *db.ConfigData) error {
					d.LocalLastBackup = time.Now().Format(time.RFC3339)
					return nil
				})
			} else {
				fmt.Printf("❌ Error al enviar backup automático: %v\n", errSend)
			}
		}
	}
}
