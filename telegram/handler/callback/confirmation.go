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

// ConfirmationHandler handles confirmation callbacks
type ConfirmationHandler struct{}

// Handle processes confirmation callbacks
func (h *ConfirmationHandler) Handle(ctx *router.Context) error {
	// Parse callback data
	if strings.HasPrefix(ctx.Update.Data, keyboard.CallbackConfirmYes+":") {
		action := strings.TrimPrefix(ctx.Update.Data, keyboard.CallbackConfirmYes+":")
		return h.handleConfirmYes(ctx, action)
	} else if ctx.Update.Data == keyboard.CallbackConfirmNo {
		return h.handleConfirmNo(ctx)
	} else if ctx.Update.Data == keyboard.CallbackCancel {
		return h.handleCancel(ctx)
	}

	log.Printf("Unknown confirmation callback: %s", ctx.Update.Data)
	return nil
}

// handleConfirmYes processes positive confirmations
func (h *ConfirmationHandler) handleConfirmYes(ctx *router.Context, action string) error {
	switch action {
	case "reset":
		return h.handleResetConfirmed(ctx)
	default:
		log.Printf("Unknown confirmation action: %s", action)
		return h.handleCancel(ctx)
	}
}

// handleConfirmNo processes negative confirmations
func (h *ConfirmationHandler) handleConfirmNo(ctx *router.Context) error {
	// Return to settings menu
	settingsMenu := keyboard.SettingsKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgSettingsMenu, settingsMenu)
}

// handleCancel cancels current operation and returns to appropriate menu
func (h *ConfirmationHandler) handleCancel(ctx *router.Context) error {
	// Reset telegram user state to authenticated
	err := mongo.UpdateTelegramUserState(ctx.Update.ChatID, model.StateAuthenticated)
	if err != nil {
		log.Printf("Error updating telegram user state: %v", err)
	}

	// Return to main menu
	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgMainMenu, mainMenu)
}

// handleResetConfirmed processes confirmed account reset
func (h *ConfirmationHandler) handleResetConfirmed(ctx *router.Context) error {
	// Delete telegram user from database
	err := mongo.DeleteTelegramUserByTelegramID(ctx.Update.ChatID)
	if err != nil {
		log.Printf("Error deleting telegram user: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Show reset confirmation and start over
	err = ctx.EditMessage(localization.MsgResetComplete)
	if err != nil {
		return err
	}

	// Create new telegram user with language selection state
	err = mongo.CreateTelegramUserWithState(ctx.Update.ChatID, model.StateLanguageSelection)
	if err != nil {
		log.Printf("Error creating new telegram user: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	// Show language selection
	languageKeyboard := keyboard.LanguageSelectionKeyboard()
	return ctx.SendMessageWithKeyboard(localization.MsgSelectLanguage, languageKeyboard)
}
