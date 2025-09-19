package callback

import (
	"log"
	"sort"

	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// MenuHandler handles main menu callbacks
type MenuHandler struct{}

// Handle processes main menu callbacks
func (h *MenuHandler) Handle(ctx *router.Context) error {
	switch ctx.Update.Data {
	case keyboard.CallbackCheckMessages:
		return h.handleCheckMessages(ctx)
	case keyboard.CallbackGetByURL:
		return h.handleGetByURL(ctx)
	case keyboard.CallbackSettings:
		return h.handleSettings(ctx)
	case keyboard.CallbackHelp:
		return h.handleHelp(ctx)
	case keyboard.CallbackBackToMenu:
		return h.handleBackToMenu(ctx)
	default:
		log.Printf("Unknown menu callback: %s", ctx.Update.Data)
		return nil
	}
}

// handleCheckMessages checks for new messages
func (h *MenuHandler) handleCheckMessages(ctx *router.Context) error {
	// Check if user is authenticated
	if ctx.User == nil || ctx.User.State != model.StateAuthenticated || ctx.User.Login == "" || ctx.User.Password == "" {
		loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
	}

	// Show processing message
	err := ctx.EditMessage(localization.MsgProcessing)
	if err != nil {
		return err
	}

	// Create gRPC client
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return ctx.EditMessage(localization.MsgServiceError)
	}
	defer client.Close()

	// Get all updates
	msgs, news, err := client.GetAllUpdates(ctx.User.Login, ctx.User.Password)
	if err != nil {
		log.Printf("Failed to get updates for user %s: %v", ctx.User.Login, err)
		return ctx.EditMessage(localization.MsgServiceError)
	}

	// Combine messages and news
	allMsgs := append(msgs, news...)
	if len(allMsgs) == 0 {
		backKeyboard := keyboard.BackToMenuKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgNoNewMessages, backKeyboard)
	}

	// Add user ID to messages
	for i := range allMsgs {
		allMsgs[i].UserID = ctx.User.Id
	}

	// Add to database (only new ones will be added)
	allMsgs, err = mongo.AddMessagesToDatabase(allMsgs, ctx.User.Id)
	if err != nil {
		log.Printf("Error adding messages to database: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Sort by date
	sort.Slice(allMsgs, func(i, j int) bool {
		return allMsgs[i].Date.Before(allMsgs[j].Date)
	})

	// Send messages
	for _, message := range allMsgs {
		if ctx.User.Language != "" {
			message.Translate(ctx.User.Language)
		}

		err = message.Send(ctx.Bot, ctx.Update.ChatID)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		// Clean up attachments
		if err := message.CleanupAttachments(); err != nil {
			log.Printf("Error cleaning up attachments: %v", err)
		}
	}

	// Show completion message with menu
	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgCheckComplete, mainMenu)
}

// handleGetByURL prompts user to send URL
func (h *MenuHandler) handleGetByURL(ctx *router.Context) error {
	// Check if user is authenticated
	if ctx.User == nil || ctx.User.State != model.StateAuthenticated || ctx.User.Login == "" || ctx.User.Password == "" {
		loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
	}

	// Update user state to awaiting URL
	err := mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAwaitingURL)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	cancelKeyboard := keyboard.CancelKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgEnterURL, cancelKeyboard)
}

// handleSettings shows settings menu
func (h *MenuHandler) handleSettings(ctx *router.Context) error {
	settingsMenu := keyboard.SettingsKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgSettingsMenu, settingsMenu)
}

// handleHelp shows help information
func (h *MenuHandler) handleHelp(ctx *router.Context) error {
	backKeyboard := keyboard.BackToMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgHelpText, backKeyboard)
}

// handleBackToMenu returns to main menu
func (h *MenuHandler) handleBackToMenu(ctx *router.Context) error {
	// Reset user state to authenticated
	err := mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAuthenticated)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
	}

	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgMainMenu, mainMenu)
}
