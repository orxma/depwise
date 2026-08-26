package bot

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

func handleCrearXray(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()

	if !data.Xray.Installed {
		return c.Edit(i18n.T(chatID, "xray.not_active"), tele.ModeHTML)
	}

	if !data.PublicAccess && !isAdmin(chatID) {
		return c.Edit(i18n.T(chatID, "xray.access_denied"), tele.ModeHTML)
	}

	// Verificar cuota de cuentas VMess (SuperAdmin sin límite)
	if !isFullAdmin(chatID) {
		maxAccounts := data.GetMaxXrayPublic()
		if isAdmin(chatID) {
			maxAccounts = data.GetMaxXrayAdmin()
		}

		// Contar cuentas existentes de este usuario
		currentCount := 0
		for _, user := range data.XrayUsers {
			if user.Owner == fmt.Sprintf("%d", chatID) {
				currentCount++
			}
		}

		if currentCount >= maxAccounts {
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			return SafeEditCtx(c, b, i18n.Tf(chatID, "xray.limit_reached", currentCount, maxAccounts), markup)
		}
	}

	if !isAdmin(chatID) {
		data, _ := db.Load()
		if data.Monetization {
			return sendAdWall(c, b, "xray")
		}
	}

	SetUserStep(chatID, "awaiting_xray_alias")
	SetTempData(chatID, make(map[string]string))
	lastMsg := GetLastBotMsg(chatID)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.alias_prompt"), markup)
	return err
}

func handleManageXrayUsers(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()

	res := i18n.T(chatID, "xray.manage_title")
	count := 0
	
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	for uid, user := range data.XrayUsers {
		ownerID, _ := strconv.ParseInt(user.Owner, 10, 64)
		if isFullAdmin(chatID) || ownerID == chatID {
			label := fmt.Sprintf("👤 %s (%s) [%s]", user.Alias, user.Expire, protoLabel(user.Protocol))
			res += fmt.Sprintf("• %s\n<code>%s</code>\n", label, uid)
			if isFullAdmin(chatID) {
				res += i18n.Tf(chatID, "info.owner_label", user.Owner)
			}
			res += "\n"
			
			// Botón de eliminación
			btnDel := markup.Data("🗑️ "+user.Alias, "del_xray_exec", uid)
			rows = append(rows, markup.Row(btnDel))
			count++
		}
	}

	btnBack := markup.Data(i18n.T(chatID, "btn.back"), "submenu_xray")
	rows = append(rows, markup.Row(btnBack))
	markup.Inline(rows...)

	if count == 0 {
		return SafeEditCtx(c, b, i18n.T(chatID, "xray.no_users"), markup)
	}

	res += i18n.T(chatID, "xray.select_delete")
	return SafeEditCtx(c, b, res, markup)
}

func handleSubMenuXray(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	status := i18n.T(chatID, "proto.not_installed")
	if data.Xray.Installed {
		status = i18n.T(chatID, "proto.installed")
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data(i18n.T(chatID, "btn.install_only"), "install_xray")
	btnManage := markup.Data(i18n.T(chatID, "btn.manage_xray"), "manage_xray_users")
	btnUninst := markup.Data(i18n.T(chatID, "btn.uninstall"), "uninstall_xray")
	btnBack := markup.Data(i18n.T(chatID, "btn.back"), "menu_protocols")

	// Solo SuperAdmin puede ver Instalar/Desinstalar
	if isFullAdmin(chatID) {
		if data.Xray.Installed {
			markup.Inline(markup.Row(btnManage), markup.Row(btnUninst), markup.Row(btnBack))
		} else {
			markup.Inline(markup.Row(btnInst), markup.Row(btnBack))
		}
	} else {
		// Admins y públicos solo ven gestión de usuarios
		if data.Xray.Installed {
			markup.Inline(markup.Row(btnManage), markup.Row(btnBack))
		} else {
			markup.Inline(markup.Row(btnBack))
		}
	}

	texto := i18n.Tf(chatID, "xray.title", status)
	return SafeEditCtx(c, b, texto, markup)
}

func processXraySteps(step string, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	switch step {
	case "awaiting_xray_alias":
		alias := strings.TrimSpace(text)
		if len(alias) < 3 {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.alias_short"), markupCancel)
			return err
		}

		data, _ := db.Load()
		if data.IsNameTaken(alias) {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.alias_taken"), markupCancel)
			return err
		}

		SetTempValue(chatID, "xray_alias", alias)
		SetUserStep(chatID, "awaiting_xray_protocol")
		return showXrayProtocolButtons(chatID, b, lastMsg)

	case "awaiting_xray_protocol":
		protocol := normalizeProtocol(text)
		if protocol == "" {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.proto_invalid"), markupCancel)
			return err
		}
		SetTempValue(chatID, "xray_protocol", protocol)

		if isFullAdmin(chatID) {
			SetUserStep(chatID, "awaiting_xray_days")
			_, err := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "xray.alias_saved", html.EscapeString(GetTempValue(chatID, "xray_alias"))), markupCancel)
			return err
		}
		data, _ := db.Load()
		days := data.GetMaxDaysPublic()
		if isAdmin(chatID) {
			days = data.GetMaxDaysAdmin()
		}
		return finishXrayCreation(c, b, chatID, lastMsg, GetTempValue(chatID, "xray_alias"), days)

	case "awaiting_xray_days":
		days, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || days <= 0 {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.days_invalid"), markupCancel)
			return err
		}
		alias := GetTempValue(chatID, "xray_alias")
		return finishXrayCreation(c, b, chatID, lastMsg, alias, days)
	}
	return nil
}

func finishXrayCreation(c tele.Context, b *tele.Bot, chatID int64, lastMsg *tele.Message, alias string, days int) error {
	DeleteUserStep(chatID)
	SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.generating"), nil)

	protocol := GetTempValue(chatID, "xray_protocol")
	log.Printf("[DEBUG] finishXrayCreation: chatID=%d, protocol from temp=%s", chatID, protocol)
	if protocol == "" {
		protocol = vpn.ProtoVMess
		log.Printf("[DEBUG] finishXrayCreation: protocol empty, defaulting to VMess")
	}
	credential := vpn.GenCredential(protocol)
	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	// 1. Agregar al sistema core
	err := vpn.AddXrayUser(protocol, credential, alias)
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "submenu_xray")))
		SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "xray.error", err.Error()), markup)
		return err
	}

	// 2. Guardar en DB
	db.Update(func(data *db.ConfigData) error {
		if data.XrayUsers == nil {
			data.XrayUsers = make(map[string]db.XrayUser)
		}
		handle := ""
		if c.Sender() != nil && c.Sender().Username != "" {
			handle = "@" + c.Sender().Username
		}
		data.XrayUsers[credential] = db.XrayUser{
			Alias:    alias,
			Expire:   expireDate,
			Owner:    fmt.Sprintf("%d", chatID),
			Handle:   handle,
			Protocol: protocol,
		}
		return nil
	})

	data, _ := db.Load()
	domain := data.CloudflareDomain

	var link string
	switch protocol {
	case vpn.ProtoVLESS:
		link = vpn.GenerateVlessLink(alias, credential, domain)
	case vpn.ProtoTrojan:
		link = vpn.GenerateTrojanLink(alias, credential, domain)
	default:
		link = vpn.GenerateVmessLink(alias, credential, domain)
	}

	label := protoLabel(protocol)
	res := i18n.Tf(chatID, "xray.created", label, alias, expireDate, domain, label, link)

	if isFullAdmin(chatID) {
		res += i18n.Tf(chatID, "info.owner_label", fmt.Sprintf("%d", chatID))
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	processReferralReward(chatID, b)

	_, err = SafeEdit(chatID, b, lastMsg, res, markup)
	return err
}

// normalizeProtocol convierte la respuesta del usuario en un protocolo válido.
func normalizeProtocol(text string) string {
	t := strings.ToLower(strings.TrimSpace(text))
	switch t {
	case "1", "vmess", "vmes":
		return vpn.ProtoVMess
	case "2", "vless":
		return vpn.ProtoVLESS
	case "3", "trojan", "trojan-ws":
		return vpn.ProtoTrojan
	}
	return ""
}

// protoLabel devuelve la etiqueta legible de un protocolo.
func protoLabel(p string) string {
	switch p {
	case vpn.ProtoVLESS:
		return "VLESS"
	case vpn.ProtoTrojan:
		return "Trojan"
	default:
		return "VMess"
	}
}

// showXrayProtocolButtons muestra botones para elegir el protocolo (VMess/VLESS/Trojan).
func showXrayProtocolButtons(chatID int64, b *tele.Bot, lastMsg *tele.Message) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(
			markup.Data("1️⃣ VMess", "xray_proto_vmess"),
			markup.Data("2️⃣ VLESS", "xray_proto_vless"),
		),
		markup.Row(
			markup.Data("3️⃣ Trojan", "xray_proto_trojan"),
		),
		markup.Row(
			markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion"),
		),
	)
	_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "xray.proto_prompt"), markup)
	return err
}

// handleXrayProtoSelect maneja la selección de protocolo vía botón.
func handleXrayProtoSelect(c tele.Context, b *tele.Bot, protocol string) error {
	chatID := c.Chat().ID
	_ = c.Respond(&tele.CallbackResponse{})

	proto := strings.TrimPrefix(protocol, "xray_proto_")
	log.Printf("[DEBUG] handleXrayProtoSelect: chatID=%d, protocol=%s, proto=%s", chatID, protocol, proto)
	SetTempValue(chatID, "xray_protocol", proto)

	alias := GetTempValue(chatID, "xray_alias")
	if alias == "" {
		alias = "user"
	}

	if isFullAdmin(chatID) {
		log.Printf("[DEBUG] handleXrayProtoSelect: isFullAdmin=true, setting step to awaiting_xray_days")
		SetUserStep(chatID, "awaiting_xray_days")
		lastMsg := GetLastBotMsg(chatID)
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
		_, err := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "xray.alias_saved", html.EscapeString(alias)), markup)
		return err
	}

	data, _ := db.Load()
	days := data.GetMaxDaysPublic()
	if isAdmin(chatID) {
		days = data.GetMaxDaysAdmin()
	}
	lastMsg := GetLastBotMsg(chatID)
	return finishXrayCreation(c, b, chatID, lastMsg, alias, days)
}

func handleDeleteXrayExec(c tele.Context, b *tele.Bot) error {
	uid := c.Data()
	data, _ := db.Load()
	user, exists := data.XrayUsers[uid]
	if !exists {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(c.Chat().ID, "xray.user_not_found"), ShowAlert: true})
	}

	// Borrar del núcleo
	vpn.RemoveXrayUser(uid)

	// Borrar de DB
	db.Update(func(data *db.ConfigData) error {
		delete(data.XrayUsers, uid)
		return nil
	})

	c.Respond(&tele.CallbackResponse{Text: i18n.Tf(c.Chat().ID, "xray.user_deleted", user.Alias), ShowAlert: true})
	return handleManageXrayUsers(c, b)
}
