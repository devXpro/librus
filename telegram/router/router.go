package router

import (
	"log"
	"strings"

	"librus/model"
	"librus/mongo"
	"librus/telegram/localization"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Router handles routing of updates to appropriate handlers
type Router struct {
	commandHandlers  map[string]CommandHandler
	callbackHandlers map[string]CallbackHandler
	stateHandlers    map[model.UserState]StateHandler
	messageHandlers  []MessageHandler
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		commandHandlers:  make(map[string]CommandHandler),
		callbackHandlers: make(map[string]CallbackHandler),
		stateHandlers:    make(map[model.UserState]StateHandler),
		messageHandlers:  make([]MessageHandler, 0),
	}
}

// RegisterCommand registers a command handler
func (r *Router) RegisterCommand(command string, handler CommandHandler) {
	r.commandHandlers[command] = handler
}

// RegisterCallback registers a callback handler
func (r *Router) RegisterCallback(callbackData string, handler CallbackHandler) {
	r.callbackHandlers[callbackData] = handler
}

// RegisterState registers a state handler
func (r *Router) RegisterState(state model.UserState, handler StateHandler) {
	r.stateHandlers[state] = handler
}

// RegisterMessage registers a message handler
func (r *Router) RegisterMessage(handler MessageHandler) {
	r.messageHandlers = append(r.messageHandlers, handler)
}

// HandleUpdate processes a Telegram update
func (r *Router) HandleUpdate(bot *tgbotapi.BotAPI, tgUpdate tgbotapi.Update) error {
	// Convert to our update type
	update := NewUpdateFromTelegram(tgUpdate)
	if update.ChatID == 0 {
		// Skip updates we can't handle
		return nil
	}

	// Load user from database
	user, err := mongo.FindUserByTelegramID(update.ChatID)
	if err != nil {
		log.Printf("Error loading user: %v", err)
		return err
	}

	// Create context
	ctx := &Context{
		Bot:          bot,
		Update:       update,
		User:         user,
		Localization: localization.NewLocalizer(getLanguageForUser(user)),
	}

	// Route based on update type
	switch update.Type {
	case UpdateTypeCommand:
		return r.handleCommand(ctx)
	case UpdateTypeCallback:
		return r.handleCallback(ctx)
	case UpdateTypeMessage:
		return r.handleMessage(ctx)
	default:
		log.Printf("Unknown update type: %v", update.Type)
		return nil
	}
}

// handleCommand routes command updates
func (r *Router) handleCommand(ctx *Context) error {
	command := strings.ToLower(ctx.Update.Data)

	// Remove the leading slash
	if strings.HasPrefix(command, "/") {
		command = command[1:]
	}

	handler, exists := r.commandHandlers[command]
	if !exists {
		log.Printf("No handler for command: %s", command)
		return nil
	}

	return handler.Handle(ctx)
}

// handleCallback routes callback query updates
func (r *Router) handleCallback(ctx *Context) error {
	// Answer the callback query first
	callback := tgbotapi.NewCallback(ctx.Update.Original.CallbackQuery.ID, "")
	if _, err := ctx.Bot.Request(callback); err != nil {
		log.Printf("Error answering callback: %v", err)
	}

	// Find handler by exact match or prefix
	callbackData := ctx.Update.Data

	// Try exact match first
	if handler, exists := r.callbackHandlers[callbackData]; exists {
		return handler.Handle(ctx)
	}

	// Try prefix match (for parameterized callbacks like "lang:en")
	for prefix, handler := range r.callbackHandlers {
		if strings.HasPrefix(callbackData, prefix+":") {
			return handler.Handle(ctx)
		}
	}

	log.Printf("No handler for callback: %s", callbackData)
	return nil
}

// handleMessage routes message updates based on user state or message handlers
func (r *Router) handleMessage(ctx *Context) error {
	// If user has a state, try state handler first
	if ctx.User != nil && ctx.User.State != "" {
		if handler, exists := r.stateHandlers[ctx.User.State]; exists {
			return handler.Handle(ctx)
		}
	}

	// Try message handlers
	for _, handler := range r.messageHandlers {
		if err := handler.Handle(ctx); err != nil {
			return err
		}
	}

	return nil
}

// getLanguageForUser returns the language code for a user
func getLanguageForUser(user *model.User) string {
	if user != nil && user.Language != "" {
		return user.Language
	}
	return "en" // Default language
}

// SendMessage sends a message with localization support
func (ctx *Context) SendMessage(messageKey localization.MessageKey, args ...interface{}) error {
	text := ctx.Localization.GetMessage(messageKey, args...)
	msg := tgbotapi.NewMessage(ctx.Update.ChatID, text)
	_, err := ctx.Bot.Send(msg)
	return err
}

// SendMessageWithKeyboard sends a message with inline keyboard
func (ctx *Context) SendMessageWithKeyboard(messageKey localization.MessageKey, keyboard tgbotapi.InlineKeyboardMarkup, args ...interface{}) error {
	text := ctx.Localization.GetMessage(messageKey, args...)
	msg := tgbotapi.NewMessage(ctx.Update.ChatID, text)
	msg.ReplyMarkup = keyboard
	_, err := ctx.Bot.Send(msg)
	return err
}

// EditMessage edits an existing message
func (ctx *Context) EditMessage(messageKey localization.MessageKey, args ...interface{}) error {
	text := ctx.Localization.GetMessage(messageKey, args...)
	edit := tgbotapi.NewEditMessageText(ctx.Update.ChatID, ctx.Update.MessageID, text)
	_, err := ctx.Bot.Send(edit)
	return err
}

// EditMessageWithKeyboard edits an existing message with keyboard
func (ctx *Context) EditMessageWithKeyboard(messageKey localization.MessageKey, keyboard tgbotapi.InlineKeyboardMarkup, args ...interface{}) error {
	text := ctx.Localization.GetMessage(messageKey, args...)
	edit := tgbotapi.NewEditMessageText(ctx.Update.ChatID, ctx.Update.MessageID, text)
	edit.ReplyMarkup = &keyboard
	_, err := ctx.Bot.Send(edit)
	return err
}
