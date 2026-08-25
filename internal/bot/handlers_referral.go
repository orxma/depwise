package bot

import (
	"fmt"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	tele "gopkg.in/telebot.v3"
)

func handleMenuReferrals(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, err := db.Load()
	if err != nil {
		return err
	}

	botUsername := b.Me.Username
	refLink := fmt.Sprintf("https://t.me/%s?start=ref_%d", botUsername, chatID)
	points := data.ReferralPoints[chatID]

	texto := i18n.Tf(chatID, "referral.menu_title", refLink, points)

	menu := &tele.ReplyMarkup{}
	
	if points > 0 {
		btnRedeemSSH := menu.Data(i18n.T(chatID, "btn.ref_redeem_ssh"), "ref_redeem_ssh")
		btnRedeemZivpn := menu.Data(i18n.T(chatID, "btn.ref_redeem_zivpn"), "ref_redeem_zivpn")
		btnRedeemXray := menu.Data(i18n.T(chatID, "btn.ref_redeem_xray"), "ref_redeem_xray")
		menu.Inline(
			menu.Row(btnRedeemSSH),
			menu.Row(btnRedeemZivpn),
			menu.Row(btnRedeemXray),
			menu.Row(menu.Data(i18n.T(chatID, "btn.back"), "back_main")),
		)
	} else {
		menu.Inline(
			menu.Row(menu.Data(i18n.T(chatID, "btn.back"), "back_main")),
		)
	}

	return SafeEditCtx(c, b, texto, menu)
}

// processReferralReward checks if a newly created account triggers a referral reward
func processReferralReward(chatID int64, b *tele.Bot) {
	data, _ := db.Load()
	referrerID := data.ReferredBy[chatID]
	
	if referrerID != 0 && !data.ReferralCompleted[chatID] {
		db.Update(func(d *db.ConfigData) error {
			d.ReferralCompleted[chatID] = true
			d.ReferralPoints[referrerID]++
			return nil
		})
		
		// Send notification to referrer
		b.Send(tele.ChatID(referrerID), i18n.T(referrerID, "referral.notify_new"), tele.ModeHTML)
	}
}

// Handler functions for redeeming
func handleRefRedeemSSH(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	if data.ReferralPoints[chatID] <= 0 {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "referral.no_points"), ShowAlert: true})
	}
	SetUserStep(chatID, "awaiting_ref_ssh_user")
	texto := i18n.T(chatID, "referral.ask_ssh_user")

	var owned []string
	chatIDStr := fmt.Sprintf("%d", chatID)
	for user, owner := range data.SSHOwners {
		if owner == chatIDStr {
			owned = append(owned, user)
		}
	}
	if len(owned) > 0 {
		texto += "\n\n" + i18n.T(chatID, "referral.your_accounts") + "\n"
		for _, u := range owned {
			texto += fmt.Sprintf("• <code>%s</code>\n", u) // u is SSH username
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return SafeEditCtx(c, b, texto, markup)
}

func handleRefRedeemZivpn(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	if data.ReferralPoints[chatID] <= 0 {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "referral.no_points"), ShowAlert: true})
	}
	SetUserStep(chatID, "awaiting_ref_zivpn_user")
	texto := i18n.T(chatID, "referral.ask_zivpn_user")

	var owned []string
	chatIDStr := fmt.Sprintf("%d", chatID)
	for user, owner := range data.ZivpnOwners {
		if owner == chatIDStr {
			owned = append(owned, user)
		}
	}
	if len(owned) > 0 {
		texto += "\n\n" + i18n.T(chatID, "referral.your_accounts") + "\n"
		for _, u := range owned {
			texto += fmt.Sprintf("• <code>%s</code>\n", u)
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return SafeEditCtx(c, b, texto, markup)
}

func handleRefRedeemXray(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()
	if data.ReferralPoints[chatID] <= 0 {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "referral.no_points"), ShowAlert: true})
	}
	SetUserStep(chatID, "awaiting_ref_xray_user")
	texto := i18n.T(chatID, "referral.ask_xray_user")

	var owned []string
	chatIDStr := fmt.Sprintf("%d", chatID)
	for uuid, user := range data.XrayUsers {
		if user.Owner == chatIDStr {
			// user.Alias or uuid? In referral they probably need to enter uuid or alias?
			// The original prompt handleRefRedeemXray expects the UUID. 
			// Wait, the prompt executeReferralRedeem uses accountID.
			// Let's show both Alias and UUID.
			owned = append(owned, fmt.Sprintf("%s (UUID: <code>%s</code>)", user.Alias, uuid))
		}
	}
	if len(owned) > 0 {
		texto += "\n\n" + i18n.T(chatID, "referral.your_accounts") + "\n"
		for _, u := range owned {
			texto += fmt.Sprintf("• %s\n", u)
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "cancelar_accion")))
	return SafeEditCtx(c, b, texto, markup)
}

// executeReferralRedeem applies 6 days to the account and deducts 1 point
func executeReferralRedeem(chatID int64, b *tele.Bot, lastMsg *tele.Message, accountID string, pType string) {
	data, _ := db.Load()
	if data.ReferralPoints[chatID] <= 0 {
		SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "referral.no_points"), nil)
		return
	}

	days := 6
	var err error
	var found bool

	if pType == "ssh" {
		_, found = data.SSHTimeUsers[accountID]
		if found {
			err = sys.RenewSSHUser(accountID, days)
			if err == nil {
				db.Update(func(d *db.ConfigData) error {
					d.ReferralPoints[chatID]--
					
					// Compute new expire
					baseDate := time.Now()
					parsed, errParse := time.Parse("2006-01-02", d.SSHTimeUsers[accountID])
					if errParse == nil && parsed.After(baseDate) {
						baseDate = parsed
					}
					d.SSHTimeUsers[accountID] = baseDate.AddDate(0, 0, days).Format("2006-01-02")
					
					return nil
				})
			}
		}
	} else if pType == "zivpn" {
		_, found = data.ZivpnUsers[accountID]
		if found {
			db.Update(func(d *db.ConfigData) error {
				d.ReferralPoints[chatID]--
				
				baseDate := time.Now()
				parsed, errParse := time.Parse("2006-01-02", d.ZivpnUsers[accountID])
				if errParse == nil && parsed.After(baseDate) {
					baseDate = parsed
				}
				d.ZivpnUsers[accountID] = baseDate.AddDate(0, 0, days).Format("2006-01-02")
				
				return nil
			})
		}
	} else if pType == "xray" {
		_, found = data.XrayUsers[accountID]
		if found {
			db.Update(func(d *db.ConfigData) error {
				d.ReferralPoints[chatID]--
				
				user := d.XrayUsers[accountID]
				baseDate := time.Now()
				parsed, errParse := time.Parse("2006-01-02", user.Expire)
				if errParse == nil && parsed.After(baseDate) {
					baseDate = parsed
				}
				user.Expire = baseDate.AddDate(0, 0, days).Format("2006-01-02")
				d.XrayUsers[accountID] = user
				
				return nil
			})
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
	
	if !found {
		SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "error.generic"), markup)
	} else if err != nil {
		SafeEdit(chatID, b, lastMsg, i18n.Tf(chatID, "error.generic", err.Error()), markup)
	} else {
		// Reload to get updated points
		d2, _ := db.Load()
		msg := i18n.Tf(chatID, "referral.success_redeem", accountID, d2.ReferralPoints[chatID])
		SafeEdit(chatID, b, lastMsg, msg, markup)
	}
}

// processRefSteps handles text input for referral redemptions
func processRefSteps(step, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	DeleteUserStep(chatID)
	
	switch step {
	case "awaiting_ref_ssh_user":
		executeReferralRedeem(chatID, b, lastMsg, text, "ssh")
		return nil
	case "awaiting_ref_zivpn_user":
		executeReferralRedeem(chatID, b, lastMsg, text, "zivpn")
		return nil
	case "awaiting_ref_xray_user":
		executeReferralRedeem(chatID, b, lastMsg, text, "xray")
		return nil
	}
	return nil
}

// registerReferralHandlers registers the buttons for redeeming
func registerReferralHandlers(b *tele.Bot) {
	b.Handle(&tele.Btn{Unique: "ref_redeem_ssh"}, func(c tele.Context) error {
		return handleRefRedeemSSH(c, b)
	})
	b.Handle(&tele.Btn{Unique: "ref_redeem_zivpn"}, func(c tele.Context) error {
		return handleRefRedeemZivpn(c, b)
	})
	b.Handle(&tele.Btn{Unique: "ref_redeem_xray"}, func(c tele.Context) error {
		return handleRefRedeemXray(c, b)
	})
}
