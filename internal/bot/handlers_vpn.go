package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

func handleProtocolDiag(c tele.Context, b *tele.Bot) error {
	report := vpn.GetSystemReport()
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
	return SafeEditCtx(c, b, report, markup)
}

// Interceptar "Protocolos" para ver e Iniciar SlowDNS, Zivpn o BadVPN
func handleMenuProtocols(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}

	btnSlowDNS := markup.Data("🐢 SlowDNS / Noiz DNS", "submenu_slowdns")
	btnVayDNS := markup.Data("🚀 VayDNS", "submenu_vaydns")
	btnSlipstream := markup.Data("🌊 Slipstream", "submenu_slipstream")
	btnZiVPN := markup.Data("🛰️ ZiVPN", "submenu_zivpn")
	btnBadVPN := markup.Data("🎮 BadVPN", "submenu_badvpn")
	btnUDPCustom := markup.Data("📡 UDP Custom", "submenu_udpcustom")
	btnProxy := markup.Data("🌐 ProxyDT", "submenu_proxydt")
	btnFalcon := markup.Data("🦅 Falcon", "submenu_falcon")
	btnSSL := markup.Data("📜 WS TLS HTTP", "submenu_ssl")
	btnDropbear := markup.Data("🐻 Dropbear", "submenu_dropbear")
	btnXray := markup.Data("💎 Xray (VMess)", "submenu_xray")
	btnScanner := markup.Data("🔍 Escaner", "submenu_scanner")
	btnCancel := markup.Data("🔙 Back", "back_main")

	markup.Inline(
		markup.Row(btnSlowDNS, btnVayDNS, btnSlipstream),
		markup.Row(btnZiVPN, btnBadVPN),
		markup.Row(btnUDPCustom, btnProxy),
		markup.Row(btnFalcon, btnSSL),
		markup.Row(btnDropbear, btnXray),
		markup.Row(btnScanner),
		markup.Row(markup.Data("🛡️ Network Diagnostic", "protocol_diag")),
		markup.Row(btnCancel),
	)

	texto := "⚙️ <b>Gestor de Protocolos VPN</b>\n\n"
	texto += "<i>Select a protocol to view installation or uninstallation options.</i>"

	return SafeEditCtx(c, b, texto, markup)
}

// Mover handleMenuAdmins a handlers_admins.go

func handleMenuBroadcast(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	if !isAdmin(chatID) {
		return c.Send(i18n.T(chatID, "bcast.admin_only"), tele.ModeHTML)
	}

	SetUserStep(chatID, "awaiting_vpn_broadcast")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "back_main")))

	return SafeEditCtx(c, b, i18n.T(chatID, "bcast.prompt"), markup)
}


// Sub-Menús de Protocolos
func handleSubMenuSlowDNS(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.SlowDNS.NS != "" {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar / Reconfigurar", "install_slowdns")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_slowdns")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🐢 <b>SlowDNS / Noiz DNS Management</b>\n\n📊 <b>Status:</b> %s\n🌍 <b>NS:</b> %s\n\nWhat would you like to do?", status, data.SlowDNS.NS)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuVayDNS(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.VayDNS.NS != "" {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar / Reconfigurar", "install_vaydns")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_vaydns")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🚀 <b>VayDNS Management</b>\n\n📊 <b>Status:</b> %s\n🌍 <b>NS:</b> %s\n\nWhat would you like to do?", status, data.VayDNS.NS)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuSlipstream(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.Slipstream.NS != "" {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar / Reconfigurar", "install_slipstream")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_slipstream")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🌊 <b>Slipstream Management</b>\n\n📊 <b>Status:</b> %s\n🌍 <b>Domain:</b> %s\n\nUltra-fast QUIC protocol over UDP 53.\nIdeal for SlipNet.\n\nWhat would you like to do?", status, data.Slipstream.NS)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuZiVPN(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.Zivpn {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar", "install_zivpn")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_zivpn")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🛰️ <b>ZiVPN Management</b>\n\n📊 <b>Status:</b> %s\n\nWhat would you like to do?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuUDPCustom(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.UDPCustom {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar", "install_udpcustom")
	btnUninst := markup.Data("🗑️ Full Uninstall", "uninstall_udpcustom")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("📡 <b>UDP Custom (HTTP Custom) Management</b>\n\n📊 <b>Status:</b> %s\n\nThis protocol is the one specifically used by the <b>HTTP Custom</b> app in its 'UDP Custom' option.\n\nWhat would you like to do?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuBadVPN(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.BadVPN {
		status = "✅ Installed (Ports: 7100, 7200, 7300)"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar", "install_badvpn")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_badvpn")
	btnBack := markup.Data("🔙 Back", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🎮 <b>BadVPN Management</b>\n\n📊 <b>Status:</b> %s\n\n⚙️ Listens on ports <code>7100</code>, <code>7200</code>, <code>7300</code> (automatic)\n\nWhat would you like to do?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuFalcon(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_falcon")),
		markup.Row(markup.Data("🗑️ Desinstall", "uninstall_falcon")),
		markup.Row(markup.Data("🔙 Back", "menu_protocols")),
	)
	return SafeEditCtx(c, b, "🦅 <b>Falcon Proxy Management</b>\n\nWhat would you like to do?", markup)
}

func handleSubMenuSSL(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.SSLTunnel != "" {
		status = "✅ Instalado"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_ssl")),
		markup.Row(markup.Data("🗑️ Desinstalar", "uninstall_ssl")),
		markup.Row(markup.Data("🔙 Back", "menu_protocols")),
	)
	texto := fmt.Sprintf("📜 <b>SSL Tunnel (HAProxy) Management</b>\n\n📊 <b>Status:</b> %s\n\n⚙️ Installs multi-protocol HAProxy on ports 443, 80, 8080\n🎮 <b>Required for gaming</b> (routes WebSocket → SSH → BadVPN)\n\nWhat would you like to do?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuDropbear(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.Dropbear != "" {
		status = "✅ Installed (Ports: " + data.Dropbear + ")"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_dropbear")),
		markup.Row(markup.Data("🗑️ Desinstalar", "uninstall_dropbear")),
		markup.Row(markup.Data("🔙 Back", "menu_protocols")),
	)
	texto := fmt.Sprintf("🐻 <b>Dropbear Management</b>\n\n📊 <b>Status:</b> %s\n\nYou can specify multiple ports separated by commas (e.g. 143,109)\n\nWhat would you like to do?", status)
	return SafeEditCtx(c, b, texto, markup)
}



func handleSubMenuProxyDT(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_proxydt")),
		markup.Row(markup.Data("🗑️ Desinstalar", "uninstall_proxydt")),
		markup.Row(markup.Data("🔙 Back", "menu_protocols")),
	)
	return SafeEditCtx(c, b, "🌐 <b>ProxyDT Management</b>\n\nWhat would you like to do?", markup)
}

// Handlers de Desinstalación
func handleUninstallProtocol(c tele.Context, b *tele.Bot, proto string) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: "⛔ Only the SuperAdmin (or Admin with Full Access) can uninstall protocols.", ShowAlert: true})
	}

	SafeEditCtx(c, b, fmt.Sprintf("⏳ <i>Desinstalando %s...</i>", proto), nil)
	var err error
	data, _ := db.Load()

	switch proto {
	case "SlowDNS":
		err = vpn.RemoveSlowDNS()
		data.SlowDNS = db.SlowDNSConfig{}
	case "VayDNS":
		err = vpn.RemoveVayDNS()
		data.VayDNS = db.VayDNSConfig{}
	case "ZiVPN":
		err = vpn.RemoveZiVPN()
		data.Zivpn = false
	case "BadVPN":
		err = vpn.RemoveBadVPN()
		data.BadVPN = false
	case "Falcon":
		err = vpn.RemoveFalcon()
		data.Falcon = ""
	case "SSL Tunnel":
		err = vpn.RemoveSSLTunnel()
		data.SSLTunnel = ""
	case "Dropbear":
		err = vpn.RemoveDropbear()
		data.Dropbear = ""
	case "ProxyDT":
		err = vpn.RemoveProxyDT()
		data.ProxyDT.Ports = make(map[string]string)
	case "Xray":
		err = vpn.RemoveXray()
		data.Xray.Installed = false
		data.XrayUsers = make(map[string]db.XrayUser)
	case "Slipstream":
		err = vpn.RemoveSlipstream()
		data.Slipstream = db.SlipstreamInfo{}
	}

	if err != nil {
		return c.Edit(fmt.Sprintf("❌ <b>Error uninstalling %s:</b>\n%v", proto, err), tele.ModeHTML)
	}

	db.Save(data)
	if proto == "SlowDNS" || proto == "VayDNS" || proto == "Slipstream" {
		vpn.SyncDNSDist()
	}
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
	return c.Edit(fmt.Sprintf("✅ <b>%s uninstalled successfully.</b>", proto), markup, tele.ModeHTML)
}

// Instaladores (Interacciones base)
func handleInstallSlowDNS(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	// El bloqueo mutuo se eliminó. Ahora dnsdist permite convivir a SlowDNS y VayDNS en UDP 53.
	// Slipstream ahora usa UDP 443.

	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_slowdns_domain")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🐢 <b>SlowDNS / Noiz DNS Installer</b>\n\n🌍 <i>Enter the subdomain (NS) already pointing to this server:</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallVayDNS(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	// El bloqueo mutuo se eliminó. Ahora dnsdist permite convivir a SlowDNS y VayDNS en UDP 53.
	// Slipstream ahora usa UDP 443.

	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_vaydns_domain")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🚀 <b>VayDNS Installer</b>\n\n🌍 <i>Enter the subdomain (NS) already pointing to this server:</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallSlipstream(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	// Slipstream ahora usa UDP 443, por lo que no hay conflictos.

	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_slipstream_domain")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🌊 <b>Slipstream Installer</b>\n\n🌍 <i>Enter the domain (or NS subdomain) for the QUIC tunnel:</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallZivpn(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	data, _ := db.Load()
	if data.UDPCustom {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
		return c.Edit("⚠️ <b>Protocol Conflict</b>\n\nYou cannot install <b>ZiVPN</b> while <b>UDP Custom</b> is active. Please uninstall UDP Custom first.", markup, tele.ModeHTML)
	}

	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <i>Installing ZiVPN (UDP Custom) on automatic port 5667...</i>", tele.ModeHTML)

	err := vpn.InstallZivpn("5667")
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing ZiVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	res := "✅ <b>ZiVPN Installed Successfully</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>UDP Port:</b> <code>5667</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>The udp-custom service is now active.</i>"

	data, _ = db.Load()
	data.Zivpn = true
	db.Save(data)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallBadVPN(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <i>Installing BadVPN (UDPGW) on ports 7100, 7200, 7300...</i>", tele.ModeHTML)

	err := vpn.InstallBadVPN("7300")
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing BadVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	res := "✅ <b>BadVPN Installed Successfully</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>Port 1:</b> <code>127.0.0.1:7100</code>\n"
	res += "⚙️ <b>Port 2:</b> <code>127.0.0.1:7200</code>\n"
	res += "⚙️ <b>Port 3:</b> <code>127.0.0.1:7300</code>\n"
	res += "👥 <b>Max Clients:</b> <code>500</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>The udpgw daemon is now listening on all 3 ports.</i>"

	data, _ := db.Load()
	data.BadVPN = true
	db.Save(data)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))
	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallFalcon(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_falcon_port")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🦅 <b>Falcon Proxy Installer</b>\n\n⚙️ <i>Enter the listening port (e.g. 8080):</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallSSL(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	data, _ := db.Load()
	if !data.BadVPN {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Back", "submenu_ssl")))
		b.Edit(lastMsg, "⚠️ <b>Missing Requirement</b>\n\nYou cannot install <b>HAProxy (SSL Tunnel)</b> without <b>BadVPN</b> installed first. HAProxy relies on BadVPN to forward online gaming traffic correctly.\n\nPlease install BadVPN first.", markup, tele.ModeHTML)
		return nil
	}

	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <b>Installing multi-protocol HAProxy...</b>\n\n<i>Configuring ports 443, 80, 8080 plus the internal SSH WebSocket proxy (10015).\nThis supports gaming, VoIP and streaming.\nPlease wait...</i>", tele.ModeHTML)

	err := vpn.InstallSSLTunnel("443")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))

	if err != nil {
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing HAProxy:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	ip := sys.GetPublicIP()
	res := "✅ <b>HAProxy Multi-Protocolo Instalado</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "🔒 <b>HTTPS/WSS:</b> <code>" + ip + ":443</code>\n"
	res += "🔓 <b>HTTP/WS:</b>  <code>" + ip + ":80</code>\n"
	res += "🔓 <b>Alt:</b>      <code>" + ip + ":8080</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "🎮 <b>For Gaming:</b> BadVPN UDPGW = <code>7300</code>\n"
	res += "<i>Traffic flows: App → HAProxy(443) → SSH-WS(10015) → SSH → BadVPN → Internet</i>"

	data, _ = db.Load()
	data.SSLTunnel = "443"
	db.Save(data)

	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallDropbear(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_dropbear_port")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🐻 <b>Dropbear Installer</b>\n\n⚙️ <i>Enter the listening ports separated by commas (e.g. 143,109):</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallXray(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: "⛔ Only the SuperAdmin (or Admin with Full Access) can install protocols.", ShowAlert: true})
	}

	data, _ := db.Load()

	// Candados de seguridad
	if data.CloudflareDomain == "" {
		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(markup.Data("⚙️ Pro Settings", "menu_admins")),
			markup.Row(markup.Data("🔙 Back", "submenu_xray")),
		)
		b.Edit(lastMsg, "⚠️ <b>Missing Requirement</b>\n\nYou cannot install <b>Xray</b> without first configuring a <b>Cloudflare Domain</b> in <i>Pro Settings</i> of the admin menu.\n\nThe VMess WebSocket protocol requires a domain to generate connection links.", markup, tele.ModeHTML)
		return nil
	}

	if data.SSLTunnel == "" {
		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(markup.Data("📜 WS TLS HTTP", "submenu_ssl")),
			markup.Row(markup.Data("🔙 Back", "submenu_xray")),
		)
		b.Edit(lastMsg, "⚠️ <b>Missing Requirement</b>\n\nYou cannot install <b>Xray</b> without <b>HAProxy (SSL Tunnel)</b> installed. HAProxy receives traffic on port 443 and forwards it to Xray.", markup, tele.ModeHTML)
		return nil
	}

	b.Edit(lastMsg, "⏳ <b>Installing Xray-core...</b>\n\n<i>Downloading the Xray core and configuring VMess over WebSocket on port 10002.\nThis may take a few seconds...</i>", tele.ModeHTML)

	err := vpn.InstallXray()
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "submenu_xray")))

	if err != nil {
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing Xray:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	data, _ = db.Load()
	data.Xray.Installed = true
	data.Xray.Port = 10002
	db.Save(data)

	res := "✅ <b>Xray (VMess) Installed Successfully</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>Protocolo:</b> <code>VMess + WebSocket</code>\n"
	res += "⚙️ <b>Internal Port:</b> <code>10002</code>\n"
	res += "🌍 <b>Domain:</b> <code>" + data.CloudflareDomain + "</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>You can now start managing users from the Xray menu.</i>"

	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallProxyDT(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_proxydt_port")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🌐 <b>ProxyDT Installer (Cracked)</b>\n\n⚙️ <i>Enter the listening port (e.g. 80 or 8080):</i>", markup, tele.ModeHTML)
	return nil
}

// Interceptor secuencial para los módulos VPN
func processVPNSteps(step string, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Back", "menu_protocols")))

	switch step {
	case "awaiting_vpn_broadcast":
		DeleteUserStep(chatID)

		data, _ := db.Load()
		total := len(data.UserHistory)
		success := 0
		failed := 0

		// Avisar al admin que empezó
		b.Edit(lastMsg, i18n.Tf(chatID, "bcast.sending", total), tele.ModeHTML)

		for _, id := range data.UserHistory {
			_, err := b.Send(tele.ChatID(id), i18n.T(chatID, "bcast.admin_prefix")+text, tele.ModeHTML)
			if err == nil {
				success++
			} else {
				failed++
			}
		}

		res := i18n.Tf(chatID, "bcast.result", success, failed)

		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "back_main")))
		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_admin_id":
		id := text

		// Solo numérico
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			markupRetry := &tele.ReplyMarkup{}
			markupRetry.Inline(markupRetry.Row(markupRetry.Data("❌ Cancelar", "menu_admins")))
			b.Edit(lastMsg, "❌ <b>Invalid ID:</b> Must be a number. Try again:", markupRetry, tele.ModeHTML)
			return nil
		}

		// Guardar ID temporalmente y pedir alias
		SetTempValue(chatID, "admin_id", id)
		SetUserStep(chatID, "awaiting_vpn_admin_alias")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))
		b.Edit(lastMsg, fmt.Sprintf("✅ ID: <code>%s</code>\n\n📝 <b>Step 2/2:</b> Enter a <b>name or alias</b> to identify this admin:\n\nExample: <code>Carlos</code>, <code>Reseller Lima</code>", id), markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_admin_alias":
		alias := strings.TrimSpace(text)
		if alias == "" {
			alias = "Admin"
		}
		SetTempValue(chatID, "admin_alias", alias)
		SetUserStep(chatID, "awaiting_vpn_admin_days")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))
		b.Edit(lastMsg, fmt.Sprintf("✅ Alias: <code>%s</code>\n\n📅 <b>Step 3/3:</b> How many days of access will this administrator have?\n\nExample: <code>30</code> for one month, <code>365</code> for one year.", alias), markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_admin_days":
		daysStr := strings.TrimSpace(text)
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))
			b.Edit(lastMsg, "⚠️ <b>Error:</b> Please enter a valid number of days greater than 0.\n\n📅 How many days of access will this administrator have?", markupCancel, tele.ModeHTML)
			return nil
		}

		id := GetTempValue(chatID, "admin_id")
		alias := GetTempValue(chatID, "admin_alias")
		SetTempValue(chatID, "admin_days", daysStr)
		SetUserStep(chatID, "") // Keep TempData for access type callback

		if id == "" {
			b.Edit(lastMsg, "❌ <b>Error:</b> Temporary ID not found. Please try again from the menu.", markup, tele.ModeHTML)
			return nil
		}

		markupAccess := &tele.ReplyMarkup{}
		btnNormal := markupAccess.Data("👤 Normal Access (Limited)", "add_admin_normal")
		btnFull := markupAccess.Data("👑 Full Access (SuperAdmin)", "add_admin_full")
		btnCancel := markupAccess.Data("❌ Cancelar", "menu_admins")
		
		markupAccess.Inline(
			markupAccess.Row(btnNormal),
			markupAccess.Row(btnFull),
			markupAccess.Row(btnCancel),
		)

		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Days assigned:</b> %d\n\n👤 <b>Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n\n🛡️ <b>Final Step:</b> Select the access level for this administrator:", days, alias, id), markupAccess, tele.ModeHTML)
		return nil

	case "awaiting_rename_admin_alias":
		alias := strings.TrimSpace(text)
		id := GetTempValue(chatID, "rename_admin_id")
		DeleteUserStep(chatID)

		if alias == "" {
			b.Edit(lastMsg, "❌ <b>Alias cannot be empty.</b>", markup, tele.ModeHTML)
			return nil
		}
		if id == "" {
			b.Edit(lastMsg, "❌ <b>Error:</b> Temporary ID not found. Please try again.", markup, tele.ModeHTML)
			return nil
		}

		db.Update(func(data *db.ConfigData) error {
			if _, exists := data.Admins[id]; exists {
				data.Admins[id] = db.AdminInfo{Alias: alias}
			}
			return nil
		})

		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "menu_admins")))
		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Admin Renombrado</b>\n\n👤 <b>Nuevo Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>", alias, id), markupBack, tele.ModeHTML)
		return nil

	case "awaiting_vpn_extrainfo":
		info := text
		DeleteUserStep(chatID)

		db.Update(func(data *db.ConfigData) error {
			data.ExtraInfo = info
			return nil
		})

		b.Edit(lastMsg, "✅ <b>Extra information updated successfully.</b>", markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_cloudflare":
		domain := text
		DeleteUserStep(chatID)
		db.Update(func(data *db.ConfigData) error {
			data.CloudflareDomain = domain
			return nil
		})
		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Dominio Cloudflare actualizado:</b> <code>%s</code>", domain), markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_cloudfront":
		domain := text
		DeleteUserStep(chatID)
		db.Update(func(data *db.ConfigData) error {
			data.CloudfrontDomain = domain
			return nil
		})
		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Dominio Cloudfront actualizado:</b> <code>%s</code>", domain), markup, tele.ModeHTML)
		return nil

	case "awaiting_promo_text":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoText = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Promotional text updated.</b>\nApplied to all individual banners.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_channel":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoChannel = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Promo channel updated.</b>\nApplied to all individual banners.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_support":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoSupport = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Promo support updated.</b>\nApplied to all individual banners.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_botname":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoBotName = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Bot name updated.</b>\nApplied to all individual banners.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_vpn_ssh_banner":
		banner := text
		DeleteUserStep(chatID)
		db.Update(func(data *db.ConfigData) error {
			data.SSHBanner = banner
			return nil
		})
		// Aplicar al sistema
		err := sys.SetSSHBanner(banner)
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_banner")))
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("⚠️ <b>Banner saved to the database but failed to apply:</b>\n%v", err), markupBack, tele.ModeHTML)
		} else {
			b.Edit(lastMsg, "✅ <b>SSH Banner updated and applied to the system.</b>", markupBack, tele.ModeHTML)
		}
		return nil

	case "awaiting_quota_days_public", "awaiting_quota_limit_public", "awaiting_quota_days_admin", "awaiting_quota_limit_admin", "awaiting_quota_xray_public", "awaiting_quota_xray_admin", "awaiting_quota_ssh_public", "awaiting_quota_ssh_admin", "awaiting_quota_zivpn_public", "awaiting_quota_zivpn_admin":
		val, err := strconv.Atoi(text)
		if err != nil || val <= 0 {
			markupRetry := &tele.ReplyMarkup{}
			markupRetry.Inline(markupRetry.Row(markupRetry.Data("❌ Cancelar", "edit_quotas")))
			SafeEdit(chatID, b, lastMsg, "⚠️ Invalid value. Enter a number greater than 0:", markupRetry)
			return nil
		}
		DeleteUserStep(chatID)

		var label string
		db.Update(func(data *db.ConfigData) error {
			switch step {
			case "awaiting_quota_days_public":
				data.MaxDaysPublic = val
				label = fmt.Sprintf("Public Days → %d", val)
			case "awaiting_quota_limit_public":
				data.MaxLimitPublic = val
				label = fmt.Sprintf("Public Devices → %d", val)
			case "awaiting_quota_days_admin":
				data.MaxDaysAdmin = val
				label = fmt.Sprintf("Admin Days → %d", val)
			case "awaiting_quota_limit_admin":
				data.MaxLimitAdmin = val
				label = fmt.Sprintf("Admin Devices → %d", val)
			case "awaiting_quota_xray_public":
				data.MaxXrayPublic = val
				label = fmt.Sprintf("VMess Public → %d accounts", val)
			case "awaiting_quota_xray_admin":
				data.MaxXrayAdmin = val
				label = fmt.Sprintf("VMess Admin → %d accounts", val)
			case "awaiting_quota_ssh_public":
				data.MaxSSHPublic = val
				label = fmt.Sprintf("SSH Limit Public → %d accounts", val)
			case "awaiting_quota_ssh_admin":
				data.MaxSSHAdmin = val
				label = fmt.Sprintf("SSH Limit Admin → %d accounts", val)
			case "awaiting_quota_zivpn_public":
				data.MaxZivpnPublic = val
				label = fmt.Sprintf("ZiVPN Limit Public → %d accounts", val)
			case "awaiting_quota_zivpn_admin":
				data.MaxZivpnAdmin = val
				label = fmt.Sprintf("ZiVPN Limit Admin → %d accounts", val)
			}
			return nil
		})

		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "edit_quotas")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ <b>Cuota actualizada:</b> %s", label), markupBack)
		return nil

	case "awaiting_ban_id":
		id := strings.TrimSpace(text)
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			markupRetry := &tele.ReplyMarkup{}
			markupRetry.Inline(markupRetry.Row(markupRetry.Data("❌ Cancelar", "menu_bans")))
			SafeEdit(chatID, b, lastMsg, "❌ <b>Invalid ID:</b> Must be a number. Try again:", markupRetry)
			return nil
		}
		SetTempValue(chatID, "ban_target_id", id)
		SetUserStep(chatID, "awaiting_ban_name")
		
		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_bans")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ ID: <code>%s</code>\n\n📝 <b>Step 2/3:</b> Enter the <b>Name or Alias</b> of the user to identify them in the list:", id), markupCancel)
		return nil

	case "awaiting_ban_name":
		name := strings.TrimSpace(text)
		if name == "" {
			name = "Desconocido"
		}
		SetTempValue(chatID, "ban_target_name", name)
		SetUserStep(chatID, "awaiting_ban_reason")
		
		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_bans")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ Name: <b>%s</b>\n\n📝 <b>Step 3/3:</b> Enter the <b>Ban Reason</b> (e.g. Spam, Non-payment, etc.):\n\n<i>Or type 'None' to skip.</i>", name), markupCancel)
		return nil

	case "awaiting_ban_reason":
		reason := strings.TrimSpace(text)
		id := GetTempValue(chatID, "ban_target_id")
		name := GetTempValue(chatID, "ban_target_name")
		DeleteUserStep(chatID)

		if reason == "" || strings.ToLower(reason) == "none" {
			reason = "No especificado"
		}

		db.Update(func(data *db.ConfigData) error {
			data.BannedUsers[id] = db.BannedUserInfo{
				Name:   name,
				Reason: reason,
				Date:   time.Now().Format("2006-01-02"),
			}
			return nil
		})

		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Back", "menu_bans")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ <b>User Banned Successfully</b>\n\n👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n📝 Reason: <i>%s</i>\n\nThe user can no longer interact with the bot.", name, id, reason), markupBack)
		return nil

	case "awaiting_vpn_slowdns_domain":
		data, _ := db.Load()
		if text == data.VayDNS.NS || text == data.Slipstream.NS {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "⚠️ <b>That domain is already in use</b> by another protocol (VayDNS or Slipstream).\nPlease enter a different domain (NS) for SlowDNS / Noiz DNS:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_slowdns_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>Which local port should SlowDNS / Noiz DNS forward to? (e.g. 110, 22 or 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_slowdns_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Downloading binaries and installing SlowDNS / Noiz DNS... (This takes a few seconds)</i>", tele.ModeHTML)

		pubKey, err := vpn.InstallSlowDNS(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing SlowDNS / Noiz DNS:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>SlowDNS / Noiz DNS Installed Successfully</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>NS:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("🔑 <b>Pub Key:</b> <code>%s</code>\n", pubKey)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>The service is now active in Systemd.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.SlowDNS.NS = domain
		data.SlowDNS.Port = port
		data.SlowDNS.Key = pubKey
		db.Save(data)
		vpn.SyncDNSDist()

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_vaydns_domain":
		data, _ := db.Load()
		if text == data.SlowDNS.NS || text == data.Slipstream.NS {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "⚠️ <b>That domain is already in use</b> by another protocol (SlowDNS / Noiz DNS or Slipstream).\nPlease enter a different domain (NS) for VayDNS:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_vaydns_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>Which local port should VayDNS forward to? (e.g. 110, 22 or 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_vaydns_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Downloading binaries and installing VayDNS... (This takes a few seconds)</i>", tele.ModeHTML)

		pubKey, err := vpn.InstallVayDNS(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing VayDNS:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>VayDNS Installed Successfully</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>NS:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("🔑 <b>Pub Key:</b> <code>%s</code>\n", pubKey)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>The service is now active in Systemd.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.VayDNS.NS = domain
		data.VayDNS.Port = port
		data.VayDNS.Key = pubKey
		db.Save(data)
		vpn.SyncDNSDist()

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_slipstream_domain":
		data, _ := db.Load()
		if text == data.SlowDNS.NS || text == data.VayDNS.NS {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "⚠️ <b>That domain is already in use</b> by another protocol (SlowDNS / Noiz DNS or VayDNS).\nPlease enter a different domain for Slipstream:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_slipstream_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>Which local port should Slipstream forward to? (e.g. 110, 22 or 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_slipstream_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Downloading binaries and configuring TLS for Slipstream... (This takes a few seconds)</i>", tele.ModeHTML)

		err := vpn.InstallSlipstream(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing Slipstream:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Slipstream Installed Successfully</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>Domain:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("⚙️ <b>Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>The QUIC service is now active on UDP 53.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.Slipstream.NS = domain
		data.Slipstream.Port = port
		db.Save(data)
		vpn.SyncDNSDist()

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_zivpn_port":
		port := text
		if _, err := strconv.Atoi(port); err != nil {
			b.Edit(lastMsg, "❌ <b>Invalid port.</b> Please enter numbers only (e.g. 7300).", markup, tele.ModeHTML)
			return nil
		}
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Installing ZiVPN (UDP Custom)...</i>", tele.ModeHTML)

		err := vpn.InstallZivpn(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing ZiVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>ZiVPN Installed Successfully</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("⚙️ <b>UDP Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>The udp-custom service is now active.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.Zivpn = true
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_badvpn_port":
		port := text
		if _, err := strconv.Atoi(port); err != nil {
			b.Edit(lastMsg, "❌ <b>Invalid port.</b> Please enter numbers only (e.g. 7200).", markup, tele.ModeHTML)
			return nil
		}
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Downloading and installing BadVPN...</i>", tele.ModeHTML)

		err := vpn.InstallBadVPN(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing BadVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>BadVPN Installed Successfully</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("⚙️ <b>TCP Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>The udpgw daemon is now listening.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.BadVPN = true
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_falcon_port":
		port := text

		data, _ := db.Load()
		if data.SSLTunnel != "" && (port == "80" || port == "443" || port == "8080" || port == data.SSLTunnel) {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Port already in use by HAProxy (SSL Tunnel).</b>\n\nPlease enter a different port:", markupCancel, tele.ModeHTML)
			return nil
		}
		if data.SSHWebSocket && (port == "10015" || port == "2082") {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Port already in use by SSH WebSocket.</b>\n\nPlease enter a different port:", markupCancel, tele.ModeHTML)
			return nil
		}

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Installing Falcon Proxy...</i>", tele.ModeHTML)
		ver, err := vpn.InstallFalcon(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing Falcon:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Falcon Proxy Instalado</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🦅 <b>Version:</b> <code>%s</code>\n", ver)
		res += fmt.Sprintf("⚙️ <b>Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"

		// Guardar estado
		data, _ = db.Load()
		data.Falcon = port
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_ssl_port":
		port := text
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Configuring SSL Tunnel (HAProxy)...</i>", tele.ModeHTML)
		err := vpn.InstallSSLTunnel(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing SSL Tunnel:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>SSL Tunnel Instalado</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("📜 <b>SSL Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"

		// Guardar estado
		data, _ := db.Load()
		data.SSLTunnel = port
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_dropbear_port":
		ports := text
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Configuring Dropbear (multi-port)...</i>", tele.ModeHTML)
		err := vpn.InstallDropbear(ports)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing Dropbear:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Dropbear Installed (Multi-Port)</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🐻 <b>Ports:</b> <code>%s</code>\n", ports)
		res += "🔧 <b>Buffer:</b> <code>65536</code>\n"
		res += "━━━━━━━━━━━━━━\n"

		// Guardar estado
		data, _ := db.Load()
		data.Dropbear = ports
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_proxydt_port":
		port := text
		if _, err := strconv.Atoi(port); err != nil {
			b.Edit(lastMsg, "❌ <b>Invalid port.</b> Please enter numbers only (e.g. 8080).", markup, tele.ModeHTML)
			return nil
		}

		data, _ := db.Load()
		if data.SSLTunnel != "" && (port == "80" || port == "443" || port == "8080" || port == data.SSLTunnel) {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Port already in use by HAProxy (SSL Tunnel).</b>\n\nPlease enter a different port:", markupCancel, tele.ModeHTML)
			return nil
		}
		if data.SSHWebSocket && (port == "10015" || port == "2082") {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Port already in use by SSH WebSocket.</b>\n\nPlease enter a different port:", markupCancel, tele.ModeHTML)
			return nil
		}

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Installing and configuring ProxyDT...</i>", tele.ModeHTML)

		if err := vpn.InstallProxyDT(); err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error installing the ProxyDT binary:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		err := vpn.OpenProxyDTPort(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error opening ProxyDT port:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>ProxyDT Online</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌐 <b>Port:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"

		// Guardar estado
		data, _ = db.Load()
		if data.ProxyDT.Ports == nil {
			data.ProxyDT.Ports = make(map[string]string)
		}
		data.ProxyDT.Ports[port] = "Online"
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil
	}
	return nil
}
