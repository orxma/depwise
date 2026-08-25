package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

func handleInfo(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	stats := sys.GetSystemStats()
	chatID := c.Chat().ID
	ip := sys.GetPublicIP()
	rx, tx := sys.GetGlobalTraffic()

	info := i18n.T(chatID, "srv.title")
	info += i18n.Tf(chatID, "srv.ip", ip)
	info += i18n.Tf(chatID, "srv.cpu", stats.CPUModel)
	info += i18n.Tf(chatID, "srv.usage", stats.CPUUsage)
	info += i18n.Tf(chatID, "srv.ram", stats.RAMUsed, stats.RAMTotal)
	info += i18n.Tf(chatID, "srv.disk", stats.DiskUsed, stats.DiskTotal)
	info += i18n.Tf(chatID, "srv.uptime", stats.UptimeStr)

	info += i18n.T(chatID, "srv.protocols_title") + "\n"
	active := false
	if data.SlowDNS.NS != "" {
		info += i18n.Tf(chatID, "srv.proto_slowdns_ns", data.SlowDNS.NS)
		if data.SlowDNS.Key != "" {
			info += i18n.Tf(chatID, "srv.proto_slowdns_key", data.SlowDNS.Key)
		}
		active = true
	}
	if data.VayDNS.NS != "" {
		info += i18n.Tf(chatID, "srv.proto_vaydns_ns", data.VayDNS.NS)
		if data.VayDNS.Key != "" {
			info += i18n.Tf(chatID, "srv.proto_vaydns_key", data.VayDNS.Key)
		}
		active = true
	}
	if data.Zivpn {
		info += i18n.T(chatID, "srv.proto_zivpn")
		active = true
	}
	if data.UDPCustom {
		info += i18n.T(chatID, "srv.proto_udpcustom")
		active = true
	}
	if data.BadVPN {
		info += i18n.T(chatID, "srv.proto_badvpn")
		active = true
	}
	if data.Falcon != "" {
		info += i18n.Tf(chatID, "srv.proto_falcon", data.Falcon)
		active = true
	}
	if data.Dropbear != "" {
		info += i18n.Tf(chatID, "srv.proto_dropbear", data.Dropbear)
		active = true
	}
	if data.SSLTunnel != "" {
		info += i18n.Tf(chatID, "srv.proto_ssltunnel", data.SSLTunnel)
		active = true
	}
	if len(data.ProxyDT.Ports) > 0 {
		var ports []string
		for p := range data.ProxyDT.Ports {
			ports = append(ports, "<code>"+p+"</code>")
		}
		info += i18n.Tf(chatID, "srv.proto_proxydt", strings.Join(ports, ", "))
		active = true
	}
	if data.CloudflareDomain != "" {
		info += i18n.Tf(chatID, "srv.proto_cf_domain", data.CloudflareDomain)
		active = true
	}
	if data.CloudfrontDomain != "" {
		info += i18n.Tf(chatID, "srv.proto_cloudfront", data.CloudfrontDomain)
		active = true
	}

	if !active {
		info += i18n.T(chatID, "srv.no_protocols")
	}

	// just a brief summary
	info += "\n" + i18n.T(chatID, "srv.traffic_title")
	info += i18n.Tf(chatID, "srv.download", fmt.Sprintf("%.2f GB", rx))
	info += i18n.Tf(chatID, "srv.upload", fmt.Sprintf("%.2f GB", tx))

	info += i18n.T(chatID, "srv.extra_info") + data.ExtraInfo

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	return SafeEditCtx(c, b, info, markup)
}

func handleMenuOnline(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	sshOnline := sys.GetOnlineUsers()
	zivpnOnline := sys.GetZivpnOnline()
	xrayOnline := vpn.GetXrayOnlineUsers()
	data, _ := db.Load()

	res := i18n.T(chatID, "monitor.title")

	res += i18n.T(chatID, "monitor.ssh")
	if len(sshOnline) > 0 {
		for _, line := range sshOnline {
			res += line + "\n"
		}
	} else {
		res += i18n.T(chatID, "monitor.no_connections")
	}

	res += i18n.T(chatID, "monitor.zivpn")
	if len(zivpnOnline) > 0 {
		for _, line := range zivpnOnline {
			res += line + "\n"
		}
	} else {
		res += i18n.T(chatID, "monitor.no_connections")
	}

	res += i18n.T(chatID, "monitor.xray")
	if len(xrayOnline) > 0 {
		for _, email := range xrayOnline {
			// Buscar alias del usuario en la DB por el email
			alias := email
			for _, user := range data.XrayUsers {
				if user.Alias == email || user.Alias+"@vmess" == email {
					alias = user.Alias
					break
				}
			}
			res += fmt.Sprintf("👤 %s\n", alias)
		}
	} else {
		res += i18n.T(chatID, "monitor.no_connections")
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	return SafeEditCtx(c, b, res, markup)
}

// Interceptamos opciones administrativas de borrado
func handleMenuEliminar(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	data, _ := db.Load()

	// Filtrar usuarios permitidos para este chatID (o todos si es SuperAdmin)
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	isSA := chatID == sa

	res := i18n.T(chatID, "del.title")
	res += "━━━━━━━━━━━━━━\n"

	count := 0
	// Listar SSH
	res += i18n.T(chatID, "del.ssh_section") + "\n"
	for user, ownerID := range data.SSHOwners {
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			handle := data.SSHHandles[user]
			if handle != "" {
				res += fmt.Sprintf("👤 <code>%s</code> (%s)\n", user, handle)
			} else {
				res += fmt.Sprintf("👤 <code>%s</code>\n", user)
			}
			count++
		}
	}

	// Listar ZiVPN
	res += i18n.T(chatID, "del.zivpn_section") + "\n"
	for pass, ownerID := range data.ZivpnOwners {
		if isSA || ownerID == fmt.Sprintf("%d", chatID) {
			handle := data.ZivpnHandles[pass]
			if handle != "" {
				res += fmt.Sprintf("🔑 <code>%s</code> (%s)\n", pass, handle)
			} else {
				res += fmt.Sprintf("🔑 <code>%s</code>\n", pass)
			}
			count++
		}
	}

	// Listar Xray
	res += i18n.T(chatID, "del.xray_section") + "\n"
	for _, user := range data.XrayUsers {
		if isSA || user.Owner == fmt.Sprintf("%d", chatID) {
			if user.Handle != "" {
				res += fmt.Sprintf("👤 <code>%s</code> (%s)\n", user.Alias, user.Handle)
			} else {
				res += fmt.Sprintf("👤 <code>%s</code>\n", user.Alias)
			}
			count++
		}
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))

	if count == 0 {
		return c.Edit(i18n.T(chatID, "del.no_accounts"), markup, tele.ModeHTML)
	}

	res += "━━━━━━━━━━━━━━\n"
	res += i18n.T(chatID, "del.select_prompt")

	// Cambiar estado a espera de texto
	SetUserStep(chatID, "awaiting_delete_user_selection")

	return SafeEditCtx(c, b, res, markup)
}

func processDeleteSteps(text string, chatID int64, c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	target := strings.TrimSpace(text)
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	isSA := chatID == sa

	lastMsg := GetLastBotMsg(chatID)

	// 1. Identificar si es SSH
	if ownerID, exists := data.SSHOwners[target]; exists {
		if !isSA && ownerID != fmt.Sprintf("%d", chatID) {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "del.ssh_no_permission"), nil)
			return err
		}

		_ = sys.DeleteSSHUser(target)
		db.Update(func(d *db.ConfigData) error {
			delete(d.SSHOwners, target)
			delete(d.SSHHandles, target)
			delete(d.SSHBannerTitles, target)
			delete(d.SSHTimeUsers, target)
			delete(d.SSHLastActive, target)
			return nil
		})

		_ = c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "del.ssh_deleted"), ShowAlert: false})
		return handleMenuEliminar(c, b)
	}

	// 2. Identificar si es ZiVPN (usamos el password como id)
	if ownerID, exists := data.ZivpnOwners[target]; exists {
		if !isSA && ownerID != fmt.Sprintf("%d", chatID) {
			_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "del.zivpn_no_permission"), nil)
			return err
		}

		_ = vpn.RemoveZivpnUser(target)
		db.Update(func(d *db.ConfigData) error {
			delete(d.ZivpnUsers, target)
			delete(d.ZivpnOwners, target)
			delete(d.ZivpnHandles, target)
			delete(d.ZivpnLastActive, target)
			return nil
		})

		_ = c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "del.zivpn_deleted"), ShowAlert: false})
		return handleMenuEliminar(c, b)
	}

	// 3. Identificar si es Xray (Por Alias o UUID)
	for uid, user := range data.XrayUsers {
		if strings.EqualFold(user.Alias, target) || uid == target {
			if !isSA && user.Owner != fmt.Sprintf("%d", chatID) {
				_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "del.xray_no_permission"), nil)
				return err
			}

			// Borrar del núcleo y DB
			_ = vpn.RemoveXrayUser(uid)
			_ = db.Update(func(d *db.ConfigData) error {
				delete(d.XrayUsers, uid)
				return nil
			})

			_ = c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "del.xray_deleted"), ShowAlert: false})
			return handleMenuEliminar(c, b)
		}
	}

	// 4. No encontrado
	_, err := SafeEdit(chatID, b, lastMsg, i18n.T(chatID, "del.not_found"), nil)
	return err
}
