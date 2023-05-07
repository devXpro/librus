package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/mongo"
	"librus/translator"
	"log"
	"strings"
)

type Language struct{}

func (_ *Language) IsApplicable(update tgbotapi.Update) bool {
	return strings.Contains(update.Message.Text, "lang:")
}

func (_ *Language) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	langParts := strings.Split(update.Message.Text, ":")
	lang := langParts[1]
	text := "The language has been successfully set, now all messages will be translated!"
	text, err := translator.TranslateText(lang, text)
	if err != nil {
		text = "Wrong language '" + lang + "'"
	} else {
		err = mongo.UpdateUserLanguageByTelegramID(update.Message.Chat.ID, lang)
		if err != nil {
			text = "Can't update user language"
		}
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
