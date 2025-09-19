package command

import (
	"log"

	"librus/model"
	"librus/mongo"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// StartHandler handles the /start command
type StartHandler struct{}

// Handle processes the /start command
func (h *StartHandler) Handle(ctx *router.Context) error {
	// Check if telegram user exists
	if ctx.User == nil {
		// Create new telegram user with language selection state
		err := mongo.CreateTelegramUserWithState(ctx.Update.ChatID, model.StateLanguageSelection)
		if err != nil {
			log.Printf("Error creating telegram user: %v", err)
			return ctx.SendMessage(localization.MsgSomethingWrong)
		}

		// Reload telegram user
		telegramUser, err := mongo.FindTelegramUserByTelegramID(ctx.Update.ChatID)
		if err != nil {
			log.Printf("Error reloading telegram user: %v", err)
			return ctx.SendMessage(localization.MsgSomethingWrong)
		}
		ctx.User = telegramUser
	}

	// Handle based on user state
	switch ctx.User.State {
	case model.StateLanguageSelection:
		return h.handleLanguageSelection(ctx)
	case model.StateAwaitingLogin, model.StateAwaitingPassword:
		return h.handleAuthenticationFlow(ctx)
	case model.StateAuthenticated:
		return h.handleAuthenticatedUser(ctx)
	default:
		// Reset to language selection if unknown state
		err := mongo.UpdateTelegramUserState(ctx.Update.ChatID, model.StateLanguageSelection)
		if err != nil {
			log.Printf("Error updating telegram user state: %v", err)
		}
		return h.handleLanguageSelection(ctx)
	}
}

// handleLanguageSelection shows language selection menu
func (h *StartHandler) handleLanguageSelection(ctx *router.Context) error {
	keyboard := keyboard.LanguageSelectionKeyboard()
	return ctx.SendMessageWithKeyboard(localization.MsgWelcome, keyboard)
}

// handleAuthenticationFlow shows login prompt
func (h *StartHandler) handleAuthenticationFlow(ctx *router.Context) error {
	// Show what the bot does and prompt for login
	err := ctx.SendMessage(localization.MsgWhatIsLibrusBot)
	if err != nil {
		return err
	}

	loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
	return ctx.SendMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
}

// handleAuthenticatedUser shows main menu
func (h *StartHandler) handleAuthenticatedUser(ctx *router.Context) error {
	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.SendMessageWithKeyboard(localization.MsgMainMenu, mainMenu)
}
