package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/model"
	"librus/mongo"
	"librus/parser"
	"librus/telegram/channel"
	"sort"
	"time"
)

func checkNewLibrusMessagesPeriodically(bot *tgbotapi.BotAPI) {
	for {
		select {
		case <-time.After(30 * time.Minute):
			fmt.Println("Start updating...")
		case <-channel.UpdateNow:
			fmt.Println("Start force update")
		}
		users := mongo.GetUsersFromDatabase()
		for _, user := range users {
			ctx, cancel, err := parser.Login(user.Login, user.Password)

			if err != nil {
				fmt.Println(err)
				continue
			}
			msgs, err := parser.GetMessages(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			news, err := parser.GetNews(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			msgs = append(msgs, news...)
			cancel()

			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(msgs) == 0 {
				continue
			}
			msgs = addUserIdToMessages(msgs, user.Id)

			msgs, err = mongo.AddMessagesToDatabase(msgs, user.Id)

			if err != nil {
				fmt.Println(err)
				continue
			}
			sort.Slice(msgs, func(i, j int) bool {
				return msgs[i].Date.Before(msgs[j].Date)
			})

			for _, message := range msgs {
				if user.Language != "" {
					message.Translate(user.Language)
				}
				for _, id := range user.TelegramIDs {
					err = message.Send(bot, id)
					if err != nil {
						fmt.Println(err)
					}
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
