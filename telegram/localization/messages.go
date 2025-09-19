package localization

import (
	"fmt"
	"librus/translator"
	"log"
)

// MessageKey represents a localization key
type MessageKey string

// Predefined message keys
const (
	// Welcome and onboarding
	MsgWelcome         MessageKey = "welcome"
	MsgSelectLanguage  MessageKey = "select_language"
	MsgLanguageSet     MessageKey = "language_set"
	MsgWhatIsLibrusBot MessageKey = "what_is_librus_bot"

	// Authentication
	MsgPleaseLogin        MessageKey = "please_login"
	MsgEnterLogin         MessageKey = "enter_login"
	MsgEnterPassword      MessageKey = "enter_password"
	MsgLoginSuccess       MessageKey = "login_success"
	MsgLoginFailed        MessageKey = "login_failed"
	MsgInvalidCredentials MessageKey = "invalid_credentials"

	// Main menu
	MsgMainMenu        MessageKey = "main_menu"
	MsgCheckMessages   MessageKey = "check_messages"
	MsgGetMessageByURL MessageKey = "get_message_by_url"
	MsgSettings        MessageKey = "settings"
	MsgHelp            MessageKey = "help"

	// Settings
	MsgSettingsMenu   MessageKey = "settings_menu"
	MsgChangeLanguage MessageKey = "change_language"
	MsgResetAccount   MessageKey = "reset_account"
	MsgBackToMenu     MessageKey = "back_to_menu"

	// Messages
	MsgProcessing    MessageKey = "processing"
	MsgMessageSent   MessageKey = "message_sent"
	MsgCannotReply   MessageKey = "cannot_reply"
	MsgInvalidURL    MessageKey = "invalid_url"
	MsgServiceError  MessageKey = "service_error"
	MsgNoNewMessages MessageKey = "no_new_messages"
	MsgCheckComplete MessageKey = "check_complete"
	MsgEnterURL      MessageKey = "enter_url"
	MsgHelpText      MessageKey = "help_text"

	// Errors
	MsgUnknownCommand MessageKey = "unknown_command"
	MsgSomethingWrong MessageKey = "something_wrong"

	// Confirmation
	MsgAreYouSure    MessageKey = "are_you_sure"
	MsgYes           MessageKey = "yes"
	MsgNo            MessageKey = "no"
	MsgCancel        MessageKey = "cancel"
	MsgResetComplete MessageKey = "reset_complete"
)

// Default messages in English
var defaultMessages = map[MessageKey]string{
	// Welcome and onboarding
	MsgWelcome:         "🎉 Welcome to Librus Bot!\n\nThis bot helps you receive messages and news from Librus (Polish school system) directly in Telegram.\n\nFirst, please select your language:",
	MsgSelectLanguage:  "🌍 Please select your language:",
	MsgLanguageSet:     "✅ Language has been set! Now all messages will be in your selected language.",
	MsgWhatIsLibrusBot: "📚 Librus Bot connects to your Librus account and:\n\n📩 Forwards personal messages\n🔔 Sends news and announcements\n💬 Allows you to reply to messages\n🌍 Translates everything to your language\n\nTo get started, please log in with your Librus credentials:",

	// Authentication
	MsgPleaseLogin:        "🔐 Log in to Librus",
	MsgEnterLogin:         "👤 Please enter your Librus login (username):",
	MsgEnterPassword:      "🔑 Please enter your Librus password:",
	MsgLoginSuccess:       "✅ Successfully logged in! You will now receive messages from Librus.",
	MsgLoginFailed:        "❌ Login failed. Please check your credentials and try again.",
	MsgInvalidCredentials: "❌ Invalid login or password. Please try again.",

	// Main menu
	MsgMainMenu:        "📋 Main Menu\n\nChoose what you'd like to do:",
	MsgCheckMessages:   "📬 Check for new messages",
	MsgGetMessageByURL: "🔗 Get message by URL",
	MsgSettings:        "⚙️ Settings",
	MsgHelp:            "ℹ️ Help",

	// Settings
	MsgSettingsMenu:   "⚙️ Settings\n\nWhat would you like to change?",
	MsgChangeLanguage: "🌍 Change language",
	MsgResetAccount:   "🔄 Reset account",
	MsgBackToMenu:     "⬅️ Back to menu",

	// Messages
	MsgProcessing:    "⏳ Processing your request...",
	MsgMessageSent:   "✅ Message sent successfully!",
	MsgCannotReply:   "❌ Cannot reply to this type of message.",
	MsgInvalidURL:    "❌ Invalid URL format. Please provide a valid Librus URL.",
	MsgServiceError:  "❌ Service connection error. Please try again later.",
	MsgNoNewMessages: "📭 No new messages found.",
	MsgCheckComplete: "✅ Message check completed!",
	MsgEnterURL:      "🔗 Please send the Librus message URL:",
	MsgHelpText:      "ℹ️ **Librus Bot Help**\n\n📬 **Check Messages** - Get new messages from Librus\n🔗 **Get by URL** - Fetch specific message by URL\n💬 **Reply** - Reply to any message by replying to it\n⚙️ **Settings** - Change language or reset account\n\n🌍 All messages are automatically translated to your language.",

	// Errors
	MsgUnknownCommand: "❓ Unknown command. Use the menu buttons or type /help for assistance.",
	MsgSomethingWrong: "❌ Something went wrong. Please try again.",

	// Confirmation
	MsgAreYouSure:    "❓ Are you sure?",
	MsgYes:           "✅ Yes",
	MsgNo:            "❌ No",
	MsgCancel:        "❌ Cancel",
	MsgResetComplete: "✅ Account reset completed! Let's start over.",
}

// Localizer provides localized messages
type Localizer struct {
	language string
}

// NewLocalizer creates a new localizer for the given language
func NewLocalizer(language string) *Localizer {
	return &Localizer{
		language: language,
	}
}

// GetMessage returns a localized message for the given key
func (l *Localizer) GetMessage(key MessageKey, args ...interface{}) string {
	// Get default message
	defaultMsg, exists := defaultMessages[key]
	if !exists {
		log.Printf("Missing message key: %s", key)
		return string(key)
	}

	// Format with arguments if provided
	message := defaultMsg
	if len(args) > 0 {
		message = fmt.Sprintf(defaultMsg, args...)
	}

	// Translate if language is not English
	if l.language != "" && l.language != "en" {
		translated, err := translator.TranslateText(l.language, message)
		if err != nil {
			log.Printf("Translation error for key %s: %v", key, err)
			return message // Return original if translation fails
		}
		return translated
	}

	return message
}

// GetLanguageName returns the display name for a language code
func GetLanguageName(langCode string) string {
	languageNames := map[string]string{
		"en": "🇬🇧 English",
		"pl": "🇵🇱 Polski",
		"uk": "🇺🇦 Українська",
		"de": "🇩🇪 Deutsch",
		"fr": "🇫🇷 Français",
		"es": "🇪🇸 Español",
		"it": "🇮🇹 Italiano",
		"ru": "🇷🇺 Русский",
	}

	if name, exists := languageNames[langCode]; exists {
		return name
	}
	return langCode
}

// GetSupportedLanguages returns a list of supported language codes
func GetSupportedLanguages() []string {
	return []string{"en", "pl", "uk", "de", "fr", "es", "it", "ru"}
}
