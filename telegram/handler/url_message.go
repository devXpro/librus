package handler

import (
	"fmt"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type URLMessage struct {
}

// IsApplicable checks if the handler should process this update
func (u *URLMessage) IsApplicable(update tgbotapi.Update) bool {
	return update.Message != nil &&
		strings.HasPrefix(strings.ToLower(update.Message.Text), "url:") &&
		len(update.Message.Text) > 5
}

// Handle processes the URL message command
func (u *URLMessage) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Extract the URL from the message
	urlText := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "url:"))
	if !strings.HasPrefix(urlText, "https://synergia.librus.pl/") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid URL format. The URL must start with https://synergia.librus.pl/")
		bot.Send(msg)
		return
	}

	// Find user by Telegram ID
	user, err := mongo.FindUserByTelegramID(update.Message.Chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error finding user data: "+err.Error())
		bot.Send(msg)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You need to be authorized first. Please send your login and password in format login:password")
		bot.Send(msg)
		return
	}

	// Let the user know we're processing their request
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Processing your request...")
	bot.Send(msg)

	// Create gRPC client
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Service connection error: "+err.Error())
		bot.Send(msg)
		return
	}
	defer client.Close()

	// Process the specific message using gRPC GetSingleMessage method
	message, err := client.GetSingleMessage(user.Login, user.Password, urlText)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to process message: "+err.Error())
		bot.Send(msg)
		return
	}

	// Make sure the message has a user ID assigned
	message.UserID = user.Id

	// Apply user's language preference
	if user.Language != "" {
		message.Translate(user.Language)
	}

	// Send the message to the user
	err = message.Send(bot, update.Message.Chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to send message: "+err.Error())
		bot.Send(msg)
		return
	}

	// Clean up attachments after sending
	if err := message.CleanupAttachments(); err != nil {
		fmt.Printf("Error cleaning up attachments: %v\n", err)
	}

	// Send translated success message using user.SendTranslatedMessage
	user.SendTranslatedMessage(bot, "Message processed successfully.")
}
