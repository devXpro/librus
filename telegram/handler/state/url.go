package state

import (
	"fmt"
	"log"
	"strings"

	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// URLHandler handles URL input state
type URLHandler struct{}

// Handle processes URL input messages
func (h *URLHandler) Handle(ctx *router.Context) error {
	// Check if telegram user has valid credentials (state should be StateAwaitingURL when this handler is called)
	if ctx.User == nil || ctx.User.LibrusLogin == "" {
		return ctx.SendMessage(localization.MsgPleaseLogin)
	}

	// Get Librus account
	librusAccount, err := mongo.FindLibrusAccount(ctx.User.LibrusLogin)
	if err != nil || librusAccount == nil {
		return ctx.SendMessage(localization.MsgPleaseLogin)
	}

	urlText := strings.TrimSpace(ctx.Update.Data)

	// Validate URL format
	if !strings.HasPrefix(urlText, "https://synergia.librus.pl/") {
		cancelKeyboard := keyboard.CancelKeyboard(ctx.Localization)
		return ctx.SendMessageWithKeyboard(localization.MsgInvalidURL, cancelKeyboard)
	}

	// Show processing message
	err = ctx.SendMessage(localization.MsgProcessing)
	if err != nil {
		log.Printf("Error sending processing message: %v", err)
	}

	// Create gRPC client
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return ctx.SendMessage(localization.MsgServiceError)
	}
	defer client.Close()

	// Process the specific message
	message, err := client.GetSingleMessage(librusAccount.Login, librusAccount.Password, urlText)
	if err != nil {
		log.Printf("Failed to process message: %v", err)
		return ctx.SendMessage(localization.MsgServiceError)
	}

	// Apply user's language preference
	if ctx.User.Language != "" {
		message.Translate(ctx.User.Language)
	}

	// Send the message
	err = message.Send(ctx.Bot, ctx.Update.ChatID)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	// Clean up attachments
	if err := message.CleanupAttachments(); err != nil {
		fmt.Printf("Error cleaning up attachments: %v\n", err)
	}

	// Reset telegram user state and show main menu
	err = mongo.UpdateTelegramUserState(ctx.Update.ChatID, model.StateAuthenticated)
	if err != nil {
		log.Printf("Error updating telegram user state: %v", err)
	}

	// Send success message with main menu
	mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
	return ctx.SendMessageWithKeyboard(localization.MsgMessageSent, mainMenu)
}
