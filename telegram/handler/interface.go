package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	IsApplicable(update tgbotapi.Update) bool
	Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}
