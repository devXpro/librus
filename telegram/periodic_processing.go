package telegram

import (
	"fmt"
	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/telegram/channel"
	"sort"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func checkNewLibrusMessagesPeriodically(bot *tgbotapi.BotAPI) {
	// Create gRPC client once and reuse it
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		fmt.Printf("Failed to create gRPC client: %v\n", err)
		return
	}
	defer client.Close()

	for {
		select {
		case <-time.After(30 * time.Minute):
			fmt.Println("Start updating...")
		case <-channel.UpdateNow:
			fmt.Println("Start force update")
		}
		accounts := mongo.GetLibrusAccountsFromDatabase()
		for _, account := range accounts {
			// Use gRPC GetAllUpdates to get both messages and news in one call
			msgs, news, err := client.GetAllUpdates(account.Login, account.Password)
			if err != nil {
				fmt.Printf("Failed to get updates for account %s: %v\n", account.Login, err)
				continue
			}

			// Combine messages and news
			allMsgs := append(msgs, news...)
			if len(allMsgs) == 0 {
				continue
			}
			allMsgs = addLibrusLoginToMessages(allMsgs, account.Login)

			allMsgs, err = mongo.AddMessagesToDatabase(allMsgs, account.Login)
			if err != nil {
				fmt.Println(err)
				continue
			}

			sort.Slice(allMsgs, func(i, j int) bool {
				return allMsgs[i].Date.Before(allMsgs[j].Date)
			})

			// Get all telegram users for this Librus account
			telegramUsers, err := mongo.GetTelegramUsersByLibrusLogin(account.Login)
			if err != nil {
				fmt.Printf("Failed to get telegram users for account %s: %v\n", account.Login, err)
				continue
			}

			// Send messages to each telegram user
			for _, message := range allMsgs {
				for _, telegramUser := range telegramUsers {
					// Check if message was already sent to this user
					if mongo.IsMessageSentToUser(telegramUser.Id, message.Id) {
						continue
					}

					// Translate message if user has language preference
					translatedMessage := message
					if telegramUser.Language != "" {
						translatedMessage.Translate(telegramUser.Language)
					}

					// Send message
					err = translatedMessage.Send(bot, telegramUser.TelegramID)
					if err != nil {
						fmt.Printf("Error sending message to user %s: %v\n", telegramUser.Id, err)
						continue
					}

					// Mark message as sent
					err = mongo.MarkMessageAsSent(telegramUser.Id, message.Id)
					if err != nil {
						fmt.Printf("Error marking message as sent: %v\n", err)
					}
				}

				// Clean up attachments directory after sending to all users
				if err := message.CleanupAttachments(); err != nil {
					fmt.Printf("Error cleaning up attachments: %v\n", err)
				}
			}
		}
	}
}

func addLibrusLoginToMessages(msgs []model.Message, login string) []model.Message {
	var result []model.Message
	for _, msg := range msgs {
		msg.LibrusLogin = login
		result = append(result, msg)
	}
	return result
}
