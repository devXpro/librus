package telegram

import (
	"crypto/tls"
	"fmt"
	"librus/helper"
	"librus/telegram/handler"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
			&handler.URLMessage{},
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
