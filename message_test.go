package main

import (
	"librus/helper"
	"librus/parser"
	"log"
	"strconv"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func TestGetAndSendSpecificMessage(t *testing.T) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Get telegram credentials from environment variables
	telegramTokenEnv := helper.GetEnv("TELEGRAM_TOKEN", "")
	telegramUserIDStr := helper.GetEnv("TELEGRAM_TEST_USER_ID", "0")
	librusLogin := helper.GetEnv("TEST_LIBRUS_USER", "")
	librusPassword := helper.GetEnv("TEST_LIBRUS_PASSWORD", "")

	// Skip test if any required env variable is missing
	if telegramTokenEnv == "" || telegramUserIDStr == "0" || librusLogin == "" || librusPassword == "" {
		t.Skip("Skipping test: One or more required environment variables are missing")
	}

	// Convert telegram user ID to int64
	telegramUserID, err := strconv.ParseInt(telegramUserIDStr, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse TELEGRAM_TEST_USER_ID: %v", err)
	}

	// URL of a specific message to retrieve and send
	messageURL := "https://synergia.librus.pl/wiadomosci/1/5/10084091/f0"

	// Log in to Librus using local browser instead of remote
	t.Log("Logging in to Librus...")
	ctx, cancel, err := TestLogin(librusLogin, librusPassword)
	if err != nil {
		t.Fatalf("Failed to login to Librus: %v", err)
	}
	defer cancel()

	// Get the message using the existing GetSingleMessage function
	t.Log("Fetching message from Librus...")
	message, err := parser.GetSingleMessage(ctx, messageURL)
	if err != nil {
		t.Fatalf("Failed to get message: %v", err)
	}

	// Print message details for debugging
	t.Logf("Message Title: %s", message.Title)
	t.Logf("Message Author: %s", message.Author)
	t.Logf("Message Date: %s", message.Date.Format(time.RFC3339))
	t.Logf("Has Attachments: %v", message.AttachmentsDir != "")

	// Translate the message to Ukrainian before sending
	t.Log("Translating message to Ukrainian...")
	message.Translate("uk")

	// Print translated message details
	t.Logf("Translated Message Title: %s", message.Title)
	t.Logf("Translated Message Author: %s", message.Author)

	// Create Telegram bot
	bot, err := tgbotapi.NewBotAPI(telegramTokenEnv)
	if err != nil {
		t.Fatalf("Failed to create Telegram bot: %v", err)
	}

	// Send the actual message
	err = message.Send(bot, telegramUserID)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Send a completion notification
	completionMsg := tgbotapi.NewMessage(telegramUserID, "✅ Test completed successfully!")
	_, err = bot.Send(completionMsg)
	if err != nil {
		t.Logf("Warning: Failed to send completion notification: %v", err)
	}
}
