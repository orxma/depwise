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
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
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
	btnCancel := markup.Data("🔙 Volver", "back_main")

	markup.Inline(
		markup.Row(btnSlowDNS, btnVayDNS, btnSlipstream),
		markup.Row(btnZiVPN, btnBadVPN),
		markup.Row(btnUDPCustom, btnProxy),
		markup.Row(btnFalcon, btnSSL),
		markup.Row(btnDropbear, btnXray),
		markup.Row(btnScanner),
		markup.Row(markup.Data("🛡️ Diagnóstico de Red", "protocol_diag")),
		markup.Row(btnCancel),
	)

	texto := "⚙️ <b>Gestor de Protocolos VPN</b>\n\n"
	texto += "<i>Selecciona un protocolo para ver las opciones de instalación o desinstalación.</i>"

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
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🐢 <b>Gestión de SlowDNS / Noiz DNS</b>\n\n📊 <b>Estado:</b> %s\n🌍 <b>NS:</b> %s\n\n¿Qué deseas hacer?", status, data.SlowDNS.NS)
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
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🚀 <b>Gestión de VayDNS</b>\n\n📊 <b>Estado:</b> %s\n🌍 <b>NS:</b> %s\n\n¿Qué deseas hacer?", status, data.VayDNS.NS)
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
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🌊 <b>Gestión de Slipstream</b>\n\n📊 <b>Estado:</b> %s\n🌍 <b>Dominio:</b> %s\n\nProtocolo QUIC ultra-rápido sobre UDP 53.\nIdeal para SlipNet.\n\n¿Qué deseas hacer?", status, data.Slipstream.NS)
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
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🛰️ <b>Gestión de ZiVPN</b>\n\n📊 <b>Estado:</b> %s\n\n¿Qué deseas hacer?", status)
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
	btnUninst := markup.Data("🗑️ Desinstalación Completa", "uninstall_udpcustom")
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("📡 <b>Gestión de UDP Custom (HTTP Custom)</b>\n\n📊 <b>Estado:</b> %s\n\nEste protocolo es el que utiliza específicamente la aplicación <b>HTTP Custom</b> en su opción 'UDP Custom'.\n\n¿Qué deseas hacer?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuBadVPN(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.BadVPN {
		status = "✅ Instalado (Puertos: 7100, 7200, 7300)"
	}

	markup := &tele.ReplyMarkup{}
	btnInst := markup.Data("📥 Instalar", "install_badvpn")
	btnUninst := markup.Data("🗑️ Desinstalar", "uninstall_badvpn")
	btnBack := markup.Data("🔙 Volver", "menu_protocols")

	markup.Inline(markup.Row(btnInst), markup.Row(btnUninst), markup.Row(btnBack))

	texto := fmt.Sprintf("🎮 <b>Gestión de BadVPN</b>\n\n📊 <b>Estado:</b> %s\n\n⚙️ Escucha en puertos <code>7100</code>, <code>7200</code>, <code>7300</code> (automático)\n\n¿Qué deseas hacer?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuFalcon(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_falcon")),
		markup.Row(markup.Data("🗑️ Desinstall", "uninstall_falcon")),
		markup.Row(markup.Data("🔙 Volver", "menu_protocols")),
	)
	return SafeEditCtx(c, b, "🦅 <b>Gestión de Falcon Proxy</b>\n\n¿Qué deseas hacer?", markup)
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
		markup.Row(markup.Data("🔙 Volver", "menu_protocols")),
	)
	texto := fmt.Sprintf("📜 <b>Gestión de SSL Tunnel (HAProxy)</b>\n\n📊 <b>Estado:</b> %s\n\n⚙️ Instala HAProxy multi-protocolo en puertos 443, 80, 8080\n🎮 <b>Requierido para juegos</b> (redirige WebSocket → SSH → BadVPN)\n\n¿Qué deseas hacer?", status)
	return SafeEditCtx(c, b, texto, markup)
}

func handleSubMenuDropbear(c tele.Context, b *tele.Bot) error {
	data, _ := db.Load()
	status := "❌ Desinstalado"
	if data.Dropbear != "" {
		status = "✅ Instalado (Puertos: " + data.Dropbear + ")"
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_dropbear")),
		markup.Row(markup.Data("🗑️ Desinstalar", "uninstall_dropbear")),
		markup.Row(markup.Data("🔙 Volver", "menu_protocols")),
	)
	texto := fmt.Sprintf("🐻 <b>Gestión de Dropbear</b>\n\n📊 <b>Estado:</b> %s\n\nPuedes especificar múltiples puertos separados por coma (Ej: 143,109)\n\n¿Qué deseas hacer?", status)
	return SafeEditCtx(c, b, texto, markup)
}



func handleSubMenuProxyDT(c tele.Context, b *tele.Bot) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("📥 Instalar", "install_proxydt")),
		markup.Row(markup.Data("🗑️ Desinstalar", "uninstall_proxydt")),
		markup.Row(markup.Data("🔙 Volver", "menu_protocols")),
	)
	return SafeEditCtx(c, b, "🌐 <b>Gestión de ProxyDT</b>\n\n¿Qué deseas hacer?", markup)
}

// Handlers de Desinstalación
func handleUninstallProtocol(c tele.Context, b *tele.Bot, proto string) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: "⛔ Solo el SuperAdmin (o Admin con Acceso Total) puede desinstalar protocolos.", ShowAlert: true})
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
		return c.Edit(fmt.Sprintf("❌ <b>Error al desinstalar %s:</b>\n%v", proto, err), tele.ModeHTML)
	}

	db.Save(data)
	if proto == "SlowDNS" || proto == "VayDNS" || proto == "Slipstream" {
		vpn.SyncDNSDist()
	}
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
	return c.Edit(fmt.Sprintf("✅ <b>%s desinstalado correctamente.</b>", proto), markup, tele.ModeHTML)
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

	b.Edit(lastMsg, "🐢 <b>Instalador de SlowDNS / Noiz DNS</b>\n\n🌍 <i>Escribe el subdominio (NS) que ya tengas apuntado a este servidor:</i>", markup, tele.ModeHTML)
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

	b.Edit(lastMsg, "🚀 <b>Instalador de VayDNS</b>\n\n🌍 <i>Escribe el subdominio (NS) que ya tengas apuntado a este servidor:</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallSlipstream(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	// Slipstream ahora usa UDP 443, por lo que no hay conflictos.

	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_slipstream_domain")
	SetTempData(chatID, make(map[string]string))

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🌊 <b>Instalador de Slipstream</b>\n\n🌍 <i>Escribe el dominio (o subdominio NS) para el túnel QUIC:</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallZivpn(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	data, _ := db.Load()
	if data.UDPCustom {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
		return c.Edit("⚠️ <b>Conflicto de Protocolo</b>\n\nNo puedes instalar <b>ZiVPN</b> mientras <b>UDP Custom</b> esté activo. Por favor, desinstala UDP Custom primero.", markup, tele.ModeHTML)
	}

	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <i>Instalando ZiVPN (UDP Custom) en puerto automático 5667...</i>", tele.ModeHTML)

	err := vpn.InstallZivpn("5667")
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar ZiVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	res := "✅ <b>ZiVPN Instalado Correctamente</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>Puerto UDP:</b> <code>5667</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>El servicio udp-custom ya está activo.</i>"

	data, _ = db.Load()
	data.Zivpn = true
	db.Save(data)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallBadVPN(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <i>Instalando BadVPN (UDPGW) en puertos 7100, 7200, 7300...</i>", tele.ModeHTML)

	err := vpn.InstallBadVPN("7300")
	if err != nil {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar BadVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	res := "✅ <b>BadVPN Instalado Correctamente</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>Puerto 1:</b> <code>127.0.0.1:7100</code>\n"
	res += "⚙️ <b>Puerto 2:</b> <code>127.0.0.1:7200</code>\n"
	res += "⚙️ <b>Puerto 3:</b> <code>127.0.0.1:7300</code>\n"
	res += "👥 <b>Max Clients:</b> <code>500</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>El demonio udpgw ya está escuchando en los 3 puertos.</i>"

	data, _ := db.Load()
	data.BadVPN = true
	db.Save(data)

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))
	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallFalcon(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_falcon_port")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🦅 <b>Instalador de Falcon Proxy</b>\n\n⚙️ <i>Escribe el puerto de escucha (Ej: 8080):</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallSSL(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	data, _ := db.Load()
	if !data.BadVPN {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data("🔙 Volver", "submenu_ssl")))
		b.Edit(lastMsg, "⚠️ <b>Requisito Faltante</b>\n\nNo puedes instalar <b>HAProxy (SSL Tunnel)</b> sin tener <b>BadVPN</b> previamente instalado. HAProxy depende de BadVPN para reenviar el tráfico de juegos online correctamente.\n\nPor favor instala BadVPN primero.", markup, tele.ModeHTML)
		return nil
	}

	chatID := c.Chat().ID
	delete(UserSteps, chatID)

	b.Edit(lastMsg, "⏳ <b>Instalando HAProxy Multi-Protocolo...</b>\n\n<i>Configurando puertos 443, 80, 8080 + proxy SSH WebSocket interno (10015).\nEsto soporta juegos, VoIP y streaming.\nPor favor espera...</i>", tele.ModeHTML)

	err := vpn.InstallSSLTunnel("443")
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))

	if err != nil {
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar HAProxy:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	ip := sys.GetPublicIP()
	res := "✅ <b>HAProxy Multi-Protocolo Instalado</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "🔒 <b>HTTPS/WSS:</b> <code>" + ip + ":443</code>\n"
	res += "🔓 <b>HTTP/WS:</b>  <code>" + ip + ":80</code>\n"
	res += "🔓 <b>Alt:</b>      <code>" + ip + ":8080</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "🎮 <b>Para Juegos:</b> BadVPN UDPGW = <code>7300</code>\n"
	res += "<i>El tráfico fluye: App → HAProxy(443) → SSH-WS(10015) → SSH → BadVPN → Internet</i>"

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

	b.Edit(lastMsg, "🐻 <b>Instalador de Dropbear</b>\n\n⚙️ <i>Escribe los puertos de escucha separados por coma (Ej: 143,109):</i>", markup, tele.ModeHTML)
	return nil
}

func handleInstallXray(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: "⛔ Solo el SuperAdmin (o Admin con Acceso Total) puede instalar protocolos.", ShowAlert: true})
	}

	data, _ := db.Load()

	// Candados de seguridad
	if data.CloudflareDomain == "" {
		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(markup.Data("⚙️ Ajustes Pro", "menu_admins")),
			markup.Row(markup.Data("🔙 Volver", "submenu_xray")),
		)
		b.Edit(lastMsg, "⚠️ <b>Requisito Faltante</b>\n\nNo puedes instalar <b>Xray</b> sin antes configurar un <b>Dominio de Cloudflare</b> en los <i>Ajustes Pro</i> del menú administrador.\n\nEl protocolo VMess WebSocket requiere un dominio para generar los links de conexión.", markup, tele.ModeHTML)
		return nil
	}

	if data.SSLTunnel == "" {
		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(markup.Data("📜 WS TLS HTTP", "submenu_ssl")),
			markup.Row(markup.Data("🔙 Volver", "submenu_xray")),
		)
		b.Edit(lastMsg, "⚠️ <b>Requisito Faltante</b>\n\nNo puedes instalar <b>Xray</b> sin tener <b>HAProxy (SSL Tunnel)</b> instalado. HAProxy es el encargado de recibir el tráfico en el puerto 443 y redirigirlo a Xray.", markup, tele.ModeHTML)
		return nil
	}

	b.Edit(lastMsg, "⏳ <b>Instalando Xray-core...</b>\n\n<i>Descargando núcleo Xray y configurando VMess sobre WebSocket en puerto 10002.\nEsto puede tardar unos segundos...</i>", tele.ModeHTML)

	err := vpn.InstallXray()
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "submenu_xray")))

	if err != nil {
		b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar Xray:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
		return nil
	}

	data, _ = db.Load()
	data.Xray.Installed = true
	data.Xray.Port = 10002
	db.Save(data)

	res := "✅ <b>Xray (VMess) Instalado Correctamente</b>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "⚙️ <b>Protocolo:</b> <code>VMess + WebSocket</code>\n"
	res += "⚙️ <b>Puerto Interno:</b> <code>10002</code>\n"
	res += "🌍 <b>Dominio:</b> <code>" + data.CloudflareDomain + "</code>\n"
	res += "━━━━━━━━━━━━━━\n"
	res += "<i>Ahora puedes comenzar a gestionar usuarios desde el menú de Xray.</i>"

	b.Edit(lastMsg, res, markup, tele.ModeHTML)
	return nil
}

func handleInstallProxyDT(c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_vpn_proxydt_port")

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("❌ Cancelar", "cancelar_accion")))

	b.Edit(lastMsg, "🌐 <b>Instalador de ProxyDT (Cracked)</b>\n\n⚙️ <i>Escribe el puerto de escucha (Ej: 80 o 8080):</i>", markup, tele.ModeHTML)
	return nil
}

// Interceptor secuencial para los módulos VPN
func processVPNSteps(step string, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data("🔙 Volver", "menu_protocols")))

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
			b.Edit(lastMsg, "❌ <b>ID Inválido:</b> Debe ser un número. Intenta de nuevo:", markupRetry, tele.ModeHTML)
			return nil
		}

		// Guardar ID temporalmente y pedir alias
		SetTempValue(chatID, "admin_id", id)
		SetUserStep(chatID, "awaiting_vpn_admin_alias")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))
		b.Edit(lastMsg, fmt.Sprintf("✅ ID: <code>%s</code>\n\n📝 <b>Paso 2/2:</b> Escribe un <b>nombre o alias</b> para identificar a este admin:\n\nEjemplo: <code>Carlos</code>, <code>Revendedor Lima</code>", id), markupCancel, tele.ModeHTML)
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
		b.Edit(lastMsg, fmt.Sprintf("✅ Alias: <code>%s</code>\n\n📅 <b>Paso 3/3:</b> ¿Cuántos días de acceso tendrá este administrador?\n\nEjemplo: <code>30</code> para un mes, <code>365</code> para un año.", alias), markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_admin_days":
		daysStr := strings.TrimSpace(text)
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_admins")))
			b.Edit(lastMsg, "⚠️ <b>Error:</b> Por favor ingresa un número de días válido y mayor a 0.\n\n📅 ¿Cuántos días de acceso tendrá este administrador?", markupCancel, tele.ModeHTML)
			return nil
		}

		id := GetTempValue(chatID, "admin_id")
		alias := GetTempValue(chatID, "admin_alias")
		SetTempValue(chatID, "admin_days", daysStr)
		SetUserStep(chatID, "") // Keep TempData for access type callback

		if id == "" {
			b.Edit(lastMsg, "❌ <b>Error:</b> No se encontró el ID temporal. Intenta de nuevo desde el menú.", markup, tele.ModeHTML)
			return nil
		}

		markupAccess := &tele.ReplyMarkup{}
		btnNormal := markupAccess.Data("👤 Acceso Normal (Limitado)", "add_admin_normal")
		btnFull := markupAccess.Data("👑 Acceso Total (SuperAdmin)", "add_admin_full")
		btnCancel := markupAccess.Data("❌ Cancelar", "menu_admins")
		
		markupAccess.Inline(
			markupAccess.Row(btnNormal),
			markupAccess.Row(btnFull),
			markupAccess.Row(btnCancel),
		)

		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Días asignados:</b> %d\n\n👤 <b>Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>\n\n🛡️ <b>Paso Final:</b> Selecciona el nivel de acceso para este administrador:", days, alias, id), markupAccess, tele.ModeHTML)
		return nil

	case "awaiting_rename_admin_alias":
		alias := strings.TrimSpace(text)
		id := GetTempValue(chatID, "rename_admin_id")
		DeleteUserStep(chatID)

		if alias == "" {
			b.Edit(lastMsg, "❌ <b>El alias no puede estar vacío.</b>", markup, tele.ModeHTML)
			return nil
		}
		if id == "" {
			b.Edit(lastMsg, "❌ <b>Error:</b> No se encontró el ID temporal. Intenta de nuevo.", markup, tele.ModeHTML)
			return nil
		}

		db.Update(func(data *db.ConfigData) error {
			if _, exists := data.Admins[id]; exists {
				data.Admins[id] = db.AdminInfo{Alias: alias}
			}
			return nil
		})

		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "menu_admins")))
		b.Edit(lastMsg, fmt.Sprintf("✅ <b>Admin Renombrado</b>\n\n👤 <b>Nuevo Alias:</b> %s\n🆔 <b>ID:</b> <code>%s</code>", alias, id), markupBack, tele.ModeHTML)
		return nil

	case "awaiting_vpn_extrainfo":
		info := text
		DeleteUserStep(chatID)

		db.Update(func(data *db.ConfigData) error {
			data.ExtraInfo = info
			return nil
		})

		b.Edit(lastMsg, "✅ <b>Información extra actualizada correctamente.</b>", markup, tele.ModeHTML)
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
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Texto Promocional actualizado.</b>\nSe aplicó a todos los banners individuales.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_channel":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoChannel = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Canal Promo actualizado.</b>\nSe aplicó a todos los banners individuales.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_support":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoSupport = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Soporte Promo actualizado.</b>\nSe aplicó a todos los banners individuales.", markupBack, tele.ModeHTML)
		return nil

	case "awaiting_promo_botname":
		db.Update(func(data *db.ConfigData) error {
			data.BannerPromoBotName = text
			return nil
		})
		DeleteUserStep(chatID)
		go sys.RefreshAllBanners()
		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_promo_menu")))
		b.Edit(lastMsg, "✅ <b>Nombre del Bot actualizado.</b>\nSe aplicó a todos los banners individuales.", markupBack, tele.ModeHTML)
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
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_banner")))
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("⚠️ <b>Banner guardado en DB pero error al aplicar:</b>\n%v", err), markupBack, tele.ModeHTML)
		} else {
			b.Edit(lastMsg, "✅ <b>Banner SSH actualizado y aplicado al sistema.</b>", markupBack, tele.ModeHTML)
		}
		return nil

	case "awaiting_quota_days_public", "awaiting_quota_limit_public", "awaiting_quota_days_admin", "awaiting_quota_limit_admin", "awaiting_quota_xray_public", "awaiting_quota_xray_admin", "awaiting_quota_ssh_public", "awaiting_quota_ssh_admin", "awaiting_quota_zivpn_public", "awaiting_quota_zivpn_admin":
		val, err := strconv.Atoi(text)
		if err != nil || val <= 0 {
			markupRetry := &tele.ReplyMarkup{}
			markupRetry.Inline(markupRetry.Row(markupRetry.Data("❌ Cancelar", "edit_quotas")))
			SafeEdit(chatID, b, lastMsg, "⚠️ Valor inválido. Escribe un número mayor a 0:", markupRetry)
			return nil
		}
		DeleteUserStep(chatID)

		var label string
		db.Update(func(data *db.ConfigData) error {
			switch step {
			case "awaiting_quota_days_public":
				data.MaxDaysPublic = val
				label = fmt.Sprintf("Días Público → %d", val)
			case "awaiting_quota_limit_public":
				data.MaxLimitPublic = val
				label = fmt.Sprintf("Dispositivos Público → %d", val)
			case "awaiting_quota_days_admin":
				data.MaxDaysAdmin = val
				label = fmt.Sprintf("Días Admin → %d", val)
			case "awaiting_quota_limit_admin":
				data.MaxLimitAdmin = val
				label = fmt.Sprintf("Dispositivos Admin → %d", val)
			case "awaiting_quota_xray_public":
				data.MaxXrayPublic = val
				label = fmt.Sprintf("VMess Público → %d cuentas", val)
			case "awaiting_quota_xray_admin":
				data.MaxXrayAdmin = val
				label = fmt.Sprintf("VMess Admin → %d cuentas", val)
			case "awaiting_quota_ssh_public":
				data.MaxSSHPublic = val
				label = fmt.Sprintf("Límite SSH Público → %d cuentas", val)
			case "awaiting_quota_ssh_admin":
				data.MaxSSHAdmin = val
				label = fmt.Sprintf("Límite SSH Admin → %d cuentas", val)
			case "awaiting_quota_zivpn_public":
				data.MaxZivpnPublic = val
				label = fmt.Sprintf("Límite ZiVPN Público → %d cuentas", val)
			case "awaiting_quota_zivpn_admin":
				data.MaxZivpnAdmin = val
				label = fmt.Sprintf("Límite ZiVPN Admin → %d cuentas", val)
			}
			return nil
		})

		markupBack := &tele.ReplyMarkup{}
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "edit_quotas")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ <b>Cuota actualizada:</b> %s", label), markupBack)
		return nil

	case "awaiting_ban_id":
		id := strings.TrimSpace(text)
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			markupRetry := &tele.ReplyMarkup{}
			markupRetry.Inline(markupRetry.Row(markupRetry.Data("❌ Cancelar", "menu_bans")))
			SafeEdit(chatID, b, lastMsg, "❌ <b>ID Inválido:</b> Debe ser un número. Intenta de nuevo:", markupRetry)
			return nil
		}
		SetTempValue(chatID, "ban_target_id", id)
		SetUserStep(chatID, "awaiting_ban_name")
		
		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "menu_bans")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ ID: <code>%s</code>\n\n📝 <b>Paso 2/3:</b> Escribe el <b>Nombre o Alias</b> del usuario para identificarlo en la lista:", id), markupCancel)
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
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ Nombre: <b>%s</b>\n\n📝 <b>Paso 3/3:</b> Escribe el <b>Motivo del Ban</b> (ej: Spam, No pago, etc.):\n\n<i>O escribe 'Ninguno' para omitir.</i>", name), markupCancel)
		return nil

	case "awaiting_ban_reason":
		reason := strings.TrimSpace(text)
		id := GetTempValue(chatID, "ban_target_id")
		name := GetTempValue(chatID, "ban_target_name")
		DeleteUserStep(chatID)

		if reason == "" || strings.ToLower(reason) == "ninguno" {
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
		markupBack.Inline(markupBack.Row(markupBack.Data("🔙 Volver", "menu_bans")))
		SafeEdit(chatID, b, lastMsg, fmt.Sprintf("✅ <b>Usuario Baneado Exitosamente</b>\n\n👤 <b>%s</b>\n🆔 ID: <code>%s</code>\n📝 Motivo: <i>%s</i>\n\nEl usuario ya no podrá interactuar con el bot.", name, id, reason), markupBack)
		return nil

	case "awaiting_vpn_slowdns_domain":
		data, _ := db.Load()
		if text == data.VayDNS.NS || text == data.Slipstream.NS {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "⚠️ <b>Ese dominio ya está en uso</b> por otro protocolo (VayDNS o Slipstream).\nPor favor, ingresa un dominio (NS) distinto para SlowDNS / Noiz DNS:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_slowdns_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>¿A qué puerto local quieres redirigir SlowDNS / Noiz DNS? (Ej: 110, 22 o 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_slowdns_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Descargando binarios e instalando SlowDNS / Noiz DNS... (Tomará unos segundos)</i>", tele.ModeHTML)

		pubKey, err := vpn.InstallSlowDNS(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar SlowDNS / Noiz DNS:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>SlowDNS / Noiz DNS Instalado Correctamente</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>NS:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("🔑 <b>Pub Key:</b> <code>%s</code>\n", pubKey)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>El servicio ya está activo en Systemd.</i>"

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
			b.Edit(lastMsg, "⚠️ <b>Ese dominio ya está en uso</b> por otro protocolo (SlowDNS / Noiz DNS o Slipstream).\nPor favor, ingresa un dominio (NS) distinto para VayDNS:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_vaydns_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>¿A qué puerto local quieres redirigir VayDNS? (Ej: 110, 22 o 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_vaydns_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Descargando binarios e instalando VayDNS... (Tomará unos segundos)</i>", tele.ModeHTML)

		pubKey, err := vpn.InstallVayDNS(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar VayDNS:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>VayDNS Instalado Correctamente</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>NS:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("🔑 <b>Pub Key:</b> <code>%s</code>\n", pubKey)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>El servicio ya está activo en Systemd.</i>"

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
			b.Edit(lastMsg, "⚠️ <b>Ese dominio ya está en uso</b> por otro protocolo (SlowDNS / Noiz DNS o VayDNS).\nPor favor, ingresa un dominio distinto para Slipstream:", markupCancel, tele.ModeHTML)
			return nil
		}
		SetTempValue(chatID, "domain", text)
		SetUserStep(chatID, "awaiting_vpn_slipstream_port")

		markupCancel := &tele.ReplyMarkup{}
		markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
		b.Edit(lastMsg, "⚙️ <i>¿A qué puerto local quieres redirigir Slipstream? (Ej: 110, 22 o 443):</i>", markupCancel, tele.ModeHTML)
		return nil

	case "awaiting_vpn_slipstream_port":
		domain := GetTempValue(chatID, "domain")
		port := text

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Descargando binarios y configurando TLS para Slipstream... (Tomará unos segundos)</i>", tele.ModeHTML)

		err := vpn.InstallSlipstream(domain, port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar Slipstream:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Slipstream Instalado Correctamente</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌍 <b>Dominio:</b> <code>%s</code>\n", domain)
		res += fmt.Sprintf("⚙️ <b>Puerto:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>El servicio QUIC ya está activo en UDP 53.</i>"

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
			b.Edit(lastMsg, "❌ <b>Puerto inválido.</b> Por favor, ingresa solo números (Ej: 7300).", markup, tele.ModeHTML)
			return nil
		}
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Instalando ZiVPN (UDP Custom)...</i>", tele.ModeHTML)

		err := vpn.InstallZivpn(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar ZiVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>ZiVPN Instalado Correctamente</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("⚙️ <b>Puerto UDP:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>El servicio udp-custom ya está activo.</i>"

		// Guardar estado
		data, _ := db.Load()
		data.Zivpn = true
		db.Save(data)

		b.Edit(lastMsg, res, markup, tele.ModeHTML)
		return nil

	case "awaiting_vpn_badvpn_port":
		port := text
		if _, err := strconv.Atoi(port); err != nil {
			b.Edit(lastMsg, "❌ <b>Puerto inválido.</b> Por favor, ingresa solo números (Ej: 7200).", markup, tele.ModeHTML)
			return nil
		}
		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Descargando e instalando BadVPN...</i>", tele.ModeHTML)

		err := vpn.InstallBadVPN(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar BadVPN:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>BadVPN Instalado Correctamente</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("⚙️ <b>Puerto TCP:</b> <code>%s</code>\n", port)
		res += "━━━━━━━━━━━━━━\n"
		res += "<i>El demonio udpgw ya está escuchando.</i>"

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
			b.Edit(lastMsg, "❌ <b>Puerto en uso por HAProxy (SSL Tunnel).</b>\n\nPor favor, ingresa un puerto diferente:", markupCancel, tele.ModeHTML)
			return nil
		}
		if data.SSHWebSocket && (port == "10015" || port == "2082") {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Puerto en uso por SSH WebSocket.</b>\n\nPor favor, ingresa un puerto diferente:", markupCancel, tele.ModeHTML)
			return nil
		}

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Instalando Falcon Proxy...</i>", tele.ModeHTML)
		ver, err := vpn.InstallFalcon(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar Falcon:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Falcon Proxy Instalado</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🦅 <b>Version:</b> <code>%s</code>\n", ver)
		res += fmt.Sprintf("⚙️ <b>Puerto:</b> <code>%s</code>\n", port)
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

		b.Edit(lastMsg, "⏳ <i>Configurando SSL Tunnel (HAProxy)...</i>", tele.ModeHTML)
		err := vpn.InstallSSLTunnel(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar SSL Tunnel:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>SSL Tunnel Instalado</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("📜 <b>Puerto SSL:</b> <code>%s</code>\n", port)
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

		b.Edit(lastMsg, "⏳ <i>Configurando Dropbear (multi-puerto)...</i>", tele.ModeHTML)
		err := vpn.InstallDropbear(ports)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar Dropbear:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>Dropbear Instalado (Multi-Puerto)</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🐻 <b>Puertos:</b> <code>%s</code>\n", ports)
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
			b.Edit(lastMsg, "❌ <b>Puerto inválido.</b> Por favor, ingresa solo números (Ej: 8080).", markup, tele.ModeHTML)
			return nil
		}

		data, _ := db.Load()
		if data.SSLTunnel != "" && (port == "80" || port == "443" || port == "8080" || port == data.SSLTunnel) {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Puerto en uso por HAProxy (SSL Tunnel).</b>\n\nPor favor, ingresa un puerto diferente:", markupCancel, tele.ModeHTML)
			return nil
		}
		if data.SSHWebSocket && (port == "10015" || port == "2082") {
			markupCancel := &tele.ReplyMarkup{}
			markupCancel.Inline(markupCancel.Row(markupCancel.Data("❌ Cancelar", "cancelar_accion")))
			b.Edit(lastMsg, "❌ <b>Puerto en uso por SSH WebSocket.</b>\n\nPor favor, ingresa un puerto diferente:", markupCancel, tele.ModeHTML)
			return nil
		}

		DeleteUserStep(chatID)

		b.Edit(lastMsg, "⏳ <i>Instalando y configurando ProxyDT...</i>", tele.ModeHTML)

		if err := vpn.InstallProxyDT(); err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al instalar binario ProxyDT:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		err := vpn.OpenProxyDTPort(port)
		if err != nil {
			b.Edit(lastMsg, fmt.Sprintf("❌ <b>Error al abrir puerto ProxyDT:</b>\n<pre>%v</pre>", err), markup, tele.ModeHTML)
			return nil
		}

		res := "✅ <b>ProxyDT Online</b>\n"
		res += "━━━━━━━━━━━━━━\n"
		res += fmt.Sprintf("🌐 <b>Puerto:</b> <code>%s</code>\n", port)
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
