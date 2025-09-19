package callback

import (
	"log"

	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// SettingsHandler handles settings-related callbacks
type SettingsHandler struct{}

// Handle processes settings callbacks
func (h *SettingsHandler) Handle(ctx *router.Context) error {
	switch ctx.Update.Data {
	case keyboard.CallbackChangeLanguage:
		return h.handleChangeLanguage(ctx)
	case keyboard.CallbackResetAccount:
		return h.handleResetAccount(ctx)
	default:
		log.Printf("Unknown settings callback: %s", ctx.Update.Data)
		return nil
	}
}

// handleChangeLanguage shows language selection menu
func (h *SettingsHandler) handleChangeLanguage(ctx *router.Context) error {
	languageKeyboard := keyboard.LanguageSelectionKeyboard()
	return ctx.EditMessageWithKeyboard(localization.MsgSelectLanguage, languageKeyboard)
}

// handleResetAccount shows confirmation for account reset
func (h *SettingsHandler) handleResetAccount(ctx *router.Context) error {
	confirmKeyboard := keyboard.ConfirmationKeyboard(
		ctx.Localization,
		keyboard.CallbackConfirmYes+":reset",
		keyboard.CallbackConfirmNo,
	)
	return ctx.EditMessageWithKeyboard(localization.MsgAreYouSure, confirmKeyboard)
}
