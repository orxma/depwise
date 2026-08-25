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
		return c.Send("⛔ Solo el Super Administrador (o Admin con Acceso Total) puede usar esta función.", tele.ModeHTML)
	}

	data, _ := db.Load()
	accStatus := "🔓 Público"
	if !data.PublicAccess {
		accStatus = "🔒 Privado"
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
	btnScanToggle := markup.Data("🔍 Escaner Público: "+scanPubStatus, "toggle_public_scanner")

	monetStatus := "🔓 ON"
	if !data.Monetization {
		monetStatus = "🔒 OFF"
	}
	btnMonetToggle := markup.Data("💸 Monetización: "+monetStatus, "toggle_monetization")
	btnConfigAds := markup.Data("⚙️ Configurar MiniApp Ads", "menu_config_ads")

	btnReboot := markup.Data("🔄 Reiniciar VPS", "reboot_vps_confirm")
	btnAutoReboot := markup.Data("🕒 Auto Reboot", "menu_autoreboot")
	btnBackup := markup.Data("📥 Respaldar", "menu_backup")
	btnRestore := markup.Data("📤 Restaurar", "restore_req")
	btnBack := markup.Data("🔙 Volver", "back_main")

	btnQuotas := markup.Data("📊 Cuotas Creación", "edit_quotas")
	btnBans := markup.Data("🚫 Gestión Bans", "menu_bans")
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

	texto := "⚙️ <b>CONFIGURACIÓN PRO (ADMIN)</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("🛡️ <b>Acceso:</b> %s\n", accStatus)
	texto += fmt.Sprintf("🔍 <b>Escaner Público:</b> %s\n", scanPubStatus)
	texto += fmt.Sprintf("💸 <b>Monetización:</b> %s\n", monetStatus)
	texto += fmt.Sprintf("👤 <b>Admins:</b> %d\n", len(data.Admins)+1)
	texto += fmt.Sprintf("👥 <b>Historial:</b> %d IDs\n", len(data.UserHistory))
	texto += fmt.Sprintf("📊 <b>Cuotas Público:</b> %d días / %d disp.\n", data.GetMaxDaysPublic(), data.GetMaxLimitPublic())
	texto += fmt.Sprintf("📊 <b>Cuotas Admin:</b> %d días / %d disp.\n", data.GetMaxDaysAdmin(), data.GetMaxLimitAdmin())
	texto += fmt.Sprintf("💎 <b>VMess Público:</b> %d cuentas max\n", data.GetMaxXrayPublic())
	texto += fmt.Sprintf("💎 <b>VMess Admin:</b> %d cuentas max\n", data.GetMaxXrayAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>Selecciona una opción avanzada:</i>"

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
	btnDaysPub := markup.Data(fmt.Sprintf("📅 Días Público: %d", data.GetMaxDaysPublic()), "quota_days_public")
	btnLimitPub := markup.Data(fmt.Sprintf("📱 Disp. Público: %d", data.GetMaxLimitPublic()), "quota_limit_public")
	btnDaysAdm := markup.Data(fmt.Sprintf("📅 Días Admin: %d", data.GetMaxDaysAdmin()), "quota_days_admin")
	btnLimitAdm := markup.Data(fmt.Sprintf("📱 Disp. Admin: %d", data.GetMaxLimitAdmin()), "quota_limit_admin")
	btnSSHPublic := markup.Data(fmt.Sprintf("👤 Max SSH Público: %d", data.GetMaxSSHPublic()), "quota_ssh_public")
	btnSSHAdmin := markup.Data(fmt.Sprintf("👤 Max SSH Admin: %d", data.GetMaxSSHAdmin()), "quota_ssh_admin")
	btnZivpnPublic := markup.Data(fmt.Sprintf("🛰️ Max ZiVPN Público: %d", data.GetMaxZivpnPublic()), "quota_zivpn_public")
	btnZivpnAdmin := markup.Data(fmt.Sprintf("🛰️ Max ZiVPN Admin: %d", data.GetMaxZivpnAdmin()), "quota_zivpn_admin")
	btnXrayPub := markup.Data(fmt.Sprintf("💎 VMess Público: %d", data.GetMaxXrayPublic()), "quota_xray_public")
	btnXrayAdm := markup.Data(fmt.Sprintf("💎 VMess Admin: %d", data.GetMaxXrayAdmin()), "quota_xray_admin")
	btnBack := markup.Data("🔙 Volver", "menu_admins")

	markup.Inline(
		markup.Row(btnDaysPub, btnLimitPub),
		markup.Row(btnDaysAdm, btnLimitAdm),
		markup.Row(btnSSHPublic, btnSSHAdmin),
		markup.Row(btnZivpnPublic, btnZivpnAdmin),
		markup.Row(btnXrayPub, btnXrayAdm),
		markup.Row(btnBack),
	)

	texto := "📊 <b>Cuotas de Creación de Usuarios</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("👥 <b>Público SSH (Params):</b> %d días / %d dispositivos\n", data.GetMaxDaysPublic(), data.GetMaxLimitPublic())
	texto += fmt.Sprintf("👤 <b>Admin SSH (Params):</b> %d días / %d dispositivos\n", data.GetMaxDaysAdmin(), data.GetMaxLimitAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("👤 <b>Límite Cuentas SSH Público:</b> máx %d\n", data.GetMaxSSHPublic())
	texto += fmt.Sprintf("👤 <b>Límite Cuentas SSH Admin:</b> máx %d\n", data.GetMaxSSHAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("🛰️ <b>Límite Cuentas ZiVPN Público:</b> máx %d\n", data.GetMaxZivpnPublic())
	texto += fmt.Sprintf("🛰️ <b>Límite Cuentas ZiVPN Admin:</b> máx %d\n", data.GetMaxZivpnAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += fmt.Sprintf("💎 <b>VMess Público:</b> máx %d cuentas\n", data.GetMaxXrayPublic())
	texto += fmt.Sprintf("💎 <b>VMess Admin:</b> máx %d cuentas\n", data.GetMaxXrayAdmin())
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>Estos valores se aplican al crear usuarios SSH, ZiVPN y VMess.\nEl SuperAdmin no tiene límites.</i>"

	return SafeEditCtx(c, b, texto, markup)
}

func handleQuotaPrompt(c tele.Context, b *tele.Bot, step string, label string) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, step)
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_quotas")))
	return SafeEditCtx(c, b, fmt.Sprintf("✏️ <b>%s</b>\n\n<i>Escribe el nuevo valor (número):</i>", label), markup)
}

func handleListAdmins(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	res := "📋 <b>LISTADO DE ADMINISTRADORES</b>\n\n"
	res += fmt.Sprintf("⭐ <b>SuperAdmin (Root):</b> <code>%s</code>\n", superAdmin)

	if len(data.Admins) == 0 {
		res += "\n<i>No hay administradores adicionales.</i>"
	} else {
		i := 1
		for id, info := range data.Admins {
			expireText := "Ilimitado"
			if info.Expire != "" {
				expireText = info.Expire
			}
			res += fmt.Sprintf("\n%d. 👤 <b>%s</b>\n   └ ID: <code>%s</code>\n   └ Vence: <code>%s</code>\n", i, info.Alias, id, expireText)
			i++
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_admins")))
	return SafeEditCtx(c, b, res, markup)
}

func handleAddAdminPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_admin_id")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	return SafeEditCtx(c, b, "➕ <b>Agregar Nuevo Administrador</b>\n\n📝 <b>Paso 1/2:</b> Escribe el <b>ID numérico</b> del usuario de Telegram:\n\nEjemplo: <code>123456789</code>", markup)
}

func handleDelAdminMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	if len(data.Admins) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "No hay administradores para quitar.", ShowAlert: true})
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for id, info := range data.Admins {
		rows = append(rows, markup.Row(markup.Data("❌ "+info.Alias+" ("+id+")", "del_adm_exec", id)))
	}
	rows = append(rows, markup.Row(markup.Data("🔙 Volver", "menu_admins")))
	markup.Inline(rows...)

	return SafeEditCtx(c, b, "➖ <b>Quitar Administrador</b>\n\nSelecciona a quién deseas retirar los permisos:", markup)
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

	// Responder al callback para desbloquear el botón
	c.Respond(&tele.CallbackResponse{Text: "✅ Admin eliminado"})

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver a Ajustes", "menu_admins")))
	return SafeEditCtx(c, b, fmt.Sprintf("✅ <b>Admin Eliminado</b>\n\n👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n\n<i>Ya no tiene permisos de administrador.</i>", alias, id), markup)
}

func handleRenameAdminMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	if len(data.Admins) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "No hay administradores para renombrar.", ShowAlert: true})
	}

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for id, info := range data.Admins {
		rows = append(rows, markup.Row(markup.Data("✏️ "+info.Alias+" ("+id+")", "rename_adm_sel", id)))
	}
	rows = append(rows, markup.Row(markup.Data("🔙 Volver", "menu_admins")))
	markup.Inline(rows...)

	return SafeEditCtx(c, b, "✏️ <b>Renombrar Administrador</b>\n\nSelecciona al admin que deseas renombrar:", markup)
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

	return SafeEditCtx(c, b, fmt.Sprintf("✏️ <b>Renombrar Admin</b>\n\n👤 <b>Actual:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n\nEscribe el <b>nuevo alias</b>:", info.Alias, id), markup)
}

func handleEditExtraInfoPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_extrainfo")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	return SafeEditCtx(c, b, "📝 <b>Editar Información Extra</b>\n\nEsta información aparecerá en el menú /info.\n\n✏️ <i>Escribe el nuevo texto (soporta HTML):</i>", markup)
}

func handleEditCloudflarePrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_cloudflare")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))
	return SafeEditCtx(c, b, "☁️ <b>Configurar Dominio Cloudflare</b>\n\n✏️ <i>Escribe el dominio :</i>\n\nEjemplo: <code>mi.host.com</code>", markup)
}

func handleEditCloudfrontPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_cloudfront")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))
	return SafeEditCtx(c, b, "🚀 <b>Configurar Dominio Cloudfront</b>\n\n✏️ <i>Escribe el dominio:</i>\n\nEjemplo: <code>xyz123.cloudfront.net</code>", markup)
}

// Banner predeterminado de ORX TUNNEL
const defaultBanner = `<html>
<h5 style="text-align:center;">
<font face="monospace" color="#00ff00">
⠀KEVIN TECH TUTORIALS⠀⠀
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
<font color='#ffffff'>Dev: </font><a href="https://t.me/Dan3651"><font color='#f1c40f'>@Dan3651</font></a>
<font color='#ffffff'>Canal: </font><a href="https://t.me/orxtunnel"><font color='#f1c40f'>@orxtunnel</font></a>
</h5>
<h4 style="text-align:center;">
<font color='#FF00FF'><b>🔥 ¡SE VENDEN SERVIDORES PREMIUM 30 DÍAS A 15 SOLES! 🔥</b></font>
</h4>
<h5 style="text-align:center;">
<font color='#ff0000'>==============================</font>
<font color='#ff0000'><b>⚡ SERVIDORES FREE ⚡</b></font>
<font color='#ff0000'>==============================</font>
</h5>
<h6 style="text-align:center;">
<font color='#ff9800'><b>⚠️ REGLAS DEL SERVIDOR ⚠️</b></font>
<font color='#ffffff'>🚫 NO Torrent / P2P</font>
<font color='#ffffff'>🚫 NO Spam / Fraude</font>
<font color='#ffffff'>🚫 NO Ataques DDoS</font>
<font color='#ff5252'><i>El incumplimiento genera ban automático</i></font>
</h6>
<h5 style="text-align:center;">
<font color='#00e676'><b>CREADO EN : @orxtunnel_bot</b></font>
</h5>
</html>`

func handleEditBannerPrompt(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()

	status := "👤 Banners Individuales (Activo)"
	bannerType := ""
	if data.SSHBanner != "" {
		status = "🌐 Banner Global (Activo)"
		bannerType = "\n\n⚠️ <i>El sistema individual está desactivado. Todas las cuentas usarán el mismo banner global.</i>"
	} else {
		bannerType = "\n\n✅ <i>Cada usuario tiene su propio banner con días y límites.</i>"
	}

	markup := &tele.ReplyMarkup{}
	btnPromo := markup.Data("📝 Editar Textos Promo", "edit_promo_menu")
	btnCustom := markup.Data("🌐 Activar Banner Global", "banner_set_custom")
	btnDeactivate := markup.Data("🚫 Desactivar Global (Usar Individual)", "banner_deactivate")
	btnBack := markup.Data("🔙 Volver", "menu_admins")

	markup.Inline(
		markup.Row(btnPromo),
		markup.Row(btnCustom),
		markup.Row(btnDeactivate),
		markup.Row(btnBack),
	)

	texto := fmt.Sprintf("📜 <b>Gestión de Banners SSH</b>\n\n📊 <b>Modo Actual:</b> %s%s\n\n¿Qué deseas hacer?", status, bannerType)
	return SafeEditCtx(c, b, texto, markup)
}

func handleEditPromoMenu(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()

	promoText := "🔥 ¡SERVIDORES PREMIUM A 8.5 SOLES! 🔥"
	if data.BannerPromoText != "" {
		promoText = data.BannerPromoText
	}

	promoChannel := "@orxtunnel"
	if data.BannerPromoChannel != "" {
		promoChannel = data.BannerPromoChannel
	}

	promoSupport := "@KTTOFICIAL"
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
	btnBack := markup.Data("🔙 Volver", "edit_banner")

	markup.Inline(
		markup.Row(btnText, btnChannel),
		markup.Row(btnSupport, btnBotName),
		markup.Row(btnBack),
	)

	texto := "📝 <b>Editar Textos Promocionales (Banners Individuales)</b>\n\n"
	texto += "Estos textos aparecerán en la parte inferior de los banners de cada usuario.\n\n"
	texto += fmt.Sprintf("💬 <b>Mensaje Promo:</b>\n<code>%s</code>\n\n", promoText)
	texto += fmt.Sprintf("📢 <b>Canal:</b>\n<code>%s</code>\n\n", promoChannel)
	texto += fmt.Sprintf("👤 <b>Soporte:</b>\n<code>%s</code>\n\n", promoSupport)
	texto += fmt.Sprintf("🤖 <b>Creado En:</b>\n✅ CREADO EN : <code>%s</code>", promoBotName)

	return SafeEditCtx(c, b, texto, markup)
}

func handleBannerSetCustom(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_ssh_banner")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_banner")))
	return SafeEditCtx(c, b, "📜 <b>Banner SSH Personalizado</b>\n\n✏️ <i>Escribe el texto del banner (admite HTML básico):</i>\n\nEsto se mostrará al conectar por SSH.", markup)
}

func handleEditPromoText(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_text")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "💬 <b>Editar Mensaje Promo</b>\n\n✏️ <i>Escribe el nuevo texto promocional (ej: 🔥 ¡OFERTA SERVIDORES A 5$! 🔥):</i>", markup)
}

func handleEditPromoChannel(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_channel")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "📢 <b>Editar Canal Promo</b>\n\n✏️ <i>Escribe el @usuario de tu canal (ej: @MiCanalVIP):</i>", markup)
}

func handleEditPromoSupport(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_support")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "👤 <b>Editar Soporte Promo</b>\n\n✏️ <i>Escribe tu @usuario de Telegram para soporte (ej: @TuUsuario):</i>", markup)
}

func handleEditPromoBotName(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_promo_botname")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "edit_promo_menu")))
	return SafeEditCtx(c, b, "🤖 <b>Editar Nombre del Bot</b>\n\n✏️ <i>Escribe el @usuario de tu bot (ej: @MiSuperVPN_bot):</i>\n\nEl banner mantendrá el prefijo \"✅ CREADO EN : \".", markup)
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
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "edit_banner")))
	return SafeEditCtx(c, b, "✅ <b>Banner Global desactivado.</b>\n\n<i>Se ha vuelto al sistema de banners individuales.</i>", markup)
}

func handleResetHistoryConfirm(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	btnYes := markup.Data("✅ Sí, Limpiar", "reset_history_exec")
	btnNo := markup.Data("❌ No, Cancelar", "menu_admins")
	markup.Inline(markup.Row(btnYes, btnNo))

	return SafeEditCtx(c, b, "⚠️ <b>¿Estás seguro de limpiar el historial?</b>\n\nSe borrarán todos los IDs de usuarios registrados (el broadcast ya no les llegará hasta que vuelvan a iniciar el bot).", markup)
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

	return SafeEditCtx(c, b, "🚨 <b>ADVERTENCIA: REINICIO DEL SERVIDOR</b>\n\n¿Estás seguro de que quieres reiniciar la VPS? Todas las conexiones actuales se cortarán.", markup)
}

func handleServerRebootExec(c tele.Context, b *tele.Bot) error {
	c.Edit("⏳ <b>Reiniciando VPS...</b> el bot estará offline unos minutos.", tele.ModeHTML)
	exec.Command("reboot").Run()
	return nil
}

// === SISTEMA DE ACTUALIZACIONES (UPDATER) ===

func handleMenuUpdater(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return c.Send("⛔ Solo administradores pueden usar esta función.")
	}

	data, _ := db.Load()
	autoStatus := "🔴 Desactivada"
	if data.AutoUpdate {
		autoStatus = "🟢 Activada"
	}

	text := "🔄 <b>Sistema de Actualizaciones (GitHub)</b>\n\n"
	text += "Versión Actual: <b>" + sys.CurrentVersion + "</b>\n"
	text += "Auto-Actualización: <b>" + autoStatus + "</b>\n\n"
	text += "Puedes buscar si hay una nueva versión disponible o activar la actualización automática (el bot revisará cada 12 horas)."

	markup := &tele.ReplyMarkup{}
	btnCheck := markup.Data("🔍 Buscar Actualización", "updater_check")
	btnAuto := markup.Data("⚙️ Auto-Update: "+autoStatus, "updater_toggle_auto")
	btnForce := markup.Data("⚠️ Forzar Reinstalación (Dev)", "updater_run")
	btnBack := markup.Data("🔙 Volver a Ajustes", "menu_admins")

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
	btnBack := markup.Data("🔙 Volver", "menu_updater")

	if err != nil {
		markup.Inline(markup.Row(btnBack))
		return SafeEditCtx(c, b, "❌ <b>Error al buscar actualizaciones:</b>\n"+err.Error(), markup)
	}

	if !hasUpdate {
		btnForceNow := markup.Data("⚠️ Forzar Reinstalación", "updater_run")
		markup.Inline(
			markup.Row(btnForceNow),
			markup.Row(btnBack),
		)
		return SafeEditCtx(c, b, "✅ <b>Estás en la última versión.</b>\nVersión actual: "+sys.CurrentVersion+"\nVersión remota: "+newVer, markup)
	}

	btnUpdateNow := markup.Data("⚡ Actualizar a v"+newVer, "updater_run")
	markup.Inline(
		markup.Row(btnUpdateNow),
		markup.Row(btnBack),
	)

	return SafeEditCtx(c, b, "🎉 <b>¡Nueva actualización encontrada!</b>\n\nVersión actual: "+sys.CurrentVersion+"\nNueva versión: <b>"+newVer+"</b>\n\n¿Deseas actualizar el bot ahora mismo? El servicio se reiniciará por unos 15 segundos.", markup)
}

func handleUpdaterRun(c tele.Context, b *tele.Bot) error {
	if !isAdmin(c.Chat().ID) {
		return nil
	}

	c.Send("⚡ <b>Iniciando actualización...</b>\nDescargando y compilando desde GitHub. El bot no responderá durante unos 15 segundos.", tele.ModeHTML)
	
	err := sys.RunUpdate()
	if err != nil {
		return c.Send("❌ Error al iniciar el actualizador: " + err.Error())
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
	btnBack := markup.Data("🔙 Volver", "menu_admins")

	markup.Inline(
		markup.Row(btnToggle),
		markup.Row(btnBack),
	)

	texto := "🕒 <b>CONFIGURACIÓN DE AUTO-REINICIO</b>\n"
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>El servidor se reiniciará automáticamente cuando alcance 24 Horas de Uptime continuo.</i>\n\n"
	texto += fmt.Sprintf("📊 <b>Estado:</b> %s\n", status)
	texto += "━━━━━━━━━━━━━━\n"
	texto += "<i>Selecciona una opción:</i>"

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
	
	btnBanUser := markup.Data("➕ Banear Usuario", "ban_user_prompt")
	btnBack := markup.Data("🔙 Volver", "menu_admins")
	
	var rows []tele.Row
	rows = append(rows, markup.Row(btnBanUser))
	
	texto := "🚫 <b>GESTIÓN DE USUARIOS BANEADOS</b>\n━━━━━━━━━━━━━━\n"
	if len(data.BannedUsers) == 0 {
		texto += "<i>No hay usuarios baneados.</i>\n\n"
	} else {
		texto += "<i>Selecciona un usuario para quitarle el Ban:</i>\n\n"
		for id, info := range data.BannedUsers {
			rows = append(rows, markup.Row(markup.Data(fmt.Sprintf("✅ Desbanear a %s", info.Name), "unban_user", id)))
			texto += fmt.Sprintf("👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n📝 Motivo: <i>%s</i>\n📅 Fecha: %s\n\n", info.Name, id, info.Reason, info.Date)
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
	return SafeEditCtx(c, b, "➕ <b>Banear Usuario</b>\n\n📝 <b>Paso 1/3:</b> Escribe el <b>ID numérico</b> del usuario de Telegram que deseas banear:", markup)
}

func handleUnbanUser(c tele.Context, b *tele.Bot) error {
	id := c.Data()
	db.Update(func(data *db.ConfigData) error {
		delete(data.BannedUsers, id)
		return nil
	})
	c.Respond(&tele.CallbackResponse{Text: "✅ Usuario desbaneado", ShowAlert: true})
	return handleMenuBans(c, b)
}

func handleAdminAccessType(c tele.Context, b *tele.Bot, isFullAccess bool) error {
	chatID := c.Chat().ID
	id := GetTempValue(chatID, "admin_id")
	alias := GetTempValue(chatID, "admin_alias")
	daysStr := GetTempValue(chatID, "admin_days")

	if id == "" || daysStr == "" {
		return c.Respond(&tele.CallbackResponse{Text: "❌ Sesión expirada o datos incompletos.", ShowAlert: true})
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "❌ Error al leer los días.", ShowAlert: true})
	}

	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	db.Update(func(data *db.ConfigData) error {
		data.Admins[id] = db.AdminInfo{Alias: alias, Expire: expireDate, FullAccess: isFullAccess}
		return nil
	})

	tipoAcceso := "Normal (Limitado)"
	if isFullAccess {
		tipoAcceso = "Total (SuperAdmin)"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver a Admins", "menu_admins")))
	
	msg := fmt.Sprintf("✅ <b>Administrador Registrado</b>\n\n👤 <b>Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n📅 <b>Vence:</b> %s (%d días)\n👑 <b>Tipo de Acceso:</b> %s", alias, id, expireDate, days, tipoAcceso)
	SafeEditCtx(c, b, msg, markup)

	// Send Welcome Message to the new admin
	newAdminID, errParse := strconv.ParseInt(id, 10, 64)
	if errParse == nil {
		welcomeMsg := "🎉 <b>¡Felicidades! Has sido ascendido a Administrador.</b>\n\n" +
			"Bienvenido a tu nuevo panel de control. Ahora tienes acceso a herramientas avanzadas para gestionar usuarios y monitorear el servicio.\n\n" +
			"📅 <b>Tu suscripción es válida hasta:</b> <code>" + expireDate + "</code>\n" +
			"👑 <b>Tipo de Acceso:</b> " + tipoAcceso + "\n\n" +
			"<i>Escribe /start o usa el menú para empezar a trabajar.</i>"
		b.Send(&tele.User{ID: newAdminID}, welcomeMsg, tele.ModeHTML)
	}

	DeleteUserStep(chatID)

	return c.Respond(&tele.CallbackResponse{Text: "✅ Admin creado exitosamente."})
}

func handleMenuAdsConfig(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_ads_config_script")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "menu_admins")))

	texto := "⚙️ <b>CONFIGURACIÓN DE MONETIZACIÓN</b>\n\n" +
		"Paso 1: Crea tu app en Monetag, genera una zona y copia la etiqueta <b>&lt;script&gt;</b> que te proporcionan (la que empieza por <code>&lt;script src='//libtl.com...</code>).\n\n" +
		"Por favor, pégala aquí y envíamela en un mensaje."

	return SafeEditCtx(c, b, texto, markup)
}

func processAdsConfigSteps(step, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))

	switch step {
	case "awaiting_ads_config_script":
		if !strings.Contains(text, "<script") || !strings.Contains(text, "</script>") {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ El texto enviado no parece una etiqueta &lt;script&gt; válida. Intenta de nuevo:", markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}
		
		temp := GetTempData(chatID)
		temp["ads_script"] = text
		SetTempData(chatID, temp)
		SetUserStep(chatID, "awaiting_ads_config_rewarded")

		texto := "✅ <b>Script guardado.</b>\n\n" +
			"Paso 2: Ahora copia el código de activación del formato <b>Rewarded Interstitial</b> (el bloque que contiene <code>show_XXXXXXX().then(...)</code>).\n\n" +
			"Por favor, pégalo aquí y envíamelo."
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ads_config_rewarded":
		re := regexp.MustCompile(`show_\d+`)
		match := re.FindString(text)
		if match == "" {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ No se encontró ninguna función <code>show_XXXXXXX</code> en el código enviado. Intenta de nuevo:", markupCancel)
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
			"Paso 3: Descarga el archivo ZIP, extrae el <code>monetag_miniapp.html</code> y súbelo a Vercel, GitHub Pages u otro hosting.\n\n" +
			"Una vez que esté en línea, envíame la <b>URL completa</b> (ej: <code>https://mi-app.vercel.app/monetag_miniapp.html</code>)."
		
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ads_config_url":
		if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
			msg, _ := SafeEdit(chatID, b, lastMsg, "⚠️ La URL debe empezar por http:// o https://. Intenta de nuevo:", markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}

		db.Update(func(data *db.ConfigData) error {
			data.WebAppURL = strings.TrimSpace(text)
			data.Monetization = true
			return nil
		})

		DeleteUserStep(chatID)
		texto := "🎉 <b>¡Monetización configurada exitosamente!</b>\n\n" +
			"El sistema ahora redirigirá a los usuarios públicos a tu MiniApp antes de crear o renovar cuentas.\n\n" +
			"Puedes desactivarla en cualquier momento desde Ajustes Pro."
		
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver a Ajustes Pro", "menu_admins")))
		
		msg, _ := SafeEdit(chatID, b, lastMsg, texto, markupBack)
		SetLastBotMsg(chatID, msg)
		return nil
	}

	return nil
}
