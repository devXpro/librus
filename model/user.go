package model

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/translator"
	"log"
)

// UserState represents the current state of user interaction
type UserState string

const (
	StateLanguageSelection UserState = "language_selection"
	StateAwaitingLogin     UserState = "awaiting_login"
	StateAwaitingPassword  UserState = "awaiting_password"
	StateAuthenticated     UserState = "authenticated"
	StateAwaitingURL       UserState = "awaiting_url"
)

type User struct {
	Id          string    `bson:"_id"`
	Login       string    `bson:"login"`
	Password    string    `bson:"password"`
	TelegramIDs []int64   `bson:"telegram_ids"`
	Language    string    `bson:"language"`
	State       UserState `bson:"state"`
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
