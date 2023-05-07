package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/translator"
	"log"
)

type User struct {
	Login      string `bson:"login"`
	Password   string `bson:"password"`
	TelegramID int64  `bson:"telegram_id"`
	Language   string `bson:"language"`
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

	msg := tgbotapi.NewMessage(user.TelegramID, text)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
