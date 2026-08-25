package bot

import (
	"fmt"
	"os"
	"strings"

	"github.com/orxma/depwise/internal/i18n"
	"github.com/orxma/depwise/internal/sys"
	tele "gopkg.in/telebot.v3"
)

func handleMenuScanner(c tele.Context, b *tele.Bot) error {
	assetOK, httpxOK := sys.GetScannerStatus()

	markup := &tele.ReplyMarkup{}
	btnBack := markup.Data(i18n.T(c.Chat().ID, "btn.back"), "back_main")

	if !assetOK || !httpxOK {
		// Herramientas no instaladas
		markup.Inline(markup.Row(btnBack))
		texto := i18n.T(c.Chat().ID, "scanner.no_tools_title")
		if !assetOK {
			texto += i18n.Tf(c.Chat().ID, "scanner.tool_not_installed", "assetfinder")
		} else {
			texto += i18n.Tf(c.Chat().ID, "scanner.tool_installed", "assetfinder")
		}
		if !httpxOK {
			texto += i18n.Tf(c.Chat().ID, "scanner.tool_not_installed", "httpx")
		} else {
			texto += i18n.Tf(c.Chat().ID, "scanner.tool_installed", "httpx")
		}
		texto += i18n.T(c.Chat().ID, "scanner.install_hint")
		return SafeEditCtx(c, b, texto, markup)
	}

	btnStart := markup.Data(i18n.T(c.Chat().ID, "scanner.btn_start"), "start_scanner_prompt")
	markup.Inline(markup.Row(btnStart), markup.Row(btnBack))

	texto := i18n.T(c.Chat().ID, "scanner.ready_title")

	return SafeEditCtx(c, b, texto, markup)
}

func handleStartScanPrompt(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	SetUserStep(chatID, "awaiting_scanner_domain")
	SetLastBotMsg(chatID, c.Message())

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.cancel"), "menu_scanner")))

	return c.Edit(i18n.T(chatID, "scanner.prompt_domain"), markup, tele.ModeHTML)
}

func processScannerSteps(step string, text string, chatID int64, c tele.Context, b *tele.Bot, lastMsg *tele.Message) error {
	if step != "awaiting_scanner_domain" {
		return nil
	}

	domain := strings.TrimSpace(text)
	DeleteUserStep(chatID)

	// Verificar herramientas antes de escanear
	assetOK, httpxOK := sys.GetScannerStatus()
	if !assetOK || !httpxOK {
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "menu_scanner")))
		b.Edit(lastMsg, i18n.T(chatID, "scanner.not_installed_alert"), markup, tele.ModeHTML)
		return nil
	}

	b.Edit(lastMsg, i18n.Tf(chatID, "scanner.scanning", domain), tele.ModeHTML)

	go func() {
		result, err := sys.RunScanner(domain)
		markup := &tele.ReplyMarkup{}
		markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "scanner.btn_back_scanner"), "menu_scanner")))

		if err != nil {
			b.Edit(lastMsg, i18n.Tf(chatID, "scanner.error", err), markup, tele.ModeHTML)
			return
		}

		// Limitar el resultado si es muy largo para Telegram (4096 chars)
		if len(result) > 3500 {
			// Enviar primero un adelanto
			preview := result[:3000] + i18n.T(chatID, "scanner.preview_footer")
			header := i18n.Tf(chatID, "scanner.preview_header", domain)
			b.Edit(lastMsg, header+preview, markup, tele.ModeHTML)

			// Crear archivo temporal para el reporte completo
			tmpFile := fmt.Sprintf("/tmp/scan_%s.txt", domain)
			_ = os.WriteFile(tmpFile, []byte(result), 0644)
			defer os.Remove(tmpFile)

			doc := &tele.Document{
				File:     tele.FromDisk(tmpFile),
				FileName: i18n.Tf(chatID, "scanner.file_name", domain),
				Caption:  i18n.Tf(chatID, "scanner.file_caption", domain),
			}
			b.Send(c.Sender(), doc)
			return
		}

		header := i18n.Tf(chatID, "scanner.success_header", domain)
		b.Edit(lastMsg, header+result, markup, tele.ModeHTML)
	}()

	return nil
}

// handleSubMenuScanner muestra el submenú de gestión del escáner en Protocolos
func handleSubMenuScanner(c tele.Context, b *tele.Bot) error {
	assetOK, httpxOK := sys.GetScannerStatus()

	installed := 0
	total := 2
	if assetOK {
		installed++
	}
	if httpxOK {
		installed++
	}

	pct := (installed * 100) / total

	statusAsset := i18n.T(c.Chat().ID, "scanner.status_not_installed")
	if assetOK {
		statusAsset = i18n.T(c.Chat().ID, "scanner.status_installed")
	}
	statusHttpx := i18n.T(c.Chat().ID, "scanner.status_not_installed")
	if httpxOK {
		statusHttpx = i18n.T(c.Chat().ID, "scanner.status_installed")
	}

	globalStatus := i18n.T(c.Chat().ID, "scanner.status_not_installed")
	if installed == total {
		globalStatus = i18n.T(c.Chat().ID, "scanner.global_complete")
	} else if installed > 0 {
		globalStatus = i18n.T(c.Chat().ID, "scanner.global_partial")
	}

	markup := &tele.ReplyMarkup{}
	btnInstall := markup.Data(i18n.T(c.Chat().ID, "scanner.btn_install_all"), "install_scanner_all")
	btnUninstall := markup.Data(i18n.T(c.Chat().ID, "scanner.btn_uninstall_all"), "uninstall_scanner_all")
	btnBack := markup.Data(i18n.T(c.Chat().ID, "btn.back"), "menu_protocols")

	markup.Inline(
		markup.Row(btnInstall),
		markup.Row(btnUninstall),
		markup.Row(btnBack),
	)

	// Barra de progreso visual
	barFull := installed
	barEmpty := total - installed
	bar := strings.Repeat("█", barFull*5) + strings.Repeat("░", barEmpty*5)

	texto := i18n.Tf(c.Chat().ID, "scanner.manage_title", globalStatus, bar, pct, installed, total, statusAsset, statusHttpx)

	return SafeEditCtx(c, b, texto, markup)
}

// handleInstallScannerAll instala todas las herramientas del escáner con progreso
func handleInstallScannerAll(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "scanner.admin_only"), ShowAlert: true})
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "submenu_scanner")))

	go func() {
		lastMsg := c.Message()
		assetOK, httpxOK := sys.GetScannerStatus()
		totalSteps := 0
		if !assetOK {
			totalSteps++
		}
		if !httpxOK {
			totalSteps++
		}

		if totalSteps == 0 {
			b.Edit(lastMsg, i18n.T(chatID, "scanner.already_installed"), markup, tele.ModeHTML)
			return
		}

		currentStep := 0

		// Instalar assetfinder
		if !assetOK {
			currentStep++
			pct := (currentStep * 100) / (totalSteps + 1)
			bar := strings.Repeat("█", pct/10) + strings.Repeat("░", 10-pct/10)
			b.Edit(lastMsg, i18n.Tf(chatID, "scanner.installing_asset", bar, pct), tele.ModeHTML)

			if err := sys.InstallScannerTool("assetfinder"); err != nil {
				b.Edit(lastMsg, i18n.Tf(chatID, "scanner.err_asset", err), markup, tele.ModeHTML)
				return
			}
		}

		// Instalar httpx
		if !httpxOK {
			currentStep++
			pct := (currentStep * 100) / (totalSteps + 1)
			bar := strings.Repeat("█", pct/10) + strings.Repeat("░", 10-pct/10)
			b.Edit(lastMsg, i18n.Tf(chatID, "scanner.installing_httpx", bar, pct), tele.ModeHTML)

			if err := sys.InstallScannerTool("httpx"); err != nil {
				b.Edit(lastMsg, i18n.Tf(chatID, "scanner.err_httpx", err), markup, tele.ModeHTML)
				return
			}
		}

		// Finalizado
		b.Edit(lastMsg, i18n.T(chatID, "scanner.installed_all"), markup, tele.ModeHTML)
	}()

	return nil
}

// handleUninstallScannerAll desinstala todas las herramientas
func handleUninstallScannerAll(c tele.Context, b *tele.Bot) error {
	chatID := c.Chat().ID
	if !isFullAdmin(chatID) {
		return c.Respond(&tele.CallbackResponse{Text: i18n.T(chatID, "scanner.admin_only"), ShowAlert: true})
	}

	SafeEditCtx(c, b, i18n.T(chatID, "scanner.uninstalling"), nil)

	err := sys.UninstallAllScannerTools()
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(markup.Data(i18n.T(chatID, "btn.back"), "submenu_scanner")))

	if err != nil {
		return SafeEditCtx(c, b, i18n.Tf(chatID, "scanner.err_partial", err), markup)
	}
	return SafeEditCtx(c, b, i18n.T(chatID, "scanner.uninstalled_all"), markup)
}
