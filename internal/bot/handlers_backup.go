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
	
	btnNow := markup.Data("🚀 One-time (Send Now)", "backup_now")
	btn1Day := markup.Data("🕐 Automatic every 1 day", "backup_auto_1")
	btn3Days := markup.Data("🕒 Automatic every 3 days", "backup_auto_3")
	btn7Days := markup.Data("🕖 Automatic every 7 days", "backup_auto_7")
	btn30Days := markup.Data("📅 Automatic every 30 days", "backup_auto_30")
	btnOff := markup.Data("❌ Disable Automatic", "backup_auto_0")
	btnBack := markup.Data("🔙 Back", "menu_admins")

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
		status = fmt.Sprintf("Every %d days", data.BackupIntervalDays)
	}

	text := fmt.Sprintf("🗄 <b>Backup Configuration</b>\n\n"+
		"Automatic Status: <b>%s</b>\n\n"+
		"Select an option for backing up your database.", status)

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
	markup.Inline(markup.Row(markup.Data("🔙 Back to Backups", "menu_backup")))

	if err != nil {
		return SafeEditCtx(c, b, "❌ Error saving the configuration.", markup)
	}

	if days == 0 {
		return SafeEditCtx(c, b, "✅ <b>Automatic Backups Disabled.</b>", markup)
	}
	
	return SafeEditCtx(c, b, fmt.Sprintf("✅ <b>Automatic Backups Enabled.</b>\nThey will be sent to this chat every %d days.", days), markup)
}

// handleLocalBackup envía el respaldo inmediatamente al chat
func handleLocalBackup(c tele.Context, b *tele.Bot) error {
	SafeEditCtx(c, b, "⏳ <i>Preparing and sending the backup...</i>", nil)
	
	doc := &tele.Document{File: tele.FromDisk(db.GetDataPath())}
	doc.FileName = fmt.Sprintf("bot_data_%s.json", time.Now().Format("2006-01-02"))
	doc.Caption = "✅ <b>Database Backup</b>"
	
	_, err := b.Send(c.Chat(), doc, &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	})
	
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_backup")))

	if err != nil {
		return SafeEditCtx(c, b, fmt.Sprintf("❌ <b>Error sending backup:</b>\n%v", err), markup)
	}

	db.Update(func(d *db.ConfigData) error {
		d.LocalLastBackup = time.Now().Format(time.RFC3339)
		return nil
	})

	return SafeEditCtx(c, b, "✅ <b>Backup sent.</b>\nPlease store the file somewhere safe.", markup)
}

// handleLocalRestoreReq pide el archivo para restaurar
func handleLocalRestoreReq(c tele.Context, b *tele.Bot) error {
	SetUserStep(c.Chat().ID, "awaiting_backup_restore")
	
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))
	
	return SafeEditCtx(c, b, "📥 <b>Restore Database</b>\n\nPlease <b>send now</b> the `.json` backup file to this chat.\n\n⚠️ <i>Make sure the file is a valid backup.</i>", markup)
}

// handleRestoreDocument processa el documento subido para restaurar
func handleRestoreDocument(c tele.Context, b *tele.Bot) error {
	doc := c.Message().Document
	if doc == nil || doc.MIME != "application/json" && doc.MIME != "text/plain" {
		return c.Send("❌ <b>Invalid format.</b> Please send a valid .json file.", &tele.SendOptions{ParseMode: tele.ModeHTML})
	}

	msg, _ := b.Send(c.Chat(), "⏳ <i>Downloading and applying the backup...</i>", &tele.SendOptions{ParseMode: tele.ModeHTML})

	err := b.Download(&doc.File, db.GetDataPath())
	if err != nil {
		_, errEdit := b.Edit(msg, fmt.Sprintf("❌ <b>Error downloading file:</b>\n%v", err), &tele.SendOptions{ParseMode: tele.ModeHTML})
		return errEdit
	}

	DeleteUserStep(c.Chat().ID)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back to Start", "menu_admins")))
	
	_, errEdit := b.Edit(msg, "✅ <b>Database Restored Successfully!</b>\nThe data has been loaded.\n\n⚠️ <i>If you restored network settings, use 'Restart VPS' in Pro Settings to apply all changes.</i>", &tele.SendOptions{
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
			fmt.Println("Running automatic Telegram backup cycle...")
			doc := &tele.Document{File: tele.FromDisk(db.GetDataPath())}
			doc.FileName = fmt.Sprintf("bot_data_auto_%s.json", time.Now().Format("2006-01-02"))
			doc.Caption = fmt.Sprintf("🤖 <b>Automatic Backup (%d days)</b>", data.BackupIntervalDays)
			
			_, errSend := b.Send(&tele.Chat{ID: data.BackupChatID}, doc, &tele.SendOptions{
				ParseMode: tele.ModeHTML,
			})
			
			if errSend == nil {
				fmt.Println("✅ Automatic backup sent via Telegram.")
				db.Update(func(d *db.ConfigData) error {
					d.LocalLastBackup = time.Now().Format(time.RFC3339)
					return nil
				})
			} else {
				fmt.Printf("❌ Error sending automatic backup: %v\n", errSend)
			}
		}
	}
}
