package bot

import (
	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

func handleInstallUDPCustom(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	if data.Zivpn {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.back"), "menu_protocols")))
		return c.Edit(i18n.T(c.Chat().ID, "udp.conflict"), markup, tele.ModeHTML)
	}

	c.Edit(i18n.T(c.Chat().ID, "udp.installing"), tele.ModeHTML)

	// Puerto de escucha UDP (2100 como en producción)
	port := "2100"

	err := vpn.InstallUDPCustom(port)
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.back"), "menu_protocols")))

	if err != nil {
		return c.Edit(i18n.Tf(c.Chat().ID, "udp.error", err), markup, tele.ModeHTML)
	}

	db.Update(func(data *db.ConfigData) error {
		data.UDPCustom = true
		return nil
	})

	return c.Edit(i18n.T(c.Chat().ID, "udp.installed"), markup, tele.ModeHTML)
}

func handleUninstallUDPCustom(c tele.Context, b *tele.Bot) error {
	c.Edit(i18n.T(c.Chat().ID, "udp.uninstalling"), tele.ModeHTML)

	err := vpn.RemoveUDPCustom()
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(c.Chat().ID, "btn.back"), "menu_protocols")))

	if err != nil {
		return c.Edit(i18n.Tf(c.Chat().ID, "udp.uninst_error", err), markup, tele.ModeHTML)
	}

	db.Update(func(data *db.ConfigData) error {
		data.UDPCustom = false
		return nil
	})

	return c.Edit(i18n.T(c.Chat().ID, "udp.uninstalled"), markup, tele.ModeHTML)
}
