package callback

import (
	"sort"

	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/pkg/logger"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"

	"go.uber.org/zap"
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
		logger.Warn("Unknown menu callback",
			zap.String("callback_data", ctx.Update.Data),
			zap.Int64("user_id", ctx.User.TelegramID),
		)
		return nil
	}
}

// handleCheckMessages checks for new messages
func (h *MenuHandler) handleCheckMessages(ctx *router.Context) error {
	// Check if telegram user is authenticated and has Librus account
	if ctx.User == nil || ctx.User.State != model.StateAuthenticated || ctx.User.LibrusLogin == "" {
		loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
	}

	// Get Librus account
	librusAccount, err := mongo.FindLibrusAccount(ctx.User.LibrusLogin)
	if err != nil || librusAccount == nil {
		loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
	}

	// Show processing message
	err = ctx.EditMessage(localization.MsgProcessing)
	if err != nil {
		return err
	}

	// Create gRPC client
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		logger.ErrorWithError("Failed to create gRPC client", err,
			zap.Int64("user_id", ctx.User.TelegramID),
		)
		return ctx.EditMessage(localization.MsgServiceError)
	}
	defer client.Close()

	// Get all updates
	msgs, news, err := client.GetAllUpdates(librusAccount.Login, librusAccount.Password)
	if err != nil {
		logger.ErrorWithError("Failed to get updates for account", err,
			zap.String("login", librusAccount.Login),
			zap.Int64("user_id", ctx.User.TelegramID),
		)
		return ctx.EditMessage(localization.MsgServiceError)
	}

	// Combine messages and news
	allMsgs := append(msgs, news...)
	if len(allMsgs) == 0 {
		backKeyboard := keyboard.BackToMenuKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgNoNewMessages, backKeyboard)
	}

	// Add to database (only new ones will be added)
	err = mongo.AddMessagesToDatabase(allMsgs)
	if err != nil {
		logger.ErrorWithError("Error adding messages to database", err,
			zap.String("librus_login", ctx.User.LibrusLogin),
			zap.Int64("user_id", ctx.User.TelegramID),
		)
		return ctx.EditMessage(localization.MsgSomethingWrong)
	}

	// Sort by date
	sort.Slice(allMsgs, func(i, j int) bool {
		return allMsgs[i].Date.Before(allMsgs[j].Date)
	})

	// Send messages
	for _, message := range allMsgs {
		// Check if message was already sent to this user
		if mongo.IsMessageSentToUser(ctx.User.Id, message.Id) {
			continue
		}

		// Translate message if user has language preference
		translatedMessage := message
		if ctx.User.Language != "" {
			translatedMessage.Translate(ctx.User.Language)
		}

		err = translatedMessage.Send(ctx.Bot, ctx.Update.ChatID)
		if err != nil {
			logger.ErrorWithError("Error sending message", err,
				zap.Int64("chat_id", ctx.Update.ChatID),
				zap.String("message_id", message.Id),
				zap.String("message_type", string(message.Type)),
			)
			continue
		}

		// Mark message as sent
		err = mongo.MarkMessageAsSent(ctx.User.Id, message.Id)
		if err != nil {
			logger.ErrorWithError("Error marking message as sent", err,
				zap.String("user_id", ctx.User.Id),
				zap.String("message_id", message.Id),
				zap.String("message_type", string(message.Type)),
			)
		}

		// Clean up attachments
		if err := message.CleanupAttachments(); err != nil {
			logger.ErrorWithError("Error cleaning up attachments", err,
				zap.String("message_id", message.Id),
				zap.String("attachments_dir", message.AttachmentsDir),
			)
		}
	}

	// Show completion message with menu
	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgCheckComplete, mainMenu)
}

// handleGetByURL prompts user to send URL
func (h *MenuHandler) handleGetByURL(ctx *router.Context) error {
	// Check if telegram user is authenticated and has Librus account
	if ctx.User == nil || ctx.User.State != model.StateAuthenticated || ctx.User.LibrusLogin == "" {
		loginKeyboard := keyboard.LoginKeyboard(ctx.Localization)
		return ctx.EditMessageWithKeyboard(localization.MsgPleaseLogin, loginKeyboard)
	}

	// Update telegram user state to awaiting URL
	err := mongo.UpdateTelegramUserState(ctx.Update.ChatID, model.StateAwaitingURL)
	if err != nil {
		logger.ErrorWithError("Error updating telegram user state", err,
			zap.Int64("chat_id", ctx.Update.ChatID),
			zap.String("state", string(model.StateAwaitingURL)),
		)
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
	// Reset telegram user state to authenticated
	err := mongo.UpdateTelegramUserState(ctx.Update.ChatID, model.StateAuthenticated)
	if err != nil {
		logger.ErrorWithError("Error updating telegram user state", err,
			zap.Int64("chat_id", ctx.Update.ChatID),
			zap.String("state", string(model.StateAuthenticated)),
		)
	}

	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.EditMessageWithKeyboard(localization.MsgMainMenu, mainMenu)
}
