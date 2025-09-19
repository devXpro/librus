package message

import (
	"librus/helper"
	"librus/mongo"
	"librus/pkg/logger"
	"librus/telegram/channel"
	"librus/telegram/localization"
	"librus/telegram/router"

	"go.uber.org/zap"
)

// AdminHandler handles admin commands with token authentication
type AdminHandler struct{}

// Handle processes admin messages
func (h *AdminHandler) Handle(ctx *router.Context) error {
	message := ctx.Update.Data

	// Check for admin commands with token authentication
	if message == "update_now_"+helper.GetEnv("TELEGRAM_TOKEN", "pass") {
		return h.handleUpdateNow(ctx)
	}

	if message == "delete_all_messages_"+helper.GetEnv("TELEGRAM_TOKEN", "pass") {
		return h.handleDeleteAllMessages(ctx)
	}

	return nil // Not an admin command, let other handlers process it
}

// handleUpdateNow triggers immediate message check
func (h *AdminHandler) handleUpdateNow(ctx *router.Context) error {
	channel.UpdateNow <- struct{}{}
	return nil // Don't send any response for security
}

// handleDeleteAllMessages deletes all messages from database
func (h *AdminHandler) handleDeleteAllMessages(ctx *router.Context) error {
	err := mongo.DeleteAllMessages()
	if err != nil {
		logger.ErrorWithError("Error deleting messages", err,
			zap.Int64("admin_user_id", ctx.User.TelegramID),
		)
		return ctx.SendMessage(localization.MsgServiceError)
	}
	return ctx.SendMessage(localization.MsgCheckComplete)
}
