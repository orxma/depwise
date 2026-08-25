package bot

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

func handleCrearZivpn(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID

	// Solo admins o si es publico
	data, _ := db.Load()
	if !data.PublicAccess && !isAdmin(chatID) {
		return c.Edit(i18n.T(chatID, "zivpn.access_denied"), tele.ModeHTML)
	}

	if !isFullAdmin(chatID) {
		maxAccounts := data.GetMaxZivpnPublic()
		if isAdmin(chatID) {
			maxAccounts = data.GetMaxZivpnAdmin()
		}

		currentCount := 0
		for _, ownerID := range data.ZivpnOwners {
			if ownerID == fmt.Sprintf("%d", chatID) {
				currentCount++
			}
		}

		if currentCount >= maxAccounts {
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
			return SafeEditCtx(c, b, i18n.Tf(chatID, "zivpn.limit_reached", currentCount, maxAccounts), markup)
		}
	}

	if !isAdmin(chatID) {
		data, _ := db.Load()
		if data.Monetization {
			return sendAdWall(c, b, "zivpn")
		}
	}

	SetUserStep(chatID, "awaiting_zivpn_pass")
	SetTempData(chatID, make(map[string]string))
	lastMsg := GetLastBotMsg(chatID)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "zivpn.create_title"), markup)
	return err
}

func finishZivpnCreation(c tele.Context, password string, days int, chatID int64, b *tele.Bot, lastMsg *tele.Message) error {
	// Bloquear estado inmediatamente para evitar spam/carreras
	DeleteUserStep(chatID)

	SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "zivpn.creating"), nil)

	err := vpn.AddZivpnUser(password)
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
		_, errEdit := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "zivpn.error", err), markup)
		return errEdit
	}

	// Guardar en DB con fecha de expiración
	expireDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")

	db.Update(func(data *db.ConfigData) error {
		if data.ZivpnUsers == nil {
			data.ZivpnUsers = make(map[string]string)
		}
		if data.ZivpnOwners == nil {
			data.ZivpnOwners = make(map[string]string)
		}
		data.ZivpnUsers[password] = expireDate
		data.ZivpnOwners[password] = fmt.Sprintf("%d", chatID)
		// Guardar @handle
		if c != nil && c.Sender() != nil && c.Sender().Username != "" {
			data.ZivpnHandles[password] = "@" + c.Sender().Username
		}
		// Inicializar actividad
		data.ZivpnLastActive[password] = time.Now().Format(time.RFC3339)
		return nil
	})

	// Construir mensaje de éxito con toda la info
	data, _ := db.Load()

	res := i18n.T(chatID, "zivpn.created_title")
	res += "━━━━━━━━━━━━━━\n"
	res += i18n.Tf(chatID, "zivpn.pass_label", html.EscapeString(password))
	res += i18n.Tf(chatID, "zivpn.days_label", days)
	res += i18n.Tf(chatID, "zivpn.expire_label", expireDate)
	res += "━━━━━━━━━━━━━━\n"
	res += i18n.Tf(chatID, "zivpn.ip_label", sys.GetPublicIP())

	if data.CloudflareDomain != "" {
		res += fmt.Sprintf("☁️ <b>Cloudflare:</b> <code>%s</code>\n", data.CloudflareDomain)
	}
	if data.CloudfrontDomain != "" {
		res += fmt.Sprintf("🚀 <b>Cloudfront:</b> <code>%s</code>\n", data.CloudfrontDomain)
	}

	res += "━━━━━━━━━━━━━━\n"
	res += "📢 <b>Canal:</b> @orxtunnel\n"
	res += "👨‍💻 <b>Dev:</b> @Dan3651\n"

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	processReferralReward(chatID, b)

	_, errEdit := SafeEdit(chatID, b, lastMsg, res, markup)
	return errEdit
}

// processZivpnSteps maneja los pasos de creación de ZiVPN
func processZivpnSteps(step string, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markupCancel := &tele.ReplyMarkup{}
	markupCancel.Inline(markupCancel.Row(markupCancel.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))

	switch step {
	case "awaiting_zivpn_pass":
		password := strings.TrimSpace(text)
		if len(password) < 1 {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "zivpn.password_empty"), markupCancel)
			return err
		}

		data, _ := db.Load()
		if data.IsNameTaken(password) {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "zivpn.password_taken"), markupCancel)
			return err
		}

		SetTempValue(chatID, "zivpn_pass", password)

		if isFullAdmin(chatID) {
			SetUserStep(chatID, "awaiting_zivpn_days")
			_, err := SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "zivpn.password_saved", html.EscapeString(password)), markupCancel)
			return err
		}

		data, _ = db.Load()
		days := data.GetMaxDaysPublic()
		if isAdmin(chatID) {
			days = data.GetMaxDaysAdmin()
		}

		return finishZivpnCreation(c, password, days, chatID, b, lastMsg)

	case "awaiting_zivpn_days":
		days, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || days <= 0 {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "zivpn.days_invalid"), markupCancel)
			return err
		}

		password := GetTempValue(chatID, "zivpn_pass")
		return finishZivpnCreation(c, password, days, chatID, b, lastMsg)
	}

	return nil
}
