package router

import (
	"librus/model"
	"librus/telegram/localization"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UpdateType represents the type of Telegram update
type UpdateType int

const (
	UpdateTypeMessage UpdateType = iota
	UpdateTypeCallback
	UpdateTypeCommand
)

// Update represents a unified Telegram update
type Update struct {
	ChatID    int64
	UserID    int64
	Type      UpdateType
	Data      string
	MessageID int
	Original  tgbotapi.Update
}

// Context provides all necessary data for handlers
type Context struct {
	Bot          *tgbotapi.BotAPI
	Update       *Update
	User         *model.TelegramUser
	Localization *localization.Localizer
}

// CommandHandler handles slash commands like /start
type CommandHandler interface {
	Handle(ctx *Context) error
}

// CallbackHandler handles inline keyboard button presses
type CallbackHandler interface {
	Handle(ctx *Context) error
}

// StateHandler handles messages based on user state
type StateHandler interface {
	Handle(ctx *Context) error
}

// MessageHandler handles regular text messages
type MessageHandler interface {
	Handle(ctx *Context) error
}

// NewUpdateFromTelegram converts tgbotapi.Update to our Update type
func NewUpdateFromTelegram(tgUpdate tgbotapi.Update) *Update {
	update := &Update{
		Original: tgUpdate,
	}

	if tgUpdate.Message != nil {
		update.ChatID = tgUpdate.Message.Chat.ID
		update.UserID = tgUpdate.Message.From.ID
		update.Data = tgUpdate.Message.Text

		// Determine if it's a command or regular message
		if tgUpdate.Message.IsCommand() {
			update.Type = UpdateTypeCommand
		} else {
			update.Type = UpdateTypeMessage
		}
	} else if tgUpdate.CallbackQuery != nil {
		update.ChatID = tgUpdate.CallbackQuery.Message.Chat.ID
		update.UserID = tgUpdate.CallbackQuery.From.ID
		update.Data = tgUpdate.CallbackQuery.Data
		update.MessageID = tgUpdate.CallbackQuery.Message.MessageID
		update.Type = UpdateTypeCallback
	}

	return update
}
