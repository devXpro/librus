package callback

import (
	"log"
	"strings"

	"librus/model"
	"librus/mongo"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// LanguageHandler handles language selection callbacks
type LanguageHandler struct{}

// Handle processes language selection callback
func (h *LanguageHandler) Handle(ctx *router.Context) error {
	// Extract language code from callback data (format: "lang:en")
	parts := strings.Split(ctx.Update.Data, ":")
	if len(parts) != 2 {
		log.Printf("Invalid language callback data: %s", ctx.Update.Data)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	langCode := parts[1]

	// Validate language code
	supportedLangs := localization.GetSupportedLanguages()
	isSupported := false
	for _, supported := range supportedLangs {
		if supported == langCode {
			isSupported = true
			break
		}
	}

	if !isSupported {
		log.Printf("Unsupported language code: %s", langCode)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Update user language
	err := mongo.UpdateUserLanguageByTelegramID(ctx.Update.ChatID, langCode)
	if err != nil {
		log.Printf("Error updating user language: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Update user state to awaiting login
	err = mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAwaitingLogin)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Update context localization with new language
	ctx.Localization = localization.NewLocalizer(langCode)

	// Show language confirmation and what the bot does
	err = ctx.EditMessage(localization.MsgLanguageSet)
	if err != nil {
		return err
	}

	// Send explanation about the bot
	err = ctx.SendMessage(localization.MsgWhatIsLibrusBot)
	if err != nil {
		return err
	}

	// Show login prompt
	loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
	return ctx.SendMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
}
