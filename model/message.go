package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/translator"
	"strings"
	"time"
)

type MessageType string

const (
	MsgTypeMessage      MessageType = "message"
	MsgTypeNotification MessageType = "notification"
	MsgTypeNews         MessageType = "news"
)

type Message struct {
	Id      string      `bson:"_id"`
	Type    MessageType `bson:"type"`
	Link    string      `bson:"link"`
	Author  string      `bson:"author"`
	Title   string      `bson:"title"`
	Content string      `bson:"content"`
	Date    time.Time   `bson:"date"`
	UserID  string      `bson:"user_id"`
}

func (message *Message) GenerateId() {
	stringToHash := ""
	if message.Type == MsgTypeNotification {
		stringToHash = message.Title + message.Content + message.Date.Format("2006-01-02")
	} else if message.Type == MsgTypeMessage {
		stringToHash = message.Link
	} else {
		panic("Wrong message type")
	}
	h := md5.New()
	h.Write([]byte(stringToHash))
	message.Id = hex.EncodeToString(h.Sum(nil))
}
func (message *Message) Translate(lang string) {
	translation, err := translator.TranslateText(lang, message.Title)
	if err == nil {
		message.Title = translation
	}

	translation, err = translator.TranslateText(lang, message.Content)
	if err == nil {
		message.Content = translation
	}

	translation, err = translator.TranslateText(lang, message.Author)
	if err == nil {
		message.Author = translation
	}
}

func (message *Message) Send(bot *tgbotapi.BotAPI, telegramId int64) error {
	msg := tgbotapi.NewMessage(telegramId, "")
	msg.ParseMode = tgbotapi.ModeHTML

	// Добавляем информацию о типе сообщения
	var typeIcon string
	switch message.Type {
	case MsgTypeMessage:
		typeIcon = "📩"
	case MsgTypeNotification:
		typeIcon = "🔔"
	default:
		typeIcon = ""
	}
	msg.Text += fmt.Sprintf("%s <b><a href=\"%s\">%s</a></b>\n\n", typeIcon, message.Link, message.Title)

	// Добавляем информацию об авторе
	msg.Text += fmt.Sprintf("👤 <i>От: %s</i>\n\n", message.Author)

	// Добавляем основной текст
	msg.Text += "📝 " + replaceBrTags(message.Content) + "\n\n"

	// Добавляем дату и время
	msg.Text += fmt.Sprintf("📅 %s", message.Date.Format("02.01.2006 15:04"))
	msg.Text += "\n_______________________________"

	// Отправляем сообщение
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func replaceBrTags(s string) string {
	return strings.ReplaceAll(s, "<br>", "")
}
