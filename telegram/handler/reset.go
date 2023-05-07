package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/mongo"
	"log"
)

type Reset struct{}

func (_ *Reset) IsApplicable(update tgbotapi.Update) bool {
	return update.Message.Text == "reset"
}

func (_ *Reset) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	user, _ := mongo.FindUserByTelegramID(update.Message.Chat.ID)

	err := mongo.DeleteUserByTelegramID(update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	user.SendTranslatedMessage(bot, "Everything is cleaned up, we can start from scratch!")
}
