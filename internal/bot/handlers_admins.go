package bot

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/sys"
	tele "gopkg.in/telebot.v3"
)

//go:embed monetag_miniapp.html
var monetagHTML []byte

func handleMenuAdmins(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Send("⛔ Only the Super Administrator (or Admin with Full Access) can use this function.", tele.ModeHTML)
	}

	data, _ := db.Load()
	accStatus := "🔓 Public"
	if !data.PublicAccess {
		accStatus = "🔒 Private"
	}

	markup := &tele.ReplyMarkup{}
	btnToggle := markup.Data("🔄 Acceso: "+accStatus, "toggle_public_access")
	btnList := markup.Data("📋 Listar Admins", "list_admins")
	btnAdd := markup.Data("➕ Agregar Admin", "add_admin")
	btnDel := markup.Data("➖ Quitar Admin", "del_admin_menu")
	btnRename := markup.Data("✏️ Renombrar Admin", "rename_admin_menu")
	btnInfo := markup.Data("📝 Editar Info Extra", "edit_extrainfo")
	btnCloudflare := markup.Data("☁️ Cloudflare Domain", "edit_cloudflare")
	btnCloudfront := markup.Data("🚀 Cloudfront Domain", "edit_cloudfront")
	btnBanner := markup.Data("📜 Banner SSH", "edit_banner")
	btnReset := markup.Data("🧹 Limpiar Historial", "reset_history")

	scanPubStatus := "🔓 ON"
	if !data.PublicScanner {
		scanPubStatus = "🔒 OFF"
	}
	btnScanToggle := markup.Data("🔍 Public Scanner: "+scanPubStatus, "toggle_public_scanner")

	monetStatus := "🔓 ON"
	if !data.Monetization {
		monetStatus = "🔒 OFF"
	}
	btnMonetToggle := markup.Data("💸 Monetization: "+monetStatus, "toggle_monetization")
	btnConfigAds := markup.Data("⚙️ Configurar MiniApp Ads", "menu_config_ads")

	btnReboot := markup.Data("🔄 Reiniciar VPS", "reboot_vps_confirm")
	btnAutoReboot := markup.Data("🕒 Auto Reboot", "menu_autoreboot")
	btnBackup := markup.Data("📥 Respaldar", "menu_backup")
	btnRestore := markup.Data("📤 Restaurar", "restore_req")
	btnBack := markup.Data("🔙 Back", "back_main")

	btnQuotas := markup.Data("📊 Creation Quotas", "edit_quotas")
	btnBans := markup.Data("🚫 Ban Management", "menu_bans")
	btnUpdater := markup.Data("🔄 Sistema Updater", "menu_updater")

	markup.Inline(
		markup.Row(btnToggle),
		markup.Row(btnList, btnAdd),
		markup.Row(btnDel, btnRename),
		markup.Row(btnInfo),
		markup.Row(btnCloudflare, btnCloudfront),
		markup.Row(btnBanner, btnQuotas),
		markup.Row(btnBans, btnScanToggle),
		markup.Row(btnMonetToggle, btnConfigAds),
		markup.Row(btnBackup, btnRestore),
		markup.Row(btnUpdater),
		markup.Row(btnReset),
		markup.Row(btnAutoReboot, btnReboot),
		markup.Row(btnBack),
	)

	texto := "⚙️ <b>PRO SETTINGS (ADMIN)</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("🛡️ <b>Acceso:</b> %s\n", accStatus)
	texto += fmt.Sprintf("🔍 <b>Public Scanner:</b> %s\n", scanPubStatus)
	texto += fmt.Sprintf("💸 <b>Monetization:</b> %s\n", monetStatus)
	texto += fmt.Sprintf("👤 <b>Admins:</b> %d\n", len(data.Admins)+1)
	texto += fmt.Sprintf("👥 <b>Historial:</b> %d IDs\n", len(data.UserHistory))
	texto += fmt.Sprintf("📊 <b>Public Quotas:</b> %d days / %d devices\n", data.GetMaxDaysPublic(), data.GetMaxLimitPublic())
	texto += fmt.Sprintf("📊 <b>Admin Quotas:</b> %d days / %d devices\n", data.GetMaxDaysAdmin(), data.GetMaxLimitAdmin())
	texto += fmt.Sprintf("💎 <b>Xray Public:</b> %d accounts max\n", data.GetMaxXrayPublic())
	texto += fmt.Sprintf("💎 <b>Xray Admin:</b> %d accounts max\n", data.GetMaxXrayAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>Select an advanced option:</i>"

	return SafeEditCtx(c, b, texto, markup)
}

func handleTogglePublicAccess(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.PublicAccess = !data.PublicAccess
		return nil
	})
	return handleMenuAdmins(c, b)
}

func handleToggleMonetization(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.Monetization = !data.Monetization
		return nil
	})
	return handleMenuAdmins(c, b)
}

func handleEditQuotas(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()

	markup := &tele.ReplyMarkup{}
	btnDaysPub := markup.Data(fmt.Sprintf("📅 Public Days: %d", data.GetMaxDaysPublic()), "quota_days_public")
	btnLimitPub := markup.Data(fmt.Sprintf("📱 Public Devices: %d", data.GetMaxLimitPublic()), "quota_limit_public")
	btnDaysAdm := markup.Data(fmt.Sprintf("📅 Admin Days: %d", data.GetMaxDaysAdmin()), "quota_days_admin")
	btnLimitAdm := markup.Data(fmt.Sprintf("📱 Admin Devices: %d", data.GetMaxLimitAdmin()), "quota_limit_admin")
	btnSSHPublic := markup.Data(fmt.Sprintf("👤 Max SSH Public: %d", data.GetMaxSSHPublic()), "quota_ssh_public")
	btnSSHAdmin := markup.Data(fmt.Sprintf("👤 Max SSH Admin: %d", data.GetMaxSSHAdmin()), "quota_ssh_admin")
	btnZivpnPublic := markup.Data(fmt.Sprintf("🛰️ Max ZiVPN Public: %d", data.GetMaxZivpnPublic()), "quota_zivpn_public")
	btnZivpnAdmin := markup.Data(fmt.Sprintf("🛰️ Max ZiVPN Admin: %d", data.GetMaxZivpnAdmin()), "quota_zivpn_admin")
	btnXrayPub := markup.Data(fmt.Sprintf("💎 Xray Public: %d", data.GetMaxXrayPublic()), "quota_xray_public")
	btnXrayAdm := markup.Data(fmt.Sprintf("💎 Xray Admin: %d", data.GetMaxXrayAdmin()), "quota_xray_admin")
	btnBack := markup.Data("🔙 Back", "menu_admins")

	markup.Inline(
		markup.Row(btnDaysPub, btnLimitPub),
		markup.Row(btnDaysAdm, btnLimitAdm),
		markup.Row(btnSSHPublic, btnSSHAdmin),
		markup.Row(btnZivpnPublic, btnZivpnAdmin),
		markup.Row(btnXrayPub, btnXrayAdm),
		markup.Row(btnBack),
	)

	texto := "📊 <b>User Creation Quotas</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("👥 <b>Public SSH (Params):</b> %d days / %d devices\n", data.GetMaxDaysPublic(), data.GetMaxLimitPublic())
	texto += fmt.Sprintf("👤 <b>Admin SSH (Params):</b> %d days / %d devices\n", data.GetMaxDaysAdmin(), data.GetMaxLimitAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("👤 <b>Max SSH Accounts Public:</b> max %d\n", data.GetMaxSSHPublic())
	texto += fmt.Sprintf("👤 <b>Max SSH Accounts Admin:</b> max %d\n", data.GetMaxSSHAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("🛰️ <b>Max ZiVPN Accounts Public:</b> max %d\n", data.GetMaxZivpnPublic())
	texto += fmt.Sprintf("🛰️ <b>Max ZiVPN Accounts Admin:</b> max %d\n", data.GetMaxZivpnAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("💎 <b>Xray Public:</b> max %d accounts\n", data.GetMaxXrayPublic())
	texto += fmt.Sprintf("💎 <b>Xray Admin:</b> max %d accounts\n", data.GetMaxXrayAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>These values apply when creating SSH, ZiVPN and Xray users.\nThe SuperAdmin has no limits.</i>"

	return SafeEditCtx(c, b, texto, markup)
}

func handleQuotaPrompt(c tele.Context, b *tele.Bot, step string, label string) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, step)
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_quotas")))
	return SafeEditCtx(c, b, fmt.Sprintf("✏️ <b>%s</b>\n\n<i>Enter the new value (number):</i>", label), markup)
}

func handleListAdmins(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	res := "📋 <b>LISTADO DE ADMINISTRADORES</b>\n\n"
	res += fmt.Sprintf("⭐ <b>SuperAdmin (Root):</b> <code>%s</code>\n", superAdmin)

	if len(data.Admins) == 0 {
		res += "\n<i>No additional administrators.</i>"
	} else {
		i := 1
		for id, info := range data.Admins {
			expireText := "Unlimited"
			if info.Expire != "" {
				expireText = info.Expire
			}
			res += fmt.Sprintf("\n%d. 👤 <b>%s</b>\n   └ ID: <code>%s</code>\n   └ Vence: <code>%s</code>\n", i, info.Alias, id, expireText)
			i++
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_admins")))
	return SafeEditCtx(c, b, res, markup)
}

func handleAddAdminPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_admin_id")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	return SafeEditCtx(c, b, "➕ <b>Add New Administrator</b>\n\n📝 <b>Step 1/2:</b> Enter the <b>numeric ID</b> of the Telegram user:\n\nExample: <code>123456789</code>", markup)
}

func handleDelAdminMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	if len(data.Admins) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "No administrators to remove.", ShowAlert: true})
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for id, info := range data.Admins {
		rows = append(rows, markup.Row(markup.Data("❌ "+info.Alias+" ("+id+")", "del_adm_exec", id)))
	}
	rows = append(rows, markup.Row(markup.Data("🔙 Back", "menu_admins")))
	markup.Inline(rows...)

	return SafeEditCtx(c, b, "➖ <b>Remove Administrator</b>\n\nSelect who you want to remove permissions from:", markup)
}

func handleDelAdminExec(c tele.Context, b *tele.Bot) error {
	id := c.Data()

	// Buscar alias antes de borrar
	data, _ := db.Load()
	alias := "Admin"
	if info, ok := data.Admins[id]; ok {
		alias = info.Alias
	}

	db.Update(func(data *db.ConfigData) error {
		delete(data.Admins, id)
		return nil
	})

	// Respond to the callback to unlock the button
	c.Respond(&tele.CallbackResponse{Text: "✅ Admin removed"})

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back to Settings", "menu_admins")))
	return SafeEditCtx(c, b, fmt.Sprintf("✅ <b>Admin Removed</b>\n\n👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n\n<i>They no longer have administrator permissions.</i>", alias, id), markup)
}

func handleRenameAdminMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	if len(data.Admins) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "No administrators to rename.", ShowAlert: true})
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for id, info := range data.Admins {
		rows = append(rows, markup.Row(markup.Data("✏️ "+info.Alias+" ("+id+")", "rename_adm_sel", id)))
	}
	rows = append(rows, markup.Row(markup.Data("🔙 Back", "menu_admins")))
	markup.Inline(rows...)

	return SafeEditCtx(c, b, "✏️ <b>Rename Administrator</b>\n\nSelect the admin you want to rename:", markup)
}

func handleRenameAdminSelect(c tele.Context, b *tele.Bot) error {
	id := c.Data()
	chatID := c.Chat().ID

	data, _ := db.Load()
	info, exists := data.Admins[id]
	if !exists {
		return c.Respond(&tele.CallbackResponse{Text: "Admin no encontrado.", ShowAlert: true})
	}

	SetTempValue(chatID, "rename_admin_id", id)
	SetUserStep(chatID, "awaiting_rename_admin_alias")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	return SafeEditCtx(c, b, fmt.Sprintf("✏️ <b>Rename Admin</b>\n\n👤 <b>Current:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n\nEnter the <b>new alias</b>:", info.Alias, id), markup)
}

func handleEditExtraInfoPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_extrainfo")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	return SafeEditCtx(c, b, "📝 <b>Edit Extra Information</b>\n\nThis information appears in the /info menu.\n\n✏️ <i>Enter the new text (HTML supported):</i>", markup)
}

func handleEditCloudflarePrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_cloudflare")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))
	return SafeEditCtx(c, b, "☁️ <b>Configure Cloudflare Domain</b>\n\n✏️ <i>Enter the domain:</i>\n\nExample: <code>my.host.com</code>", markup)
}

func handleEditCloudfrontPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_cloudfront")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))
	return SafeEditCtx(c, b, "🚀 <b>Configure Cloudfront Domain</b>\n\n✏️ <i>Enter the domain:</i>\n\nExample: <code>xyz123.cloudfront.net</code>", markup)
}

// Default banner for ORX TUNNEL
const defaultBanner = `<html>
<h5 style="text-align:center;">
<font face="monospace" color="#00ff00">
⠀ORX TUNNEL⠀⠀
</font>
</h5>
<h1 style="text-align:center;">
<font face="monospace" color="#00ff00"><b>ORX TUNNEL</b></font>
</h1>
<h5 style="text-align:center;">
<font color='#29b6f6'>==============================</font>
<font color='#29b6f6'><b>✈ TELEGRAM ✈</b></font>
<font color='#29b6f6'>==============================</font>
</h5>
<h5 style="text-align:center;">
<font color='#ffffff'>Dev: </font><a href="https://t.me/orxtunnel"><font color='#f1c40f'>@orxtunnel</font></a>
<font color='#ffffff'>Channel: </font><a href="https://t.me/orxtunnel"><font color='#f1c40f'>@orxtunnel</font></a>
</h5>
<h4 style="text-align:center;">
<font color='#FF00FF'><b>🔥 PREMIUM SERVERS AVAILABLE - 30 DAYS 🔥</b></font>
</h4>
<h5 style="text-align:center;">
<font color='#ff0000'>==============================</font>
<font color='#ff0000'><b>⚡ FREE SERVERS ⚡</b></font>
<font color='#ff0000'>==============================</font>
</h5>
<h6 style="text-align:center;">
<font color='#ff9800'><b>⚠️ SERVER RULES ⚠️</b></font>
<font color='#ffffff'>🚫 NO Torrent / P2P</font>
<font color='#ffffff'>🚫 NO Spam / Fraud</font>
<font color='#ffffff'>🚫 NO DDoS Attacks</font>
<font color='#ff5252'><i>Violation results in an automatic ban</i></font>
</h6>
<h5 style="text-align:center;">
<font color='#00e676'><b>CREATED IN : @orxtunnel_bot</b></font>
</h5>
</html>`

func handleEditBannerPrompt(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()

	status := "👤 Individual Banners (Active)"
	bannerType := ""
	if data.SSHBanner != "" {
		status = "🌐 Global Banner (Active)"
		bannerType = "\n\n⚠️ <i>The individual system is disabled. All accounts will use the same global banner.</i>"
	} else {
		bannerType = "\n\n✅ <i>Each user has their own banner with days and limits.</i>"
	}

	markup := &tele.ReplyMarkup{}
	btnPromo := markup.Data("📝 Editar Textos Promo", "edit_promo_menu")
	btnCustom := markup.Data("🌐 Activar Banner Global", "banner_set_custom")
	btnDeactivate := markup.Data("🚫 Desactivar Global (Usar Individual)", "banner_deactivate")
	btnBack := markup.Data("🔙 Back", "menu_admins")

	markup.Inline(
		markup.Row(btnPromo),
		markup.Row(btnCustom),
		markup.Row(btnDeactivate),
		markup.Row(btnBack),
	)

	texto := fmt.Sprintf("📜 <b>SSH Banner Management</b>\n\n📊 <b>Current Mode:</b> %s%s\n\nWhat would you like to do?", status, bannerType)
	return SafeEditCtx(c, b, texto, markup)
}

func handleEditPromoMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()

	promoText := "🔥 PREMIUM SERVERS AVAILABLE! 🔥"
	if data.BannerPromoText != "" {
		promoText = data.BannerPromoText
	}

	promoChannel := "@orxtunnel"
	if data.BannerPromoChannel != "" {
		promoChannel = data.BannerPromoChannel
	}

	promoSupport := "@orxtunnel"
	if data.BannerPromoSupport != "" {
		promoSupport = data.BannerPromoSupport
	}

	promoBotName := "@orxtunnel_bot"
	if data.BannerPromoBotName != "" {
		promoBotName = data.BannerPromoBotName
	}

	markup := &tele.ReplyMarkup{}
	btnText := markup.Data("📝 Editar Mensaje", "edit_promo_text")
	btnChannel := markup.Data("📢 Editar Canal", "edit_promo_channel")
	btnSupport := markup.Data("👤 Editar Soporte", "edit_promo_support")
	btnBotName := markup.Data("🤖 Editar Nombre Bot", "edit_promo_botname")
	btnBack := markup.Data("🔙 Back", "edit_banner")

	markup.Inline(
		markup.Row(btnText, btnChannel),
		markup.Row(btnSupport, btnBotName),
		markup.Row(btnBack),
	)

	texto := "📝 <b>Editar Textos Promocionales (Banners Individuales)</b>\n\n"
	texto += "These texts appear at the bottom of every user's banner.\n\n"
	texto += fmt.Sprintf("💬 <b>Mensaje Promo:</b>\n<code>%s</code>\n\n", promoText)
	texto += fmt.Sprintf("📢 <b>Channel:</b>\n<code>%s</code>\n\n", promoChannel)
	texto += fmt.Sprintf("👤 <b>Support:</b>\n<code>%s</code>\n\n", promoSupport)
	texto += fmt.Sprintf("🤖 <b>Creado En:</b>\n✅ CREADO EN : <code>%s</code>", promoBotName)

	return SafeEditCtx(c, b, texto, markup)
}

func handleBannerSetCustom(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_ssh_banner")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_banner")))
	return SafeEditCtx(c, b, "📜 <b>Custom SSH Banner</b>\n\n✏️ <i>Enter the banner text (basic HTML supported):</i>\n\nThis is shown when connecting over SSH.", markup)
}

func handleEditPromoText(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_text")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "💬 <b>Edit Promo Message</b>\n\n✏️ <i>Enter the new promotional text (e.g. 🔥 SERVER DEALS FROM $5! 🔥):</i>", markup)
}

func handleEditPromoChannel(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_channel")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "📢 <b>Edit Promo Channel</b>\n\n✏️ <i>Enter your channel @username (e.g. @MyVIPChannel):</i>", markup)
}

func handleEditPromoSupport(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_support")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "👤 <b>Edit Promo Support</b>\n\n✏️ <i>Enter your Telegram @username for support (e.g. @YourUsername):</i>", markup)
}

func handleEditPromoBotName(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_botname")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "🤖 <b>Edit Bot Name</b>\n\n✏️ <i>Enter your bot @username (e.g. @MySuperVPN_bot):</i>\n\nThe banner keeps the prefix \"✅ CREATED IN : \".", markup)
}

func handleBannerDeactivate(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.SSHBanner = ""
		return nil
	})

	// Quitar banner global del sistema
	exec.Command("sh", "-c", "rm -f /etc/sshd_banner").Run()
	exec.Command("sed", "-i", "/^Banner/d", "/etc/ssh/sshd_config").Run()

	// Restaurar banners individuales (Match User)
	go sys.RefreshAllBanners()

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "edit_banner")))
	return SafeEditCtx(c, b, "✅ <b>Global Banner disabled.</b>\n\n<i>The individual banner system has been restored.</i>", markup)
}

func handleResetHistoryConfirm(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	btnYes := markup.Data("✅ Yes, Clear", "reset_history_exec")
	btnNo := markup.Data("❌ No, Cancelar", "menu_admins")
	markup.Inline(markup.Row(btnYes, btnNo))

	return SafeEditCtx(c, b, "⚠️ <b>Are you sure you want to clear the history?</b>\n\nAll registered user IDs will be deleted (broadcasts will no longer reach them until they restart the bot).", markup)
}

func handleResetHistoryExec(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.UserHistory = []int64{}
		return nil
	})
	return c.Respond(&tele.CallbackResponse{Text: "Historial de IDs reseteado.", ShowAlert: true})
}

func handleServerRebootConfirm(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	btnYes := markup.Data("🔄 Reiniciar AHORA", "reboot_vps_exec")
	btnNo := markup.Data("🔙 Cancelar", "menu_admins")
	markup.Inline(markup.Row(btnYes, btnNo))

	return SafeEditCtx(c, b, "🚨 <b>WARNING: SERVER REBOOT</b>\n\nAre you sure you want to reboot the VPS? All current connections will be dropped.", markup)
}

func handleServerRebootExec(c tele.Context, b *tele.Bot) error {
	c.Edit("⏳ <b>Rebooting VPS...</b> the bot will be offline for a few minutes.", tele.ModeHTML)
	exec.Command("reboot").Run()
	return nil
}

// === SISTEMA DE ACTUALIZACIONES (UPDATER) ===

func handleMenuUpdater(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return c.Send("⛔ Only administrators can use this function.")
	}

	data, _ := db.Load()
	autoStatus := "🔴 Desactivada"
	if data.AutoUpdate {
		autoStatus = "🟢 Activada"
	}

	text := "🔄 <b>Sistema de Actualizaciones (GitHub)</b>\n\n"
	text += "Current Version: <b>" + sys.CurrentVersion + "</b>\n"
	text += "Auto-Update: <b>" + autoStatus + "</b>\n\n"
	text += "You can check for a new version or enable automatic updates (the bot checks every 12 hours)."

	markup := &tele.ReplyMarkup{}
	btnCheck := markup.Data("🔍 Check for Update", "updater_check")
	btnAuto := markup.Data("⚙️ Auto-Update: "+autoStatus, "updater_toggle_auto")
	btnForce := markup.Data("⚠️ Force Reinstall (Dev)", "updater_run")
	btnBack := markup.Data("🔙 Back to Settings", "menu_admins")

	markup.Inline(
		markup.Row(btnCheck),
		markup.Row(btnAuto),
		markup.Row(btnForce),
		markup.Row(btnBack),
	)

	return SafeEditCtx(c, b, text, markup)
}

func handleUpdaterToggleAuto(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return nil
	}

	db.Update(func(d *db.ConfigData) error {
		d.AutoUpdate = !d.AutoUpdate
		return nil
	})

	return handleMenuUpdater(c, b)
}

func handleUpdaterCheck(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return nil
	}

	hasUpdate, newVer, err := sys.CheckForUpdate()

	markup := &tele.ReplyMarkup{}
	btnBack := markup.Data("🔙 Back", "menu_updater")

	if err != nil {
		markup.Inline(markup.Row(btnBack))
		return SafeEditCtx(c, b, "❌ <b>Error checking for updates:</b>\n"+err.Error(), markup)
	}

	if !hasUpdate {
		btnForceNow := markup.Data("⚠️ Force Reinstall", "updater_run")
		markup.Inline(
			markup.Row(btnForceNow),
			markup.Row(btnBack),
		)
		return SafeEditCtx(c, b, "✅ <b>You are on the latest version.</b>\nCurrent version: "+sys.CurrentVersion+"\nRemote version: "+newVer, markup)
	}

	btnUpdateNow := markup.Data("⚡ Actualizar a v"+newVer, "updater_run")
	markup.Inline(
		markup.Row(btnUpdateNow),
		markup.Row(btnBack),
	)

	return SafeEditCtx(c, b, "🎉 <b>New update found!</b>\n\nCurrent version: "+sys.CurrentVersion+"\nNew version: <b>"+newVer+"</b>\n\nDo you want to update the bot now? The service will restart for about 15 seconds.", markup)
}

func handleUpdaterRun(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return nil
	}

	c.Send("⚡ <b>Starting update...</b>\nDownloading and compiling from GitHub. The bot will not respond for about 15 seconds.", tele.ModeHTML)
	
	err := sys.RunUpdate()
	if err != nil {
		return c.Send("❌ Error starting the updater: " + err.Error())
	}
	return nil
}

func handleTogglePublicScanner(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.PublicScanner = !data.PublicScanner
		return nil
	})
	return handleMenuAdmins(c, b)
}

func handleAutoRebootMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desactivado"
	if data.AutoReboot {
		status = "✅ Activado"
	}

	markup := &tele.ReplyMarkup{}
	btnToggle := markup.Data("🔄 Switch: "+status, "toggle_autoreboot")
	btnBack := markup.Data("🔙 Back", "menu_admins")

	markup.Inline(
		markup.Row(btnToggle),
		markup.Row(btnBack),
	)

	texto := "🕒 <b>AUTO-REBOOT CONFIGURATION</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>The server reboots automatically once it reaches 24 hours of continuous uptime.</i>\n\n"
	texto += fmt.Sprintf("📊 <b>Status:</b> %s\n", status)
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>Select an option:</i>"

	return SafeEditCtx(c, b, texto, markup)
}

func handleToggleAutoReboot(c tele.Context, b *tele.Bot) error {
	db.Update(func(data *db.ConfigData) error {
		data.AutoReboot = !data.AutoReboot
		return nil
	})
	return handleAutoRebootMenu(c, b)
}

func handleMenuBans(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	markup := &tele.ReplyMarkup{}
	
	btnBanUser := markup.Data("➕ Ban User", "ban_user_prompt")
	btnBack := markup.Data("🔙 Back", "menu_admins")
	
	var rows []tele.Row
	rows = append(rows, markup.Row(btnBanUser))
	
	texto := "🚫 <b>BANNED USERS MANAGEMENT</b>\n━━━━━━━━━━━━━━\n"
	if len(data.BannedUsers) == 0 {
		texto += "<i>No banned users.</i>\n\n"
	} else {
		texto += "<i>Select a user to unban:</i>\n\n"
		for id, info := range data.BannedUsers {
			rows = append(rows, markup.Row(markup.Data(fmt.Sprintf("✅ Desbanear a %s", info.Name), "unban_user", id)))
			texto += fmt.Sprintf("👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n📝 Reason: <i>%s</i>\n📅 Date: %s\n\n", info.Name, id, info.Reason, info.Date)
		}
	}
	
	rows = append(rows, markup.Row(btnBack))
	markup.Inline(rows...)
	
	return SafeEditCtx(c, b, texto, markup)
}

func handleBanUserPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_ban_id")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_bans")))
	return SafeEditCtx(c, b, "➕ <b>Ban User</b>\n\n📝 <b>Step 1/3:</b> Enter the <b>numeric ID</b> of the Telegram user you want to ban:", markup)
}

func handleUnbanUser(c tele.Context, b *tele.Bot) error {
	id := c.Data()
	db.Update(func(data *db.ConfigData) error {
		delete(data.BannedUsers, id)
		return nil
	})
	c.Respond(&tele.CallbackResponse{Text: "✅ User unbanned", ShowAlert: true})
	return handleMenuBans(c, b)
}

func handleAdminAccessType(c tele.Context, b *tele.Bot, isFullAccess bool) error {
	chatID := c.Chat().ID
	id := GetTempValue(chatID, "admin_id")
	alias := GetTempValue(chatID, "admin_alias")
	daysStr := GetTempValue(chatID, "admin_days")

	if id == "" || daysStr == "" {
		return c.Respond(&tele.CallbackResponse{Text: "❌ Session expired or incomplete data.", ShowAlert: true})
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "❌ Error reading days.", ShowAlert: true})
	}

	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	db.Update(func(data *db.ConfigData) error {
		data.Admins[id] = db.AdminInfo{Alias: alias, Expire: expireDate, FullAccess: isFullAccess}
		return nil
	})

	tipoAcceso := "Normal (Limited)"
	if isFullAccess {
		tipoAcceso = "Full (SuperAdmin)"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back to Admins", "menu_admins")))
	
	msg := fmt.Sprintf("✅ <b>Administrator Registered</b>\n\n👤 <b>Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n📅 <b>Expires:</b> %s (%d days)\n👑 <b>Access Type:</b> %s", alias, id, expireDate, days, tipoAcceso)
	SafeEditCtx(c, b, msg, markup)

	// Send Welcome Message to the new admin
	newAdminID, errParse := strconv.ParseInt(id, 10, 64)
	if errParse == nil {
		welcomeMsg := "🎉 <b>Congratulations! You have been promoted to Administrator.</b>\n\n" +
			"Welcome to your new control panel. You now have access to advanced tools to manage users and monitor the service.\n\n" +
			"📅 <b>Your subscription is valid until:</b> <code>" + expireDate + "</code>\n" +
			"👑 <b>Tipo de Acceso:</b> " + tipoAcceso + "\n\n" +
			"<i>Type /start or use the menu to get started.</i>"
		b.Send(&tele.User{ID: newAdminID}, welcomeMsg, tele.ModeHTML)
	}

	DeleteUserStep(chatID)

	return c.Respond(&tele.CallbackResponse{Text: "✅ Admin created successfully."})
}

func handleMenuAdsConfig(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_ads_config_script")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	texto := "⚙️ <b>MONETIZATION CONFIGURATION</b>\n\n" +
		"Step 1: Create your app in Monetag, generate a zone and copy the <b>&lt;script&gt;</b> tag they provide (the one starting with <code>&lt;script src='//libtl.com...</code>).\n\n" +
		"Please paste it here and send it in a message."

	return SafeEditCtx(c, b, texto, markup)
}

func processAdsConfigSteps(step, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))

	switch step {
	case "awaiting_ads_config_script":
		if !strings.Contains(text, "<script") || !strings.Contains(text, "</script>") {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ The text sent does not look like a valid &lt;script&gt; tag. Try again:", markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}
		
		temp := GetTempData(chatID)
		temp["ads_script"] = text
		SetTempData(chatID, temp)
		SetUserStep(chatID, "awaiting_ads_config_rewarded")

		texto := "✅ <b>Script guardado.</b>\n\n" +
			"Step 2: Now copy the activation code for the <b>Rewarded Interstitial</b> format (the block containing <code>show_XXXXXXX().then(...)</code>).\n\n" +
			"Please paste it here and send it to me."
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ads_config_rewarded":
		re := regexp.MustCompile(`show_\d+`)
		match := re.FindString(text)
		if match == "" {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ No <code>show_XXXXXXX</code> function was found in the code you sent. Try again:", markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}

		temp := GetTempData(chatID)
		scriptTag := temp["ads_script"]
		functionName := match

		// Utilizar la plantilla HTML embebida
		htmlStr := string(monetagHTML)
		
		// Reemplazos de Plantilla
		// Asumiendo que la plantilla original tiene: <script src='//libtl.com/sdk.js' data-zone='11209533' data-sdk='show_11209533'></script>
		// Y la función original es show_11209533
		htmlStr = regexp.MustCompile(`(?i)<script src='//libtl\.com[^>]+></script>`).ReplaceAllString(htmlStr, scriptTag)
		htmlStr = strings.ReplaceAll(htmlStr, "show_11209533", functionName)
		htmlStr = strings.ReplaceAll(htmlStr, "ORX_TUNNEL_BOT", b.Me.Username)

		zipName := fmt.Sprintf("miniapp_monetag_%d.zip", chatID)
		zipFile, err := os.Create(zipName)
		if err == nil {
			archive := zip.NewWriter(zipFile)
			f, _ := archive.Create("monetag_miniapp.html")
			f.Write([]byte(htmlStr))
			archive.Close()
			zipFile.Close()

			doc := &tele.Document{File: tele.FromDisk(zipName), FileName: "monetag_miniapp.zip"}
			b.Send(c.Chat(), doc)
			os.Remove(zipName)
		}

		SetUserStep(chatID, "awaiting_ads_config_url")
		texto := "✅ <b>Archivos generados.</b>\n\n" +
			"Step 3: Download the ZIP file, extract <code>monetag_miniapp.html</code> and upload it to Vercel, GitHub Pages or another host.\n\n" +
			"Once it is online, send me the <b>full URL</b> (e.g. <code>https://my-app.vercel.app/monetag_miniapp.html</code>)."
		
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ads_config_url":
		if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ The URL must start with http:// or https://. Try again:", markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}

		db.Update(func(data *db.ConfigData) error {
			data.WebAppURL = strings.TrimSpace(text)
			data.Monetization = true
			return nil
		})

		DeleteUserStep(chatID)
		texto := "🎉 <b>Monetization configured successfully!</b>\n\n" +
			"The system will now redirect public users to your MiniApp before creating or renewing accounts.\n\n" +
			"You can disable it at any time from Pro Settings."
		
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back to Pro Settings", "menu_admins")))
		
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupBack)
		SetLastBotMsg(chatID, msg)
		return nil
	}

	return nil
}
