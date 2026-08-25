package bot

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	tele "gopkg.in/telebot.v3"
)

func handleCrearSSH(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()

	if !isFullAdmin(chatID) {
		maxAccounts := data.GetMaxSSHPublic()
		if isAdmin(chatID) {
			maxAccounts = data.GetMaxSSHAdmin()
		}

		currentCount := 0
		for _, ownerID := range data.SSHOwners {
			if ownerID == fmt.Sprintf("%d", chatID) {
				currentCount++
			}
		}

		if currentCount >= maxAccounts {
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			return SafeEditCtx(c, b, i18n.Tf(chatID, "ssh.err_limit_reached", currentCount, maxAccounts), markup)
		}
	}

	if !isAdmin(chatID) {
		data, _ := db.Load()
		if data.Monetization {
			return sendAdWall(c, b, "ssh")
		}
	}

	// 1. Iniciar registro de estado
	SetUserStep(chatID, "awaiting_ssh_username")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	btnCancel := markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")
	markup.Inline(markup.Row(btnCancel))

	lastMsg := GetLastBotMsg(chatID)
	msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.create_title"), markup)
	SetLastBotMsg(chatID, msg)
	return nil
}

func handleTextInputs(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	text := c.Text()
	step, ok := GetUserStepWithOk(chatID)
	if !ok {
		return nil
	}

	// Borrar el mensaje del usuario de inmediato para mantener el chat limpio (Sink Global)
	_ = c.Delete()

	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	// Dispatcher para otros protocolos
	if strings.HasPrefix(step, "awaiting_zivpn_") {
		lastMsg := GetLastBotMsg(chatID)
		return processZivpnSteps(step, text, chatID, c, b, lastMsg)
	}
	if strings.HasPrefix(step, "awaiting_vpn_") || strings.HasPrefix(step, "awaiting_quota_") || strings.HasPrefix(step, "awaiting_rename_") || strings.HasPrefix(step, "awaiting_promo_") || strings.HasPrefix(step, "awaiting_ban_") {
		lastMsg := GetLastBotMsg(chatID)
		return processVPNSteps(step, text, chatID, c, b, lastMsg)
	}
	if strings.HasPrefix(step, "awaiting_ads_config_") {
		lastMsg := GetLastBotMsg(chatID)
		return processAdsConfigSteps(step, text, chatID, c, b, lastMsg)
	}
	if strings.HasPrefix(step, "awaiting_scanner_") {
		lastMsg := GetLastBotMsg(chatID)
		return processScannerSteps(step, text, chatID, c, b, lastMsg)
	}
	if strings.HasPrefix(step, "awaiting_xray_") {
		lastMsg := GetLastBotMsg(chatID)
		return processXraySteps(step, text, chatID, c, b, lastMsg)
	}
	if strings.HasPrefix(step, "awaiting_ref_") {
		lastMsg := GetLastBotMsg(chatID)
		return processRefSteps(step, text, chatID, c, b, lastMsg)
	}

	lastMsg := GetLastBotMsg(chatID)
	textLower := strings.ToLower(strings.TrimSpace(text))

	// Interceptar comandos de navegación para cancelar estado
	if (strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "//")) || textLower == "menu" || textLower == "salir" || textLower == "atrás" || textLower == "atras" || textLower == "cancelar" || textLower == "exit" || textLower == "back" || textLower == "cancel" {
		DeleteUserStep(chatID)
		return handleStart(c, b)
	}

	switch step {
	case "awaiting_ssh_username":
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(text) {
			msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.username_invalid"), markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}

		data, _ := db.Load()
		if data.IsNameTaken(text) {
			msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.name_taken"), markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}

		SetTempValue(chatID, "username", text)
		SetUserStep(chatID, "awaiting_ssh_password")

		markup := &tele.ReplyMarkup{}
		btnRnd := markup.Data(i18n.T(chatID, "btn.random_pass"), "ssh_rnd_pass")
		btnCancel := markup.Data(i18n.T(chatID, "btn.cancel"), "back_main")
		markup.Inline(markup.Row(btnRnd), markup.Row(btnCancel))

		msg, _ := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "ssh.username_saved", html.EscapeString(text)), markup)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ssh_password":
		SetTempValue(chatID, "password", text)
		if !isFullAdmin(chatID) {
			data, _ := db.Load()
			if isAdmin(chatID) {
				SetTempValue(chatID, "days", strconv.Itoa(data.GetMaxDaysAdmin()))
				SetTempValue(chatID, "limit", strconv.Itoa(data.GetMaxLimitAdmin()))
			} else {
				SetTempValue(chatID, "days", strconv.Itoa(data.GetMaxDaysPublic()))
				SetTempValue(chatID, "limit", strconv.Itoa(data.GetMaxLimitPublic()))
			}
			if data.SSHBanner != "" {
				return finishSSHCreation(c, b, chatID, lastMsg)
			}
			// Pedir título del banner
			SetUserStep(chatID, "awaiting_ssh_banner_title")
			markupTitle := &tele.ReplyMarkup{}
			btnDefault := markupTitle.Data(i18n.T(chatID, "btn.default_title"), "ssh_default_title")
			btnCancel2 := markupTitle.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")
			markupTitle.Inline(markupTitle.Row(btnDefault), markupTitle.Row(btnCancel2))
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.banner_title_prompt"), markupTitle)
			return err
		}
		SetUserStep(chatID, "awaiting_ssh_days")
		msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.days_prompt"), markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ssh_days":
		days, err := strconv.Atoi(text)
		if err != nil || days <= 0 {
			msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.days_invalid"), markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}
		SetTempValue(chatID, "days", text)
		SetUserStep(chatID, "awaiting_ssh_limit")
		msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.limit_prompt"), markupCancel)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ssh_limit":
		limit, err := strconv.Atoi(text)
		if err != nil || limit < 0 {
			msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.limit_invalid"), markupCancel)
			SetLastBotMsg(chatID, msg)
			return nil
		}
		SetTempValue(chatID, "limit", text)
		data, _ := db.Load()
		if data.SSHBanner != "" {
			return finishSSHCreation(c, b, chatID, lastMsg)
		}
		// Pedir título del banner
		SetUserStep(chatID, "awaiting_ssh_banner_title")

		markup := &tele.ReplyMarkup{}
		btnDef := markup.Data(i18n.T(chatID, "btn.default_title"), "ssh_default_title")
		btnCancel := markup.Data(i18n.T(chatID, "btn.cancel"), "back_main")
		markup.Inline(markup.Row(btnDef), markup.Row(btnCancel))

		msg, _ := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.banner_title_prompt"), markup)
		SetLastBotMsg(chatID, msg)
		return nil

	case "awaiting_ssh_banner_title":
		SetTempValue(chatID, "banner_title", strings.ToUpper(strings.TrimSpace(text)))
		return finishSSHCreation(c, b, chatID, lastMsg)

	case "awaiting_broadcast":
		msg := i18n.T(chatID, "broadcast.msg") + text
		data, _ := db.Load()
		success := 0
		for _, id := range data.UserHistory {
			_, err := b.Send(tele.ChatID(id), msg, tele.ModeHTML)
			if err == nil {
				success++
			}
		}
		DeleteUserStep(chatID)
		return c.Send(i18n.Tf(chatID, "broadcast.success", success))

	case "awaiting_edit_user_selection":
		user := text
		userData, _ := db.Load()
		sa, _ := strconv.ParseInt(superAdmin, 10, 64)
		if chatID != sa {
			if ownerID, ok := userData.SSHOwners[user]; !ok || ownerID != fmt.Sprintf("%d", chatID) {
				SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.not_allowed"), markupCancel)
				return nil
			}
		} else if _, ok := userData.SSHOwners[user]; !ok {
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.not_found"), markupCancel)
			return nil
		}
		SetTempValue(chatID, "edit_target", user)
		SetUserStep(chatID, "") // Clear step but retain TempData for subsequent edits
		return showEditUserMenu(c, b, user)

	case "awaiting_info_cuenta":
		DeleteUserStep(chatID)
		return processInfoCuenta(text, chatID, c, b)

	case "awaiting_edit_pass_val":
		user := GetTempValue(chatID, "edit_target")

		// Verificar que el usuario aún existe
		checkData, _ := db.Load()
		if _, userExists := checkData.SSHTimeUsers[user]; !userExists {
			DeleteUserStep(chatID)
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.user_expired"), markup)
			return nil
		}

		err := sys.UpdateSSHUserPassword(user, text)
		DeleteUserStep(chatID)
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back_edit"), "menu_editar")))
		if err != nil {
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.err_generic")+err.Error(), markup)
		} else {
			SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "edit.pass_changed", user), markup)
		}
		return nil

	case "awaiting_edit_renew_val":
		user := GetTempValue(chatID, "edit_target")
		days, _ := strconv.Atoi(text)

		// Verificar que el usuario aún existe (pudo haber expirado entre selección y envío)
		checkData, _ := db.Load()
		existingExpire, userExists := checkData.SSHTimeUsers[user]
		if !userExists {
			DeleteUserStep(chatID)
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.user_expired"), markup)
			return nil
		}

		err := sys.RenewSSHUser(user, days)
		DeleteUserStep(chatID)
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back_edit"), "menu_editar")))
		if err != nil {
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.err_generic")+err.Error(), markup)
		} else {
			// Calcular nueva fecha aditiva (sumar a la existente o desde hoy)
			db.Update(func(data *db.ConfigData) error {
				baseDate := time.Now()
				parsed, errParse := time.Parse("2006-01-02", existingExpire)
				if errParse == nil && parsed.After(baseDate) {
					baseDate = parsed
				}
				newExpire := baseDate.AddDate(0, 0, days).Format("2006-01-02")
				data.SSHTimeUsers[user] = newExpire
				delete(data.Alerts1DaySent, "SSH:"+user)
				delete(data.Alerts1HourSent, "SSH:"+user)
				title := data.SSHBannerTitles[user]
				limit := sys.GetUserMaxLogins(user)
				sys.WriteUserBanner(user, title, limit, newExpire, data)
				return nil
			})
			sys.SyncSSHDBanners()
			SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "edit.renewed", days, user), markup)
		}
		return nil

	case "awaiting_edit_limit_val":
		user := GetTempValue(chatID, "edit_target")
		limit, _ := strconv.Atoi(text)

		// Verificar que el usuario aún existe
		checkData, _ := db.Load()
		if _, userExists := checkData.SSHTimeUsers[user]; !userExists {
			DeleteUserStep(chatID)
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.user_expired"), markup)
			return nil
		}

		err := sys.SetConnectionLimit(user, limit)
		DeleteUserStep(chatID)
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back_edit"), "menu_editar")))
		if err != nil {
			SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "edit.err_generic")+err.Error(), markup)
		} else {
			// Regenerar banner con nuevo límite
			data, _ := db.Load()
			if expire, ok := data.SSHTimeUsers[user]; ok {
				title := data.SSHBannerTitles[user]
				sys.WriteUserBanner(user, title, limit, expire, data)
				sys.SyncSSHDBanners()
			}
			SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "edit.limit_changed", user), markup)
		}
		return nil

	case "awaiting_delete_user_selection":
		return processDeleteSteps(text, chatID, c, b)
	}

	return nil
}

func handleMenuEditar(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	isSA := chatID == sa
	res := i18n.T(chatID, "edit.title")
	count := 0
	for user, ownerID := range data.SSHOwners {
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			handle := data.SSHHandles[user]
			if handle != "" {
				res += fmt.Sprintf("👤 <code>%s</code> (%s)\n", user, handle)
			} else {
				res += "👤 <code>" + user + "</code>\n"
			}
			count++
		}
	}
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
	if count == 0 {
		return c.Edit(i18n.T(chatID, "edit.no_users"), markup, tele.ModeHTML)
	}
	res += i18n.T(chatID, "edit.select_user")
	SetUserStep(chatID, "awaiting_edit_user_selection")
	SetTempData(chatID, make(map[string]string))
	return c.Edit(res, markup, tele.ModeHTML)
}

func showEditUserMenu(c tele.Context, b *tele.Bot, user string) error {
	chatID := c.Chat().ID
	markup := &tele.ReplyMarkup{}
	btnPass := markup.Data(i18n.T(chatID, "btn.edit_pass"), "edit_pass")
	btnRenew := markup.Data(i18n.T(chatID, "btn.edit_renew"), "edit_renew")
	btnLimit := markup.Data(i18n.T(chatID, "btn.edit_limit"), "edit_limit")
	btnBack := markup.Data(i18n.T(chatID, "btn.back"), "menu_editar")
	markup.Inline(markup.Row(btnPass, btnRenew), markup.Row(btnLimit), markup.Row(btnBack))
	texto := i18n.Tf(chatID, "edit.user_menu", user)

	lastMsg := GetLastBotMsg(c.Chat().ID)
	_, err := SafeEdit(c.Chat().ID, b, lastMsg, texto, markup)
	return err
}

func finishSSHCreation(c tele.Context, b *tele.Bot, chatID int64, lastMsg *tele.Message) error {
	// Bloquear estado inmediatamente para evitar spam/carreras
	mData := GetTempData(chatID)
	DeleteUserStep(chatID)

	user := mData["username"]
	pass := mData["password"]
	days, _ := strconv.Atoi(mData["days"])
	limit, _ := strconv.Atoi(mData["limit"])

	SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "ssh.creating"), nil)

	// Crear usuario en el sistema
	err := sys.CreateSSHUser(user, pass, days)
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
		SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "ssh.err_create", err), markup)
		return err
	}

	// Aplicar límite
	sys.SetConnectionLimit(user, limit)

	// Guardar en DB
	db.Update(func(data *db.ConfigData) error {
		if data.SSHOwners == nil {
			data.SSHOwners = make(map[string]string)
		}
		data.SSHOwners[user] = fmt.Sprintf("%d", chatID)

		if data.SSHTimeUsers == nil {
			data.SSHTimeUsers = make(map[string]string)
		}
		// Calcular fecha de vencimiento (YYYY-MM-DD)
		expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")
		data.SSHTimeUsers[user] = expireDate

		if c.Sender() != nil && c.Sender().Username != "" {
			if data.SSHHandles == nil {
				data.SSHHandles = make(map[string]string)
			}
			data.SSHHandles[user] = "@" + c.Sender().Username
		}

		// Guardar título del banner
		bannerTitle := mData["banner_title"]
		if bannerTitle == "" {
			bannerTitle = "INTERNET ILIMITADO"
		}
		if data.SSHBannerTitles == nil {
			data.SSHBannerTitles = make(map[string]string)
		}
		data.SSHBannerTitles[user] = bannerTitle

		return nil
	})

	// Generar banner individual
	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")
	bannerTitle := mData["banner_title"]
	if bannerTitle == "" {
		bannerTitle = "INTERNET ILIMITADO"
	}
	dataBanner, _ := db.Load()
	sys.WriteUserBanner(user, bannerTitle, limit, expireDate, dataBanner)
	sys.SyncSSHDBanners()

	// Respuesta final
	ip := sys.GetPublicIP()
	dataFinal, _ := db.Load()
	res := i18n.T(chatID, "ssh.created_title")
	res += "━━━━━━━━━━━━━━\n"
	res += i18n.Tf(chatID, "ssh.user_label", user)
	res += i18n.Tf(chatID, "ssh.pass_label", html.EscapeString(pass))
	res += i18n.Tf(chatID, "ssh.days_label", days)
	res += i18n.Tf(chatID, "ssh.limit_label", limit)
	res += "━━━━━━━━━━━━━━\n"
	res += i18n.Tf(chatID, "ssh.ip_label", ip)

	res += i18n.T(chatID, "ssh.ports_title")
	res += i18n.T(chatID, "ssh.direct_port")
	if dataFinal.Dropbear != "" {
		res += i18n.Tf(chatID, "ssh.dropbear_port", dataFinal.Dropbear)
	}
	if dataFinal.SSLTunnel != "" {
		res += i18n.Tf(chatID, "ssh.ssl_port", dataFinal.SSLTunnel)
	}
	if dataFinal.Falcon != "" {
		res += i18n.Tf(chatID, "ssh.falcon_port", dataFinal.Falcon)
	}
	res += "\n"

	if dataFinal.CloudflareDomain != "" || dataFinal.CloudfrontDomain != "" {
		res += i18n.T(chatID, "ssh.cdn_title")
		if dataFinal.CloudflareDomain != "" {
			res += i18n.Tf(chatID, "ssh.cloudflare_line", dataFinal.CloudflareDomain)
		}
		if dataFinal.CloudfrontDomain != "" {
			res += i18n.Tf(chatID, "ssh.cloudfront_line", dataFinal.CloudfrontDomain)
		}
		res += "\n"
	}

	if dataFinal.SlowDNS.NS != "" {
		res += i18n.T(chatID, "ssh.slowdns_section")
		res += i18n.Tf(chatID, "ssh.ns_line", dataFinal.SlowDNS.NS)
		if dataFinal.SlowDNS.Key != "" {
			res += i18n.Tf(chatID, "ssh.key_line", dataFinal.SlowDNS.Key)
		}
		res += "\n"
	}

	if dataFinal.VayDNS.NS != "" {
		res += i18n.T(chatID, "ssh.vaydns_section")
		res += i18n.Tf(chatID, "ssh.ns_line", dataFinal.VayDNS.NS)
		if dataFinal.VayDNS.Key != "" {
			res += i18n.Tf(chatID, "ssh.key_line", dataFinal.VayDNS.Key)
		}
		res += "\n"
	}

	if dataFinal.Slipstream.NS != "" {
		res += i18n.T(chatID, "ssh.slipstream_section")
		res += i18n.Tf(chatID, "ssh.ns_line", dataFinal.Slipstream.NS)
		res += "\n"
	}

	if dataFinal.SSLTunnel != "" {
		res += i18n.T(chatID, "ssh.ws_tls_section")
		res += i18n.Tf(chatID, "ssh.ws_line", ip)
		res += i18n.Tf(chatID, "ssh.wss_line", ip)
		res += i18n.Tf(chatID, "ssh.ws_cdn_line", ip)
		res += "\n"
	}
	res += "━━━━━━━━━━━━━━\n"

	if dataFinal.CloudflareDomain != "" {
		domain := dataFinal.CloudflareDomain
		res += i18n.T(chatID, "ssh.payloads_title")
		
		res += i18n.T(chatID, "ssh.payload_ws")
		res += i18n.Tf(chatID, "ssh.payload_ws_body", domain)
		res += i18n.T(chatID, "ssh.payload_wss")
		res += i18n.Tf(chatID, "ssh.payload_wss_body", domain, domain)
		res += i18n.T(chatID, "ssh.payload_injector")
		res += i18n.Tf(chatID, "ssh.payload_inj_body", domain)

		res += i18n.T(chatID, "ssh.http_custom_title")
		res += i18n.Tf(chatID, "ssh.http_ws", domain, user, pass)
		
		if dataFinal.SSLTunnel != "" {
			res += i18n.Tf(chatID, "ssh.http_wss", domain, user, pass)
		}
		
		if dataFinal.Dropbear != "" {
			dpPort := strings.Split(dataFinal.Dropbear, ",")[0]
			dpPort = strings.TrimSpace(dpPort)
			res += i18n.Tf(chatID, "ssh.http_dropbear", domain, dpPort, user, pass)
		}
		
		res += "━━━━━━━━━━━━━━\n"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
	
	processReferralReward(chatID, b)
	
	_, err = SafeEdit(chatID, b, lastMsg, res, markup)
	return err
}

func handleCancel(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	DeleteUserStep(chatID)

	// Volver al menú
	return handleStart(c, b)
}

func handleRandomPass(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	pass := fmt.Sprintf("%06d", 100000+((time.Now().UnixNano()/1000)%900000))
	SetTempValue(chatID, "password", pass)
	lastMsg := GetLastBotMsg(chatID)

	if !isFullAdmin(chatID) {
		data, _ := db.Load()
		if isAdmin(chatID) {
			SetTempValue(chatID, "days", strconv.Itoa(data.GetMaxDaysAdmin()))
			SetTempValue(chatID, "limit", strconv.Itoa(data.GetMaxLimitAdmin()))
		} else {
			SetTempValue(chatID, "days", strconv.Itoa(data.GetMaxDaysPublic()))
			SetTempValue(chatID, "limit", strconv.Itoa(data.GetMaxLimitPublic()))
		}
		if data.SSHBanner != "" {
			return finishSSHCreation(c, b, chatID, lastMsg)
		}
		// Pedir título del banner
		SetUserStep(chatID, "awaiting_ssh_banner_title")
		markupTitle := &tele.ReplyMarkup{}
		btnDefault := markupTitle.Data(i18n.T(chatID, "btn.default_title"), "ssh_default_title")
		btnCancel := markupTitle.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")
		markupTitle.Inline(markupTitle.Row(btnDefault), markupTitle.Row(btnCancel))
		_, err := SafeEdit(chatID, b, lastMsg, "✅ Pass: "+pass+"\n\n"+i18n.T(chatID, "ssh.banner_title_prompt"), markupTitle)
		return err
	}

	SetUserStep(chatID, "awaiting_ssh_days")
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	_, err := SafeEdit(chatID, b, lastMsg, "✅ Pass: "+pass+"\n"+i18n.T(chatID, "ssh.days_prompt"), markupCancel)
	return err
}

func handleDefaultTitle(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetTempValue(chatID, "banner_title", "INTERNET ILIMITADO")
	lastMsg := GetLastBotMsg(chatID)
	return finishSSHCreation(c, b, chatID, lastMsg)
}

func HandleEditPass(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	user := GetTempValue(chatID, "edit_target")
	SetUserStep(chatID, "awaiting_edit_pass_val")
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return c.Edit(i18n.Tf(chatID, "edit.pass_prompt", user), markupCancel, tele.ModeHTML)
}

func HandleEditRenew(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	user := GetTempValue(chatID, "edit_target")
	SetUserStep(chatID, "awaiting_edit_renew_val")
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return c.Edit(i18n.Tf(chatID, "edit.renew_prompt", user), markupCancel, tele.ModeHTML)
}

func HandleEditLimit(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	user := GetTempValue(chatID, "edit_target")
	SetUserStep(chatID, "awaiting_edit_limit_val")
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return c.Edit(i18n.Tf(chatID, "edit.limit_prompt", user), markupCancel, tele.ModeHTML)
}

func handleEditSelection(c tele.Context, b *tele.Bot) error {
	return handleMenuEditar(c, b)
}

func handleDeleteSelection(c tele.Context, b *tele.Bot) error {
	return handleMenuEliminar(c, b)
}

func handleMenuInfoCuenta(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	isSA := chatID == sa

	res := i18n.T(chatID, "info.title")
	count := 0

	// 1. SSH Users
	for user, ownerID := range data.SSHOwners {
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			handle := data.SSHHandles[user]
			if handle != "" {
				res += fmt.Sprintf("👤 SSH: <code>%s</code> (%s)\n", user, handle)
			} else {
				res += fmt.Sprintf("👤 SSH: <code>%s</code>\n", user)
			}
			count++
		}
	}

	// 2. ZiVPN Users
	for pass, ownerID := range data.ZivpnOwners {
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			handle := data.ZivpnHandles[pass]
			if handle != "" {
				res += fmt.Sprintf("🛰️ ZiVPN: <code>%s</code> (%s)\n", pass, handle)
			} else {
				res += fmt.Sprintf("🛰️ ZiVPN: <code>%s</code>\n", pass)
			}
			count++
		}
	}

	// 3. Xray Users
	for _, user := range data.XrayUsers {
		if isSA || user.Owner == fmt.Sprintf("%d", chatID) {
			if user.Handle != "" {
				res += fmt.Sprintf("💎 Xray: <code>%s</code> (%s)\n", user.Alias, user.Handle)
			} else {
				res += fmt.Sprintf("💎 Xray: <code>%s</code>\n", user.Alias)
			}
			count++
		}
	}

	res += "━━━━━━━━━━━━━━\n"
	if count == 0 {
		res += i18n.T(chatID, "info.no_accounts")
	}

	res += i18n.T(chatID, "info.select_prompt")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	SetUserStep(chatID, "awaiting_info_cuenta")
	return SafeEditCtx(c, b, res, markup)
}

func processInfoCuenta(target string, chatID int64, c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	isSA := chatID == sa
	target = strings.TrimSpace(target)

	res := i18n.T(chatID, "info.result_title")

	found := false
	ownerID := ""
	accType := ""
	expire := ""
	details := ""

	// 1. Buscar en SSH
	if exp, ok := data.SSHTimeUsers[target]; ok {
		ownerID = data.SSHOwners[target]
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			found = true
			accType = "🔒 SSH / Dropbear"
			expire = exp
			limit := sys.GetUserMaxLogins(target)
			details = fmt.Sprintf("👤 <b>User:</b> <code>%s</code>\n💻 <b>Limit:</b> %d", target, limit)
		}
	}

	// 2. Buscar en ZiVPN
	if !found {
		if exp, ok := data.ZivpnUsers[target]; ok {
			ownerID = data.ZivpnOwners[target]
			if isSA || ownerID == fmt.Sprintf("%d", chatID) {
				found = true
				accType = "🛰️ ZiVPN UDP"
				expire = exp
				details = fmt.Sprintf("🔑 <b>Password:</b> <code>%s</code>", target)
			}
		}
	}

	// 3. Buscar en Xray
	if !found {
		for uid, user := range data.XrayUsers {
			if strings.EqualFold(user.Alias, target) || uid == target {
				ownerID = user.Owner
				if isSA || ownerID == fmt.Sprintf("%d", chatID) {
					found = true
					accType = "💎 Xray (VMess/VLESS/Trojan)"
					expire = user.Expire
					details = fmt.Sprintf("👤 <b>Alias:</b> <code>%s</code>\n🆔 <b>UUID:</b> <code>%s</code>", user.Alias, uid)
					break
				}
			}
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	if !found {
		return SafeEditCtx(c, b, i18n.T(chatID, "edit.not_allowed"), markup)
	}

	// Calcular días restantes
	daysLeft := 0
	parsedExpire, err := time.Parse("2006-01-02", expire)
	if err == nil {
		daysLeft = int(time.Until(parsedExpire).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
	}

	res += i18n.Tf(chatID, "info.type_label", accType)
	res += details + "\n"
	res += i18n.Tf(chatID, "info.expires_label", expire, daysLeft)
	if isSA {
		res += i18n.Tf(chatID, "info.owner_label", ownerID)
	}
	res += "━━━━━━━━━━━━━━\n"

	return SafeEditCtx(c, b, res, markup)
}
