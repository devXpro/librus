package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/mongo"
)

type NoAction struct{}

func (_ *NoAction) IsApplicable(update tgbotapi.Update) bool {
	return true
}

func (_ *NoAction) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	user, _ := mongo.FindUserByTelegramID(update.Message.Chat.ID)
	user.SendTranslatedMessage(bot, "Please don't worry, the subscription is already active. "+
		"If you'd like to reset your subscription, simply type <nt>\"reset\"</nt>")
}
