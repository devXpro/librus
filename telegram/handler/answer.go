package handler

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/librus"
	"librus/mongo"
	"librus/translator"
	"log"
	"strings"
)

type Answer struct{}

func (a *Answer) IsApplicable(update tgbotapi.Update) bool {
	return update.Message.ReplyToMessage != nil
}

func (a *Answer) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	replyToMessage := update.Message.ReplyToMessage
	user, _ := mongo.FindUserByTelegramID(update.Message.Chat.ID)
	link, err := a.extractLink(replyToMessage)
	if err != nil {
		user.SendTranslatedMessage(bot, "Wrong link")
		log.Println(err)
		return
	}
	if !strings.Contains(link, "wiadomosci") {
		user.SendTranslatedMessage(bot, "It is not possible to reply to this type of messages.")
		return
	}
	ctx, cancel, err := librus.Login(user.Login, user.Password)
	defer cancel()
	if err != nil {
		user.SendTranslatedMessage(bot, "Login issues...")
		return
	}
	text := update.Message.Text
	if user.Language != "" {
		text, err = translator.TranslateText("pl", text)
		if err != nil {
			log.Println(err)
			user.SendTranslatedMessage(bot, "Translation error")
			return
		}
	}
	err = librus.AnswerMessage(link, text, ctx)
	if err != nil {
		user.SendTranslatedMessage(bot, "Can't answer, something went wrong")
	}
	user.SendTranslatedMessage(bot, "Message sent successfully")
}

func (a *Answer) extractLink(message *tgbotapi.Message) (string, error) {
	for _, entity := range message.Entities {
		if entity.Type == "text_link" {
			return entity.URL, nil
		}
	}

	return "", errors.New("can't extract link")
}
