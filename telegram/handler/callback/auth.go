package callback

import (
	"log"

	"librus/model"
	"librus/mongo"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// AuthHandler handles authentication-related callbacks
type AuthHandler struct{}

// Handle processes authentication callbacks
func (h *AuthHandler) Handle(ctx *router.Context) error {
	switch ctx.Update.Data {
	case keyboard.CallbackLogin:
		return h.handleLoginStart(ctx)
	default:
		log.Printf("Unknown auth callback: %s", ctx.Update.Data)
		return nil
	}
}

// handleLoginStart initiates the login process
func (h *AuthHandler) handleLoginStart(ctx *router.Context) error {
	// Update user state to awaiting login
	err := mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAwaitingLogin)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Show login prompt with cancel option
	cancelKeyboard := keyboard.CancelKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgEnterLogin, cancelKeyboard)
}
