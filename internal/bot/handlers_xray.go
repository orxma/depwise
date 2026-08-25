package bot

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/vpn"
	"github.com/google/uuid"
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
			label := fmt.Sprintf("👤 %s (%s)", user.Alias, user.Expire)
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

		if isFullAdmin(chatID) {
			SetUserStep(chatID, "awaiting_xray_days")
			_, err := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "xray.alias_saved", html.EscapeString(alias)), markupCancel)
			return err
		}

		data, _ = db.Load()
		days := data.GetMaxDaysPublic()
		if isAdmin(chatID) {
			days = data.GetMaxDaysAdmin()
		}
		return finishXrayCreation(c, b, chatID, lastMsg, alias, days)

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

	newUUID := uuid.New().String()
	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	// 1. Agregar al sistema core
	err := vpn.AddXrayUser(newUUID, alias)
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
		data.XrayUsers[newUUID] = db.XrayUser{
			Alias:  alias,
			Expire: expireDate,
			Owner:  fmt.Sprintf("%d", chatID),
			Handle: handle,
		}
		return nil
	})

	data, _ := db.Load()
	vmessLink := vpn.GenerateVmessLink(alias, newUUID, data.CloudflareDomain)

	res := i18n.Tf(chatID, "xray.created", alias, expireDate, data.CloudflareDomain, vmessLink)

	if isFullAdmin(chatID) {
		res += i18n.Tf(chatID, "info.owner_label", fmt.Sprintf("%d", chatID))
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
	
	processReferralReward(chatID, b)
	
	_, err = SafeEdit(chatID, b, lastMsg, res, markup)
	return err
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
