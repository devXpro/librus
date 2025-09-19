package keyboard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/telegram/localization"
)

// Callback data constants
const (
	// Language selection
	CallbackLanguageSelect = "lang"

	// Main menu
	CallbackCheckMessages = "check_messages"
	CallbackGetByURL      = "get_by_url"
	CallbackSettings      = "settings"
	CallbackHelp          = "help"

	// Settings
	CallbackChangeLanguage = "change_language"
	CallbackResetAccount   = "reset_account"
	CallbackBackToMenu     = "back_to_menu"

	// Authentication
	CallbackLogin    = "login"
	CallbackContinue = "continue"

	// Confirmation
	CallbackConfirmYes = "confirm_yes"
	CallbackConfirmNo  = "confirm_no"
	CallbackCancel     = "cancel"
)

// LanguageSelectionKeyboard creates a keyboard for language selection
func LanguageSelectionKeyboard() tgbotapi.InlineKeyboardMarkup {
	languages := localization.GetSupportedLanguages()
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create rows with 2 languages per row
	for i := 0; i < len(languages); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// First language in row
		langCode := languages[i]
		langName := localization.GetLanguageName(langCode)
		button := tgbotapi.NewInlineKeyboardButtonData(langName, CallbackLanguageSelect+":"+langCode)
		row = append(row, button)

		// Second language in row (if exists)
		if i+1 < len(languages) {
			langCode2 := languages[i+1]
			langName2 := localization.GetLanguageName(langCode2)
			button2 := tgbotapi.NewInlineKeyboardButtonData(langName2, CallbackLanguageSelect+":"+langCode2)
			row = append(row, button2)
		}

		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// LoginKeyboard creates a keyboard for login prompt
func LoginKeyboard(localizer *localization.Localizer) tgbotapi.InlineKeyboardMarkup {
	loginButton := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgPleaseLogin),
		CallbackLogin,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(loginButton),
	)
}

// MainMenuKeyboard creates the main menu keyboard for authenticated users
func MainMenuKeyboard(localizer *localization.Localizer) tgbotapi.InlineKeyboardMarkup {
	checkMessagesBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgCheckMessages),
		CallbackCheckMessages,
	)

	getByURLBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgGetMessageByURL),
		CallbackGetByURL,
	)

	settingsBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgSettings),
		CallbackSettings,
	)

	helpBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgHelp),
		CallbackHelp,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(checkMessagesBtn),
		tgbotapi.NewInlineKeyboardRow(getByURLBtn),
		tgbotapi.NewInlineKeyboardRow(settingsBtn),
		tgbotapi.NewInlineKeyboardRow(helpBtn),
	)
}

// SettingsKeyboard creates the settings menu keyboard
func SettingsKeyboard(localizer *localization.Localizer) tgbotapi.InlineKeyboardMarkup {
	changeLanguageBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgChangeLanguage),
		CallbackChangeLanguage,
	)

	resetAccountBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgResetAccount),
		CallbackResetAccount,
	)

	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgBackToMenu),
		CallbackBackToMenu,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(changeLanguageBtn),
		tgbotapi.NewInlineKeyboardRow(resetAccountBtn),
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)
}

// ConfirmationKeyboard creates a yes/no confirmation keyboard
func ConfirmationKeyboard(localizer *localization.Localizer, yesCallback, noCallback string) tgbotapi.InlineKeyboardMarkup {
	yesBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgYes),
		yesCallback,
	)

	noBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgNo),
		noCallback,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yesBtn, noBtn),
	)
}

// CancelKeyboard creates a simple cancel keyboard
func CancelKeyboard(localizer *localization.Localizer) tgbotapi.InlineKeyboardMarkup {
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgCancel),
		CallbackCancel,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(cancelBtn),
	)
}

// BackToMenuKeyboard creates a simple back to menu keyboard
func BackToMenuKeyboard(localizer *localization.Localizer) tgbotapi.InlineKeyboardMarkup {
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		localizer.GetMessage(localization.MsgBackToMenu),
		CallbackBackToMenu,
	)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)
}
