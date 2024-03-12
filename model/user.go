package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/translator"
	"log"
)

type User struct {
	Id          string  `bson:"_id"`
	Login       string  `bson:"login"`
	Password    string  `bson:"password"`
	TelegramIDs []int64 `bson:"telegram_ids"`
	Language    string  `bson:"language"`
}

func (user *User) SendTranslatedMessage(bot *tgbotapi.BotAPI, text string, forceLanguage ...string) {
	var err error
	if user.Language != "" {
		lang := user.Language
		if len(forceLanguage) > 0 {
			lang = forceLanguage[0]
		}
		text, err = translator.TranslateText(lang, text)
		if err != nil {
			log.Println(err)
			return
		}
	}
	for _, id := range user.TelegramIDs {
		msg := tgbotapi.NewMessage(id, text)
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
