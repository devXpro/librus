package message

import (
	"librus/model"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// FallbackHandler handles messages that don't match any other handler
type FallbackHandler struct{}

// Handle processes fallback messages
func (h *FallbackHandler) Handle(ctx *router.Context) error {
	// If user is not authenticated, show login prompt
	if ctx.User == nil {
		return ctx.SendMessage(localization.MsgPleaseLogin)
	}

	// If user is authenticated, show helpful message with main menu
	if ctx.User.State == model.StateAuthenticated {
		mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
		return ctx.SendMessageWithKeyboard(localization.MsgUnknownCommand, mainMenu)
	}

	// For other states, show appropriate message
	return ctx.SendMessage(localization.MsgUnknownCommand)
}
