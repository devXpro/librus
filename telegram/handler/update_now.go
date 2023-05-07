package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/helper"
	"librus/telegram/channel"
)

type UpdateNow struct{}

func (_ *UpdateNow) IsApplicable(update tgbotapi.Update) bool {
	return update.Message.Text == "update_now_"+helper.GetEnv("TELEGRAM_TOKEN", "pass")
}

func (_ *UpdateNow) Handle(_ *tgbotapi.BotAPI, _ tgbotapi.Update) {
	channel.UpdateNow <- struct{}{}
}
