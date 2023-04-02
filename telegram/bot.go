package telegram

import (
	"crypto/tls"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"librus/helper"
	"librus/librus"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

var client *mongo.Client

func checkLoginAndPassword(login string, password string) bool {
	if login == "" {
		return false
	}
	_, cancel, err := librus.Login(login, password)
	if err != nil {
		fmt.Println(err)
		return false
	}
	cancel()
	return true
}

func checkNewMessagesPeriodically(bot *tgbotapi.BotAPI) {
	for {
		time.Sleep(5 * time.Minute)
		users := getUsersFromDatabase()
		for _, user := range users {
			ctx, cancel, err := librus.Login(user.Login, user.Password)

			if err != nil {
				cancel()
				fmt.Println(err)
				continue
			}
			msgs, err := librus.GetMessages(ctx)
			cancel()

			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(msgs) == 0 {
				continue
			}
			msgs = addUserIdToMessages(msgs, user.TelegramID)

			msgs, err = addMessagesToDatabase(msgs)

			if err != nil {
				fmt.Println(err)
				continue
			}
			sort.Slice(msgs, func(i, j int) bool {
				return msgs[i].Date.Before(msgs[j].Date)
			})

			for _, message := range msgs {
				msg := tgbotapi.NewMessage(user.TelegramID, "")
				msg.ParseMode = tgbotapi.ModeHTML

				// добавляем заголовок
				titleLink := fmt.Sprintf("<a href='%s'>%s</a>", message.Link, message.Title)
				msg.Text += fmt.Sprintf("📰 <b>%s</b>\n\n", titleLink)

				// добавляем информацию об авторе
				msg.Text += fmt.Sprintf("👤 <i>От: %s</i>\n\n", message.Author)

				// добавляем основной текст
				msg.Text += "📝 " + replaceBrTags(message.Content) + "\n\n"

				// добавляем дату и время
				msg.Text += fmt.Sprintf("📅 %s", message.Date.Format("02.01.2006 15:04"))
				msg.Text += "\n_______________________________"

				// отправляем сообщение
				_, err = bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func addUserIdToMessages(msgs []librus.Message, id int64) []librus.Message {
	var result []librus.Message
	for _, msg := range msgs {
		msg.TelegramID = id
		result = append(result, msg)
	}
	return result
}

func replaceBrTags(s string) string {
	return strings.ReplaceAll(s, "<br>", "")
}

func Start() {
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	bot, err := tgbotapi.NewBotAPIWithClient(
		helper.GetEnv("TELEGRAM_TOKEN", "token"),
		"https://api.telegram.org/bot%s/%s",
		httpClient,
	)
	if err != nil {
		log.Fatal(err)
	}
	go checkNewMessagesPeriodically(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		fmt.Println("Received message: " + update.Message.Text)

		if update.Message.Text == "reset" {
			err = deleteUserByTelegramID(update.Message.Chat.ID)
			if err != nil {
				log.Println(err)
			}
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Все почищено, можно начать все с начала!",
			)
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			continue
		}
		user, _ := findUserByTelegramID(update.Message.Chat.ID)
		if user != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Не дрочи бот по чем зря, подписька уже оформлена, для сброса подписьки напиши reset",
			)
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			continue
		}
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

		authorization(login, password, update, bot)
	}

	// ждем, пока программа не будет остановлена
	select {}
}

func authorization(login string, password string, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if checkLoginAndPassword(login, password) {

		user := User{
			Login:      login,
			Password:   password,
			TelegramID: update.Message.Chat.ID,
		}
		addUserToDatabase(user)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно авторизованы!")
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	} else {
		// отправляем сообщение о неправильном логине или пароле
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный логин или пароль!")
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
