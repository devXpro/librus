package telegram

import (
	"crypto/tls"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/helper"
	"librus/librus"
	"librus/model"
	"librus/mongo"
	"librus/telegram/channel"
	"librus/telegram/handler"
	"log"
	"net/http"
	"sort"
	"time"
)

func Start() {
	bot := createBot()
	go checkNewLibrusMessagesPeriodically(bot)
	u := createUpdateConfig()
	for update := range bot.GetUpdatesChan(u) {
		if update.Message == nil {
			continue
		}

		logReceivedMessage(update)

		handlers := []handler.Handler{
			&handler.DeleteAllMessages{},
			&handler.UpdateNow{},
			&handler.Reset{},
			&handler.Login{},
			&handler.Answer{},
			&handler.Language{},
			&handler.NoAction{},
		}

		for _, h := range handlers {
			if h.IsApplicable(update) {
				h.Handle(bot, update)
				break
			}
		}
	}

	select {}
}

func checkNewLibrusMessagesPeriodically(bot *tgbotapi.BotAPI) {
	for {
		select {
		case <-time.After(10000 * time.Second):
			fmt.Println("Start updating...")
		case <-channel.UpdateNow:
			fmt.Println("Start force update")
		}
		users := mongo.GetUsersFromDatabase()
		for _, user := range users {
			ctx, cancel, err := librus.Login(user.Login, user.Password)

			if err != nil {
				fmt.Println(err)
				continue
			}
			msgs, err := librus.GetMessages(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			news, err := librus.GetNews(ctx)
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
			msgs = addUserIdToMessages(msgs, user.TelegramID)

			msgs, err = mongo.AddMessagesToDatabase(msgs, user.TelegramID)

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
				err = message.Send(bot, user.TelegramID)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func addUserIdToMessages(msgs []model.Message, id int64) []model.Message {
	var result []model.Message
	for _, msg := range msgs {
		msg.TelegramID = id
		result = append(result, msg)
	}
	return result
}

func createBot() *tgbotapi.BotAPI {
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	bot, err := tgbotapi.NewBotAPIWithClient(
		helper.GetEnv("TELEGRAM_TOKEN", "token"),
		"https://api.telegram.org/bot%s/%s",
		httpClient,
	)
	if err != nil {
		log.Fatal(err)
	}

	return bot
}

func createUpdateConfig() tgbotapi.UpdateConfig {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3
	return u
}

func logReceivedMessage(update tgbotapi.Update) {
	fmt.Println("Received message: " + update.Message.Text)
	if update.Message.ReplyToMessage != nil {
		fmt.Println("ReplyToMessage message: " + update.Message.ReplyToMessage.Text)
	}
}
