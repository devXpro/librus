package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/helper"
	"librus/mongo"
	"log"
)

type DeleteAllMessages struct{}

func (_ *DeleteAllMessages) IsApplicable(update tgbotapi.Update) bool {
	return update.Message.Text == "delete_all_messages_"+helper.GetEnv("TELEGRAM_TOKEN", "pass")
}

func (_ *DeleteAllMessages) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	_ = mongo.DeleteAllMessages()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "All messages was deleted")
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
