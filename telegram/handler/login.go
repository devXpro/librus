package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/model"
	"librus/mongo"
	"librus/parser"
	"log"
	"strings"
)

type Login struct {
}

func (l *Login) IsApplicable(update tgbotapi.Update) bool {
	user, _ := mongo.FindUserByTelegramID(update.Message.Chat.ID)
	return user == nil
}

func (l *Login) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	loginAndPass := strings.Split(update.Message.Text, ":")
	var login string
	var password string
	if len(loginAndPass) != 2 {
		login = ""
		password = ""
	} else {
		login = loginAndPass[0]
		password = loginAndPass[1]
	}
	l.authorization(login, password, update, bot)
}

func (l *Login) authorization(login string, password string, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if l.checkLoginAndPassword(login, password) {

		user := model.User{
			Login:      login,
			Password:   password,
			TelegramID: update.Message.Chat.ID,
		}
		mongo.AddUserToDatabase(user)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are authorized successfully")
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid login or password")
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
func (l *Login) checkLoginAndPassword(login string, password string) bool {
	if login == "" {
		return false
	}
	_, cancel, err := parser.Login(login, password)
	if err != nil {
		fmt.Println(err)
		return false
	}
	cancel()
	return true
}
