package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/orxma/depwise/internal/db"
	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	"github.com/orxma/depwise/internal/vpn"
	tele "gopkg.in/telebot.v3"
)

var (
	botToken   = os.Getenv("BOT_TOKEN")
	superAdmin = os.Getenv("SUPER_ADMIN")

	// Estado Global de Conversación (Sincronizado)
	stateMu    sync.RWMutex
	UserSteps  = make(map[int64]string)
	TempData   = make(map[int64]map[string]string)
	LastBotMsg = make(map[int64]*tele.Message)
)

func GetUserStep(chatID int64) string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return UserSteps[chatID]
}

func GetUserStepWithOk(chatID int64) (string, bool) {
	stateMu.RLock()
	defer stateMu.RUnlock()
	step, ok := UserSteps[chatID]
	return step, ok
}

func SetUserStep(chatID int64, step string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	UserSteps[chatID] = step
}

func DeleteUserStep(chatID int64) {
	stateMu.Lock()
	defer stateMu.Unlock()
	delete(UserSteps, chatID)
	delete(TempData, chatID)
	delete(LastBotMsg, chatID)
}

func GetTempData(chatID int64) map[string]string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return TempData[chatID]
}

func SetTempData(chatID int64, data map[string]string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	TempData[chatID] = data
}

func GetTempValue(chatID int64, key string) string {
	stateMu.RLock()
	defer stateMu.RUnlock()
	if TempData[chatID] == nil {
		return ""
	}
	return TempData[chatID][key]
}

func SetTempValue(chatID int64, key, value string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if TempData[chatID] == nil {
		TempData[chatID] = make(map[string]string)
	}
	TempData[chatID][key] = value
}

func DeleteTempValue(chatID int64, key string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if TempData[chatID] != nil {
		delete(TempData[chatID], key)
	}
}

func GetLastBotMsg(chatID int64) *tele.Message {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return LastBotMsg[chatID]
}

func SetLastBotMsg(chatID int64, msg *tele.Message) {
	stateMu.Lock()
	defer stateMu.Unlock()
	LastBotMsg[chatID] = msg
}

// StartBot inicializa el bot de Telegram y registra los handlers
func StartBot() {
	if botToken == "" || superAdmin == "" {
		log.Fatal("BOT_TOKEN and SUPER_ADMIN variables are required")
	}

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Middleware de Baneo
	b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender() != nil {
				chatID := fmt.Sprintf("%d", c.Sender().ID)
				data, errLoad := db.Load()
				if errLoad == nil {
					if info, banned := data.BannedUsers[chatID]; banned {
						// Ignorar si es SuperAdmin (protección extra)
						if !isSuperAdminID(c.Sender().ID) {
							if c.Callback() != nil {
								return c.Respond(&tele.CallbackResponse{
									Text:      i18n.Tf(c.Sender().ID, "mid.banned_alert", info.Reason),
									ShowAlert: true,
								})
							}
							return c.Send(i18n.Tf(c.Sender().ID, "mid.banned_msg", info.Reason), tele.ModeHTML)
						}
					}
				}
			}
			return next(c)
		}
	})

	// Handlers
	b.Handle("/start", func(c tele.Context) error {
		return handleStart(c, b)
	})

	b.Handle("/referidos", func(c tele.Context) error {
		return handleMenuReferrals(c, b)
	})

	b.Handle("/menu", func(c tele.Context) error {
		return handleStart(c, b)
	})

	// Text Interceptor para conversacion
	b.Handle(tele.OnText, func(c tele.Context) error {
		return handleTextInputs(c, b)
	})

	// Document Interceptor para restaurar backup
	b.Handle(tele.OnDocument, func(c tele.Context) error {
		step, ok := GetUserStepWithOk(c.Chat().ID)
		if ok && step == "awaiting_backup_restore" {
			return handleRestoreDocument(c, b)
		}
		return nil
	})


	// Opciones del Menú Principal
	b.Handle(&tele.Btn{Unique: "menu_crear"}, func(c tele.Context) error {
		return SafeEditCtx(c, b, menuCrearText(c.Chat().ID), menuCrearMarkup(c.Chat().ID))
	})
	b.Handle(&tele.Btn{Unique: "menu_info"}, func(c tele.Context) error {
		return handleInfo(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_broadcast"}, func(c tele.Context) error {
		return handleMenuBroadcast(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_scanner"}, func(c tele.Context) error {
		return handleMenuScanner(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_eliminar"}, func(c tele.Context) error {
		return handleMenuEliminar(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_referrals"}, func(c tele.Context) error {
		return handleMenuReferrals(c, b)
	})

	// Opciones de Configuración Avanzada
	b.Handle(&tele.Btn{Unique: "menu_info_cuenta"}, func(c tele.Context) error {
		return handleMenuInfoCuenta(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_editar"}, func(c tele.Context) error {
		return handleMenuEditar(c, b)
	})
	b.Handle(&tele.Btn{Unique: "edit_pass"}, func(c tele.Context) error {
		return HandleEditPass(c, b)
	})
	b.Handle(&tele.Btn{Unique: "edit_renew"}, func(c tele.Context) error {
		return HandleEditRenew(c, b)
	})
	b.Handle(&tele.Btn{Unique: "edit_limit"}, func(c tele.Context) error {
		return HandleEditLimit(c, b)
	})

	b.Handle(&tele.Btn{Unique: "renew_ssh"}, func(c tele.Context) error {
		return HandleAutoRenew(c, b, "ssh")
	})
	b.Handle(&tele.Btn{Unique: "renew_zi"}, func(c tele.Context) error {
		return HandleAutoRenew(c, b, "zi")
	})
	b.Handle(&tele.Btn{Unique: "renew_xray"}, func(c tele.Context) error {
		return HandleAutoRenew(c, b, "xray")
	})

	b.Handle(&tele.Btn{Unique: "menu_protocols"}, func(c tele.Context) error {
		return handleMenuProtocols(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_admins"}, func(c tele.Context) error {
		return handleMenuAdmins(c, b)
	})
	b.Handle(&tele.Btn{Unique: "menu_online"}, func(c tele.Context) error {
		return handleMenuOnline(c, b)
	})

	// VPNs
	b.Handle(&tele.Btn{Unique: "install_slowdns"}, func(c tele.Context) error {
		return handleInstallSlowDNS(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_vaydns"}, func(c tele.Context) error {
		return handleInstallVayDNS(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_zivpn"}, func(c tele.Context) error {
		return handleInstallZivpn(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_badvpn"}, func(c tele.Context) error {
		return handleInstallBadVPN(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_falcon"}, func(c tele.Context) error {
		return handleInstallFalcon(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_ssl"}, func(c tele.Context) error {
		return handleInstallSSL(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_dropbear"}, func(c tele.Context) error {
		return handleInstallDropbear(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_proxydt"}, func(c tele.Context) error {
		return handleInstallProxyDT(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_udpcustom"}, func(c tele.Context) error {
		return handleInstallUDPCustom(c, b)
	})
	b.Handle(&tele.Btn{Unique: "install_scanner_deps"}, func(c tele.Context) error {
		return handleInstallScannerAll(c, b)
	})
	b.Handle(&tele.Btn{Unique: "submenu_scanner"}, func(c tele.Context) error {
		return handleSubMenuScanner(c, b)
	})
	b.Handle(&tele.Btn{Unique: "install_scanner_all"}, func(c tele.Context) error {
		return handleInstallScannerAll(c, b)
	})
	b.Handle(&tele.Btn{Unique: "uninstall_scanner_all"}, func(c tele.Context) error {
		return handleUninstallScannerAll(c, b)
	})
	b.Handle(&tele.Btn{Unique: "install_xray"}, func(c tele.Context) error {
		return handleInstallXray(c, b, c.Message())
	})
	b.Handle(&tele.Btn{Unique: "install_slipstream"}, func(c tele.Context) error {
		return handleInstallSlipstream(c, b, c.Message())
	})

	// Sub-Menús de Protocolos
	b.Handle(&tele.Btn{Unique: "submenu_slowdns"}, func(c tele.Context) error { return handleSubMenuSlowDNS(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_vaydns"}, func(c tele.Context) error { return handleSubMenuVayDNS(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_zivpn"}, func(c tele.Context) error { return handleSubMenuZiVPN(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_badvpn"}, func(c tele.Context) error { return handleSubMenuBadVPN(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_falcon"}, func(c tele.Context) error { return handleSubMenuFalcon(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_ssl"}, func(c tele.Context) error { return handleSubMenuSSL(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_dropbear"}, func(c tele.Context) error { return handleSubMenuDropbear(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_proxydt"}, func(c tele.Context) error { return handleSubMenuProxyDT(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_udpcustom"}, func(c tele.Context) error { return handleSubMenuUDPCustom(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_xray"}, func(c tele.Context) error { return handleSubMenuXray(c, b) })
	b.Handle(&tele.Btn{Unique: "submenu_slipstream"}, func(c tele.Context) error { return handleSubMenuSlipstream(c, b) })
	b.Handle(&tele.Btn{Unique: "manage_xray_users"}, func(c tele.Context) error { return handleManageXrayUsers(c, b) })
	b.Handle(&tele.Btn{Unique: "protocol_diag"}, func(c tele.Context) error { return handleProtocolDiag(c, b) })
	b.Handle(&tele.Btn{Unique: "menu_protocols"}, func(c tele.Context) error { return handleMenuProtocols(c, b) })

	// Desinstaladores
	b.Handle(&tele.Btn{Unique: "uninstall_slowdns"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "SlowDNS") })
	b.Handle(&tele.Btn{Unique: "uninstall_vaydns"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "VayDNS") })
	b.Handle(&tele.Btn{Unique: "uninstall_zivpn"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "ZiVPN") })
	b.Handle(&tele.Btn{Unique: "uninstall_badvpn"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "BadVPN") })
	b.Handle(&tele.Btn{Unique: "uninstall_falcon"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "Falcon") })
	b.Handle(&tele.Btn{Unique: "uninstall_ssl"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "SSL Tunnel") })
	b.Handle(&tele.Btn{Unique: "uninstall_dropbear"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "Dropbear") })
	b.Handle(&tele.Btn{Unique: "uninstall_proxydt"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "ProxyDT") })
	b.Handle(&tele.Btn{Unique: "uninstall_udpcustom"}, func(c tele.Context) error { return handleUninstallUDPCustom(c, b) })
	b.Handle(&tele.Btn{Unique: "uninstall_xray"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "Xray") })
	b.Handle(&tele.Btn{Unique: "uninstall_slipstream"}, func(c tele.Context) error { return handleUninstallProtocol(c, b, "Slipstream") })

	// Callbacks Dinámicos (One-Tap Selection)
	b.Handle("\fed_user:", func(c tele.Context) error { return handleEditSelection(c, b) })
	b.Handle("\fdel_confirm:", func(c tele.Context) error { return handleDeleteSelection(c, b) })
	b.Handle("\fdel_adm_exec", func(c tele.Context) error { return handleDelAdminExec(c, b) })
	b.Handle("\fdel_xray_exec", func(c tele.Context) error { return handleDeleteXrayExec(c, b) })
	b.Handle("\frename_adm_sel", func(c tele.Context) error { return handleRenameAdminSelect(c, b) })
	b.Handle("\funban_user", func(c tele.Context) error { return handleUnbanUser(c, b) })
	b.Handle(&tele.Btn{Unique: "rename_admin_menu"}, func(c tele.Context) error { return handleRenameAdminMenu(c, b) })

	// Ajustes Pro
	b.Handle(&tele.Btn{Unique: "toggle_public_access"}, func(c tele.Context) error { return handleTogglePublicAccess(c, b) })
	b.Handle(&tele.Btn{Unique: "list_admins"}, func(c tele.Context) error { return handleListAdmins(c, b) })
	b.Handle(&tele.Btn{Unique: "add_admin"}, func(c tele.Context) error { return handleAddAdminPrompt(c, b) })
	b.Handle(&tele.Btn{Unique: "add_admin_normal"}, func(c tele.Context) error { return handleAdminAccessType(c, b, false) })
	b.Handle(&tele.Btn{Unique: "add_admin_full"}, func(c tele.Context) error { return handleAdminAccessType(c, b, true) })
	b.Handle(&tele.Btn{Unique: "del_admin_menu"}, func(c tele.Context) error { return handleDelAdminMenu(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_extrainfo"}, func(c tele.Context) error { return handleEditExtraInfoPrompt(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_cloudflare"}, func(c tele.Context) error { return handleEditCloudflarePrompt(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_cloudfront"}, func(c tele.Context) error { return handleEditCloudfrontPrompt(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_banner"}, func(c tele.Context) error { return handleEditBannerPrompt(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_promo_menu"}, func(c tele.Context) error { return handleEditPromoMenu(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_promo_text"}, func(c tele.Context) error { return handleEditPromoText(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_promo_channel"}, func(c tele.Context) error { return handleEditPromoChannel(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_promo_support"}, func(c tele.Context) error { return handleEditPromoSupport(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_promo_botname"}, func(c tele.Context) error { return handleEditPromoBotName(c, b) })
	b.Handle(&tele.Btn{Unique: "banner_set_custom"}, func(c tele.Context) error { return handleBannerSetCustom(c, b) })
	b.Handle(&tele.Btn{Unique: "banner_deactivate"}, func(c tele.Context) error { return handleBannerDeactivate(c, b) })
	b.Handle(&tele.Btn{Unique: "edit_quotas"}, func(c tele.Context) error { return handleEditQuotas(c, b) })
	b.Handle(&tele.Btn{Unique: "quota_days_public"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_days_public", "Maximum days for public users")
	})
	b.Handle(&tele.Btn{Unique: "quota_limit_public"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_limit_public", "Maximum devices for public users")
	})
	b.Handle(&tele.Btn{Unique: "quota_days_admin"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_days_admin", "Maximum days for Admins")
	})
	b.Handle(&tele.Btn{Unique: "quota_limit_admin"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_limit_admin", "Maximum devices for Admins")
	})
	b.Handle(&tele.Btn{Unique: "quota_xray_public"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_xray_public", "Max VMess accounts for Public")
	})
	b.Handle(&tele.Btn{Unique: "quota_xray_admin"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_xray_admin", "Max VMess accounts for Admins")
	})
	b.Handle(&tele.Btn{Unique: "quota_ssh_public"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_ssh_public", "Max SSH account limit (Public)")
	})
	b.Handle(&tele.Btn{Unique: "quota_ssh_admin"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_ssh_admin", "Max SSH account limit (Admins)")
	})
	b.Handle(&tele.Btn{Unique: "quota_zivpn_public"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_zivpn_public", "Max ZiVPN account limit (Public)")
	})
	b.Handle(&tele.Btn{Unique: "quota_zivpn_admin"}, func(c tele.Context) error {
		return handleQuotaPrompt(c, b, "awaiting_quota_zivpn_admin", "Max ZiVPN account limit (Admins)")
	})
	b.Handle(&tele.Btn{Unique: "reset_history"}, func(c tele.Context) error { return handleResetHistoryConfirm(c, b) })
	b.Handle(&tele.Btn{Unique: "reset_history_exec"}, func(c tele.Context) error { return handleResetHistoryExec(c, b) })
	b.Handle(&tele.Btn{Unique: "reboot_vps_confirm"}, func(c tele.Context) error { return handleServerRebootConfirm(c, b) })
	b.Handle(&tele.Btn{Unique: "reboot_vps_exec"}, func(c tele.Context) error { return handleServerRebootExec(c, b) })
	b.Handle(&tele.Btn{Unique: "toggle_public_scanner"}, func(c tele.Context) error { return handleTogglePublicScanner(c, b) })
	b.Handle(&tele.Btn{Unique: "toggle_monetization"}, func(c tele.Context) error { return handleToggleMonetization(c, b) })
	b.Handle(&tele.Btn{Unique: "menu_config_ads"}, func(c tele.Context) error { return handleMenuAdsConfig(c, b) })
	b.Handle(&tele.Btn{Unique: "menu_autoreboot"}, func(c tele.Context) error { return handleAutoRebootMenu(c, b) })
	b.Handle(&tele.Btn{Unique: "toggle_autoreboot"}, func(c tele.Context) error { return handleToggleAutoReboot(c, b) })
	b.Handle(&tele.Btn{Unique: "menu_bans"}, func(c tele.Context) error { return handleMenuBans(c, b) })

	// Updater
	b.Handle(&tele.Btn{Unique: "menu_updater"}, func(c tele.Context) error { return handleMenuUpdater(c, b) })
	b.Handle(&tele.Btn{Unique: "updater_check"}, func(c tele.Context) error { return handleUpdaterCheck(c, b) })
	b.Handle(&tele.Btn{Unique: "updater_run"}, func(c tele.Context) error { return handleUpdaterRun(c, b) })
	b.Handle(&tele.Btn{Unique: "updater_toggle_auto"}, func(c tele.Context) error { return handleUpdaterToggleAuto(c, b) })
	b.Handle(&tele.Btn{Unique: "ban_user_prompt"}, func(c tele.Context) error { return handleBanUserPrompt(c, b) })

	// Referidos
	registerReferralHandlers(b)

	// Telegram Backups
	b.Handle(&tele.Btn{Unique: "menu_backup"}, func(c tele.Context) error { return handleBackupMenu(c, b) })
	b.Handle(&tele.Btn{Unique: "backup_now"}, func(c tele.Context) error { return handleLocalBackup(c, b) })
	b.Handle(&tele.Btn{Unique: "backup_auto_1"}, func(c tele.Context) error { return handleSetBackupInterval(c, b, 1) })
	b.Handle(&tele.Btn{Unique: "backup_auto_3"}, func(c tele.Context) error { return handleSetBackupInterval(c, b, 3) })
	b.Handle(&tele.Btn{Unique: "backup_auto_7"}, func(c tele.Context) error { return handleSetBackupInterval(c, b, 7) })
	b.Handle(&tele.Btn{Unique: "backup_auto_30"}, func(c tele.Context) error { return handleSetBackupInterval(c, b, 30) })
	b.Handle(&tele.Btn{Unique: "backup_auto_0"}, func(c tele.Context) error { return handleSetBackupInterval(c, b, 0) })
	b.Handle(&tele.Btn{Unique: "restore_req"}, func(c tele.Context) error { return handleLocalRestoreReq(c, b) })

	// Generar Usuario SSH / ZIVPN Handler
	b.Handle(&tele.Btn{Unique: "crear_ssh"}, func(c tele.Context) error {
		return handleCrearSSH(c, b)
	})
	b.Handle(&tele.Btn{Unique: "crear_zivpn"}, func(c tele.Context) error {
		return handleCrearZivpn(c, b)
	})
	b.Handle(&tele.Btn{Unique: "crear_xray"}, func(c tele.Context) error {
		return handleCrearXray(c, b)
	})
	b.Handle(&tele.Btn{Unique: "ssh_rnd_pass"}, func(c tele.Context) error {
		return handleRandomPass(c, b)
	})
	b.Handle(&tele.Btn{Unique: "ssh_default_title"}, func(c tele.Context) error {
		return handleDefaultTitle(c, b)
	})
	b.Handle(&tele.Btn{Unique: "cancelar_accion"}, func(c tele.Context) error {
		return handleCancel(c, b)
	})

	b.Handle(&tele.Btn{Unique: "back_main"}, func(c tele.Context) error {
		return handleStart(c, b) // Vuelve al inicio redibujando o editando
	})

	b.Handle(&tele.Btn{Unique: "start_scanner_prompt"}, func(c tele.Context) error {
		return handleStartScanPrompt(c, b)
	})

	// Parchar config de Xray existente para habilitar access log y configurar resiliencia
	if initData, _ := db.Load(); initData.Xray.Installed {
		if err := vpn.EnsureXrayAccessLog(); err != nil {
			log.Printf("Warning: could not enable Xray access log: %v", err)
		}
		if err := vpn.EnsureXrayServiceResilience(); err != nil {
			log.Printf("Warning: could not ensure Xray service resilience: %v", err)
		}
	}

	// Restaurar reglas de iptables que se borran al reiniciar (SlowDNS, ZiVPN)
	vpn.RestoreIptablesRules()

	// Verificar y reiniciar HAProxy si quedó caído tras un reboot del VPS
	if initSSL, _ := db.Load(); initSSL.SSLTunnel != "" {
		vpn.EnsureHAProxyRunning()
		log.Println("HAProxy: verified and restored successfully")
	}

	// Restaurar contraseñas ZiVPN en config.json tras reinicio de VPS
	if initZivpn, _ := db.Load(); initZivpn.Zivpn && len(initZivpn.ZivpnUsers) > 0 {
		var passwords []string
		for pass := range initZivpn.ZivpnUsers {
			passwords = append(passwords, pass)
		}
		if err := vpn.RestoreZivpnPasswords(passwords); err != nil {
			log.Printf("Warning: Could not restore ZiVPN passwords: %v", err)
		} else {
			log.Printf("ZiVPN: %d passwords synchronized with config.json", len(passwords))
		}
	}

	// Instalar sistema de banners individuales por usuario SSH
	if err := sys.EnsureBannerSystem(); err != nil {
		log.Printf("Warning: could not initialize the banner system: %v", err)
	}
	// Regenerar todos los banners existentes al iniciar
	go sys.RefreshAllBanners()

	// Iniciar hilo de auto-limpieza (Rutina concurrente)
	go sys.AutoCleanupLoop(b)

	// Iniciar hilo de respaldos automáticos por Telegram
	go autoBackupLoop(b)

	// Iniciar hilo de notificaciones de expiración
	go autoExpirationAlertLoop(b)

	log.Println("Bot started successfully...")
	b.Start()
}

func isAdmin(chatID int64) bool {
	if isSuperAdminID(chatID) {
		return true
	}
	data, _ := db.Load()
	_, exists := data.Admins[fmt.Sprintf("%d", chatID)]
	return exists
}

func isSuperAdminID(chatID int64) bool {
	sa, _ := strconv.ParseInt(superAdmin, 10, 64)
	return chatID == sa
}

func isFullAdmin(chatID int64) bool {
	if isSuperAdminID(chatID) {
		return true
	}
	data, _ := db.Load()
	if info, exists := data.Admins[fmt.Sprintf("%d", chatID)]; exists {
		return info.FullAccess
	}
	return false
}

// SafeEdit intenta editar un mensaje, y si falla lo envía nuevo
func SafeEdit(chatID int64, b *tele.Bot, msg *tele.Message, text string, markup *tele.ReplyMarkup) (*tele.Message, error) {
	var newMsg *tele.Message
	var err error

	if msg != nil {
		newMsg, err = b.Edit(msg, text, markup, tele.ModeHTML)
	} else {
		err = fmt.Errorf("nil message")
	}

	if err != nil {
		// Fallback: tratar de borrar el viejo para no dejar spam
		if msg != nil {
			b.Delete(msg)
		}
		// Enviar nuevo
		newMsg, err = b.Send(tele.ChatID(chatID), text, markup, tele.ModeHTML)
	}

	if err == nil {
		SetLastBotMsg(chatID, newMsg)
	}
	return newMsg, err
}

// SafeEditCtx es un helper que facilita el uso de SafeEdit con tele.Context
func SafeEditCtx(c tele.Context, b *tele.Bot, text string, markup *tele.ReplyMarkup) error {
	var lastMsg *tele.Message
	if c.Callback() != nil {
		lastMsg = c.Message()
	} else {
		lastMsg = GetLastBotMsg(c.Chat().ID)
	}

		_, err := SafeEdit(c.Chat().ID, b, lastMsg, text, markup)
	return err
}

func handleStart(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	payload := c.Message().Payload

	if payload == "adcompleted" || payload == "aderror" {
		_, ok := GetUserStepWithOk(chatID)
		if ok {
			action := "ad_completed"
			if payload == "aderror" {
				action = "ad_error"
			}
			return ProcessAdCompletion(c, b, action)
		}
	}

	if strings.HasPrefix(payload, "ref_") {
		referrerStr := strings.TrimPrefix(payload, "ref_")
		referrerID, err := strconv.ParseInt(referrerStr, 10, 64)
		if err == nil && referrerID != chatID {
			data, _ := db.Load()
			if data.ReferredBy[chatID] == 0 && !data.ReferralCompleted[chatID] {
				db.Update(func(d *db.ConfigData) error {
					if d.ReferredBy == nil {
						d.ReferredBy = make(map[int64]int64)
					}
					d.ReferredBy[chatID] = referrerID
					return nil
				})
			}
		}
	}

	// Limpiar cualquier estado activo al volver al menú
	DeleteUserStep(chatID)

	data, _ := db.Load()

	// Registrar historial
	found := false
	for _, id := range data.UserHistory {
		if id == chatID {
			found = true
			break
		}
	}
	if !found {
		data.UserHistory = append(data.UserHistory, chatID)
		db.Save(data)
	}

	// Comprobar Acceso Público
	if !data.PublicAccess && !isAdmin(chatID) {
		textoDenegado := i18n.T(chatID, "main.private_system")

		if c.Callback() != nil {
			return c.Edit(textoDenegado, tele.ModeHTML)
		}
		return c.Send(textoDenegado, tele.ModeHTML)
	}

	// Mostrar Menú Principal
	textoMenu := buildMainMenuText(chatID, data)
	markup := buildMainMenuMarkup(chatID)

	var msg *tele.Message
	var err error
	if c.Callback() != nil {
		msg, err = b.Edit(c.Message(), textoMenu, markup, tele.ModeHTML)
	} else {
		msg, err = b.Send(c.Chat(), textoMenu, markup, tele.ModeHTML)
	}

	if err == nil {
		SetLastBotMsg(chatID, msg)
	}
	return err
}

func buildMainMenuText(chatID int64, data *db.ConfigData) string {
	texto := i18n.T(chatID, "main.title")
	texto += i18n.T(chatID, "main.subtitle")

	stats := sys.GetSystemStats()

	// CPU Formatter
	barraCPU := sys.GenerarBarra(stats.CPUUsage, 100.0, 10)
	texto += i18n.Tf(chatID, "main.cpu", barraCPU, stats.CPUUsage, stats.Cores)

	// RAM Formatter
	barraRAM := sys.GenerarBarra(float64(stats.RAMUsed), float64(stats.RAMTotal), 10)
	texto += i18n.Tf(chatID, "main.ram", barraRAM, stats.RAMUsed, stats.RAMTotal)

	// Disco
	barraDisk := sys.GenerarBarra(float64(stats.DiskUsed), float64(stats.DiskTotal), 10)
	texto += i18n.Tf(chatID, "main.disk", barraDisk, stats.DiskUsed, stats.DiskTotal)

	texto += i18n.Tf(chatID, "main.uptime", stats.UptimeStr)

	if !data.PublicAccess {
		texto += i18n.T(chatID, "main.private_off")
	}
	return texto
}

func buildMainMenuMarkup(chatID int64) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	btnCrear := menu.Data(i18n.T(chatID, "btn.create_ssh"), "menu_crear")
	btnInfo := menu.Data(i18n.T(chatID, "btn.server_info"), "menu_info")
	btnEditar := menu.Data(i18n.T(chatID, "btn.edit_ssh"), "menu_editar")
	btnDelete := menu.Data(i18n.T(chatID, "btn.delete_ssh"), "menu_eliminar")
	btnInfoCuenta := menu.Data(i18n.T(chatID, "btn.account_info"), "menu_info_cuenta")
	btnGlobal := menu.Data(i18n.T(chatID, "btn.broadcast"), "menu_broadcast")
	btnScanner := menu.Data(i18n.T(chatID, "btn.scanner"), "menu_scanner")
	btnOnline := menu.Data(i18n.T(chatID, "btn.monitor"), "menu_online")
	btnProtocols := menu.Data(i18n.T(chatID, "btn.protocols"), "menu_protocols")
	btnSettings := menu.Data(i18n.T(chatID, "btn.pro_settings"), "menu_admins")
	btnReferrals := menu.Data(i18n.T(chatID, "btn.referrals"), "menu_referrals")

	data, _ := db.Load()
	isFull := isFullAdmin(chatID)
	isAdm := isAdmin(chatID)

	// Construir filas dinámicamente
	var rows []tele.Row

	// Fila 1: Crear e Info
	rows = append(rows, menu.Row(btnCrear, btnInfo))

	// Fila 2: Scanner (Always for Admins, conditional for Public)
	if isFull || isAdm || data.PublicScanner {
		rows = append(rows, menu.Row(btnScanner))
	}

	// Fila 3: Editar y Online
	if isFull || isAdm {
		rows = append(rows, menu.Row(btnEditar, btnOnline))
	} else {
		rows = append(rows, menu.Row(btnOnline))
	}

	// Fila 4: Info Cuenta y Eliminar
	rows = append(rows, menu.Row(btnInfoCuenta, btnDelete))

	// Fila 4.5: Referidos
	rows = append(rows, menu.Row(btnReferrals))

	// Fila 6: SuperAdmin / Admin Config
	if isFull {
		rows = append(rows, menu.Row(btnGlobal, btnProtocols))
		rows = append(rows, menu.Row(btnSettings))
	}

	// Asignar filas al menú
	menu.Inline(rows...)

	return menu
}

func menuCrearText(chatID int64) string {
	return i18n.T(chatID, "create.title")
}

func menuCrearMarkup(chatID int64) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	btnSSH := menu.Data(i18n.T(chatID, "btn.ssh_client"), "crear_ssh")
	btnZivpn := menu.Data(i18n.T(chatID, "btn.zivpn_access"), "crear_zivpn")
	btnXray := menu.Data(i18n.T(chatID, "btn.vmess_xray"), "crear_xray")
	btnBack := menu.Data(i18n.T(chatID, "btn.back"), "back_main")

	data, _ := db.Load()
	var rows []tele.Row
	rows = append(rows, menu.Row(btnSSH))
	rows = append(rows, menu.Row(btnZivpn))
	if data.Xray.Installed {
		rows = append(rows, menu.Row(btnXray))
	}
	rows = append(rows, menu.Row(btnBack))

	menu.Inline(rows...)
	return menu
}


