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
		users := mongo.GetUsersFromDatabase()
		for _, user := range users {
			// Skip users who are not authenticated or don't have credentials
			if user.State != model.StateAuthenticated || user.Login == "" || user.Password == "" {
				fmt.Printf("Skipping user %s: not authenticated or missing credentials\n", user.Id)
				continue
			}

			// Use gRPC GetAllUpdates to get both messages and news in one call
			msgs, news, err := client.GetAllUpdates(user.Login, user.Password)
			if err != nil {
				fmt.Printf("Failed to get updates for user %s: %v\n", user.Login, err)
				continue
			}

			// Combine messages and news
			allMsgs := append(msgs, news...)
			if len(allMsgs) == 0 {
				continue
			}
			allMsgs = addUserIdToMessages(allMsgs, user.Id)

			allMsgs, err = mongo.AddMessagesToDatabase(allMsgs, user.Id)

			if err != nil {
				fmt.Println(err)
				continue
			}
			sort.Slice(allMsgs, func(i, j int) bool {
				return allMsgs[i].Date.Before(allMsgs[j].Date)
			})

			for _, message := range allMsgs {
				if user.Language != "" {
					message.Translate(user.Language)
				}
				for _, id := range user.TelegramIDs {
					err = message.Send(bot, id)
					if err != nil {
						fmt.Println(err)
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

func addUserIdToMessages(msgs []model.Message, id string) []model.Message {
	var result []model.Message
	for _, msg := range msgs {
		msg.UserID = id
		result = append(result, msg)
	}
	return result
}
