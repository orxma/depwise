package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	tele "gopkg.in/telebot.v3"
)

// HandleAutoRenew triggers when a user clicks the "Renew" button on an expiry alert.
// It sets the state and shows the Adsgram wall.
func HandleAutoRenew(c tele.Context, b *tele.Bot, cbType string) error {
	identifier := strings.TrimSpace(c.Callback().Data)
	if identifier == "" {
		return c.Respond(&tele.CallbackResponse{Text: "Invalid data"})
	}

	chatID := c.Chat().ID
	data, _ := db.Load()
	if !data.Monetization || data.WebAppURL == "" {
		return executeRenewal(c, b, cbType, identifier)
	}

	// Save pending renewal state
	SetUserStep(chatID, "pending_ad_renew_"+cbType)
	
	temp := make(map[string]string)
	temp["renew_identifier"] = identifier
	SetTempData(chatID, temp)

	// URL of the Monetag Mini App
	webAppURL := data.WebAppURL

	menu := &tele.ReplyMarkup{}
	btnAd := menu.WebApp(i18n.T(chatID, "ads.watch_button"), &tele.WebApp{URL: webAppURL})
	btnCancel := menu.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")

	menu.Inline(menu.Row(btnAd), menu.Row(btnCancel))

	b.Send(c.Chat(), i18n.T(chatID, "alert.renew_prompt"), menu, tele.ModeHTML)
	return c.Respond()
}

// executeRenewal applies the renewal logic based on the user's tier and protocol
func executeRenewal(c tele.Context, b *tele.Bot, cbType, identifier string) error {
	chatID := c.Chat().ID
	data, err := db.Load()
	if err != nil {
		return err
	}

	days := data.GetMaxDaysPublic()
	if _, isAdmin := data.Admins[fmt.Sprintf("%d", chatID)]; isAdmin {
		days = data.GetMaxDaysAdmin()
	}

	var renewed bool

	// Helper to add days to existing expire date string
	calcNewExpire := func(existing string) string {
		baseDate := time.Now()
		parsed, errParse := time.Parse("2006-01-02", existing)
		if errParse == nil && parsed.After(baseDate) {
			baseDate = parsed
		}
		return baseDate.AddDate(0, 0, days).Format("2006-01-02")
	}

	switch cbType {
	case "ssh":
		errRenew := sys.RenewSSHUser(identifier, days)
		if errRenew == nil {
			renewed = true
			db.Update(func(d *db.ConfigData) error {
				if d.SSHTimeUsers != nil {
					d.SSHTimeUsers[identifier] = calcNewExpire(d.SSHTimeUsers[identifier])
				}
				alertKey := fmt.Sprintf("ssh:%s", identifier)
				if d.Alerts1DaySent != nil { d.Alerts1DaySent[alertKey] = false }
				if d.Alerts1HourSent != nil { d.Alerts1HourSent[alertKey] = false }
				return nil
			})
		}
	case "zi":
		renewed = true
		db.Update(func(d *db.ConfigData) error {
			if d.ZivpnUsers != nil {
				d.ZivpnUsers[identifier] = calcNewExpire(d.ZivpnUsers[identifier])
			}
			alertKey := fmt.Sprintf("zi:%s", identifier)
			if d.Alerts1DaySent != nil { d.Alerts1DaySent[alertKey] = false }
			if d.Alerts1HourSent != nil { d.Alerts1HourSent[alertKey] = false }
			return nil
		})
	case "xray":
		renewed = true
		db.Update(func(d *db.ConfigData) error {
			if d.XrayUsers != nil {
				user := d.XrayUsers[identifier]
				user.Expire = calcNewExpire(user.Expire)
				d.XrayUsers[identifier] = user
			}
			alertKey := fmt.Sprintf("xray:%s", identifier)
			if d.Alerts1DaySent != nil { d.Alerts1DaySent[alertKey] = false }
			if d.Alerts1HourSent != nil { d.Alerts1HourSent[alertKey] = false }
			return nil
		})
	}

	if renewed {
		msg := i18n.Tf(chatID, "alert.renew_success", days)
		b.Send(c.Chat(), msg, tele.ModeHTML)
	} else {
		b.Send(c.Chat(), i18n.T(chatID, "error.general"), tele.ModeHTML)
	}

	DeleteUserStep(chatID)
	return nil
}
