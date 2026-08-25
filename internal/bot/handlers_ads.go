package bot

import (
	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	tele "gopkg.in/telebot.v3"
)

// WebAppData struct to parse the data from the mini app
type WebAppData struct {
	Action string `json:"action"`
}

func sendAdWall(c tele.Context, b *tele.Bot, protocol string) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "pending_ad_"+protocol)

	data, _ := db.Load()
	webAppURL := data.WebAppURL

	// If WebAppURL is empty, skip ad and proceed immediately
	if webAppURL == "" {
		return ProcessAdCompletion(c, b, "ad_completed")
	}

	menu := &tele.ReplyMarkup{}
	btnAd := menu.WebApp(i18n.T(chatID, "ads.watch_button"), &tele.WebApp{URL: webAppURL})
	btnCancel := menu.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")

	menu.Inline(menu.Row(btnAd), menu.Row(btnCancel))

	_, err := b.Send(c.Chat(), i18n.T(chatID, "ads.required"), menu, tele.ModeHTML)
	return err
}

func ProcessAdCompletion(c tele.Context, b *tele.Bot, action string) error {
	if action == "ad_completed" {
		b.Send(c.Chat(), i18n.T(c.Chat().ID, "ads.completed"), tele.ModeHTML)

		step, ok := GetUserStepWithOk(c.Chat().ID)
		if !ok {
			return nil
		}

		// Proceed based on the pending protocol
		switch step {
		case "pending_ad_ssh":
			SetUserStep(c.Chat().ID, "awaiting_ssh_username")
			SetTempData(c.Chat().ID, make(map[string]string))
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.cancel"), "cancelar_accion")))
			msg, _ := b.Send(c.Chat(), i18n.T(c.Chat().ID, "ssh.create_title"), markup, tele.ModeHTML)
			SetLastBotMsg(c.Chat().ID, msg)

		case "pending_ad_xray":
			SetUserStep(c.Chat().ID, "awaiting_xray_alias")
			SetTempData(c.Chat().ID, make(map[string]string))
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.cancel"), "cancelar_accion")))
			msg, _ := b.Send(c.Chat(), i18n.T(c.Chat().ID, "xray.alias_prompt"), markup, tele.ModeHTML)
			SetLastBotMsg(c.Chat().ID, msg)

		case "pending_ad_zivpn":
			SetUserStep(c.Chat().ID, "awaiting_zivpn_pass")
			SetTempData(c.Chat().ID, make(map[string]string))
			markup := &tele.ReplyMarkup{}
			markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.cancel"), "cancelar_accion")))
			msg, _ := b.Send(c.Chat(), i18n.T(c.Chat().ID, "zivpn.create_title"), markup, tele.ModeHTML)
			SetLastBotMsg(c.Chat().ID, msg)

		case "pending_ad_renew_ssh":
			temp := GetTempData(c.Chat().ID)
			identifier := temp["renew_identifier"]
			return executeRenewal(c, b, "ssh", identifier)

		case "pending_ad_renew_xray":
			temp := GetTempData(c.Chat().ID)
			identifier := temp["renew_identifier"]
			return executeRenewal(c, b, "xray", identifier)

		case "pending_ad_renew_zi":
			temp := GetTempData(c.Chat().ID)
			identifier := temp["renew_identifier"]
			return executeRenewal(c, b, "zi", identifier)
		}
	} else if action == "ad_error" {
		b.Send(c.Chat(), i18n.T(c.Chat().ID, "ads.error"), tele.ModeHTML)
		DeleteUserStep(c.Chat().ID)
	}

	return nil
}
