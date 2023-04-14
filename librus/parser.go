package librus

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"librus/helper"
	"log"
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
	Id         string      `bson:"_id"`
	Type       MessageType `bson:"type"`
	Link       string      `bson:"link"`
	Author     string      `bson:"author"`
	Title      string      `bson:"title"`
	Content    string      `bson:"content"`
	Date       time.Time   `bson:"date"`
	TelegramID int64       `bson:"telegram_id"`
}

func (message *Message) GenerateId() {
	stringToHash := ""
	if message.Type == MsgTypeNotification {
		stringToHash = message.Title + message.Date.Format("2006-01-02")
	} else if message.Type == MsgTypeMessage {
		stringToHash = message.Link
	} else {
		panic("Wrong message type")
	}
	h := md5.New()
	h.Write([]byte(stringToHash))
	message.Id = hex.EncodeToString(h.Sum(nil))
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
	msg.Text += fmt.Sprintf("%s <b>%s</b>\n\n", typeIcon, message.Title)

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

func Login(login string, password string) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

	headless := true
	var actx context.Context
	if !headless {
		options := []chromedp.ExecAllocatorOption{
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134"),
			chromedp.WindowSize(2000, 2000),
			chromedp.Flag("headless", false),
		}

		actx, _ = chromedp.NewExecAllocator(ctx, options...)
	} else {
		actx, _ = chromedp.NewRemoteAllocator(context.Background(), "ws://surfshark:3000")
		//actx, _ = chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:3900")
	}
	ctx, cancel = chromedp.NewContext(actx)

	// Операции для авторизации
	ops := []chromedp.Action{
		chromedp.EmulateViewport(2000, 2000),
		chromedp.Navigate(`https://portal.librus.pl/rodzina`),
		logAction("Навигация на страницу https://portal.librus.pl/rodzina"),
		chromedp.WaitVisible(`.modal-button__primary`, chromedp.ByQuery),
		logAction("Ожидание появления элемента .modal-button__primary"),
		chromedp.Click(`.modal-button__primary`, chromedp.ByQuery),
		logAction("Нажатие на элемент .modal-button__primary"),
		chromedp.Sleep(3 * time.Second),
		logAction("Задержка на 3 секунды"),
		chromedp.WaitVisible(`a.btn-synergia-top[href="#"]`, chromedp.ByQuery),
		logAction("Ожидание появления элемента a.btn-synergia-top[href=\"#\"]"),
		chromedp.Click(`a[href="#"]`, chromedp.ByQuery),
		logAction("Нажатие на элемент a[href=\"#\"]"),
		chromedp.WaitVisible(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		logAction("Ожидание появления элемента a[href=\"/rodzina/synergia/loguj\"]"),
		chromedp.Click(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		logAction("Нажатие на элемент a[href=\"/rodzina/synergia/loguj\"]"),
		chromedp.Sleep(3 * time.Second),
		chromedp.Click(`#loginLabel`),
		logAction("Установка фокуса на логин"),
		chromedp.SendKeys(`#Login`, login),
		chromedp.Sleep(1 * time.Second),
		logAction("Ввод логина"),
		chromedp.Click(`#passwordLabel`),
		logAction("Установка фокуса на пароль"),
		chromedp.SendKeys(`#Pass`, password),
		chromedp.Sleep(1 * time.Second),
		logAction("Ввод пароля"),
		chromedp.Click(`#LoginBtn`),
		logAction("Нажатие на кнопку входа"),
		chromedp.Sleep(3 * time.Second),
	}

	if err := chromedp.Run(ctx, ops...); err != nil {
		cancel()
		return nil, cancel, err
	}
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		cancel()
		return nil, cancel, err
	}
	if currentURL == "https://portal.librus.pl/rodzina/synergia/loguj" {
		cancel()
		return nil, nil, errors.New("invalid Login")
	}
	return ctx, cancel, nil
}
func logAction(name string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if helper.IsDebug() {
			log.Printf("Операция: %s", name)
		}
		return nil
	})
}

func GetMessages(ctx context.Context) ([]Message, error) {
	err := chromedp.Run(ctx, chromedp.Navigate(`https://synergia.librus.pl/wiadomosci`),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		logAction("Навигация на страницу wiadomosci"),
		chromedp.Sleep(3*time.Second))
	if err != nil {
		return nil, err
	}

	linksCtx, linksCancel := context.WithTimeout(ctx, time.Second)
	defer linksCancel()
	// Получаем HTML-код страницы
	var html string
	if err = chromedp.Run(linksCtx, chromedp.InnerHTML(`html`, &html)); err != nil {
		return nil, err
	}
	printLog("Таблица получена! Загружаем HTML-код страницы в объект goquery.Document")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	printLog("Ищем все ссылки в таблице с классом \"decorated\" ...")
	var links []string
	//table.decorated td a
	doc.Find("table.decorated td[style=\"font-weight: bold;\"] a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok {
			links = append(links, href)
		}
	})
	printLog("линки готовы!")
	var messages []Message
	printLog("Spin this shit!")
	links = removeDuplicates(links)
	for _, link := range links {
		link = "https://synergia.librus.pl" + link
		if strings.Contains(link, "javascript") {
			continue
		}
		err = chromedp.Run(ctx, chromedp.Navigate(link))
		printLog(link)
		if err != nil {
			return nil, err
		}
		var dateString string
		err = chromedp.Run(ctx, chromedp.Text(
			"//*[@id=\"formWiadomosci\"]/div/div/table/tbody/tr/td[2]/table[2]/tbody/tr[3]/td[2]",
			&dateString, chromedp.NodeVisible,
			chromedp.BySearch),
		)
		var author string
		err = chromedp.Run(ctx, chromedp.Text(
			"//*[@id=\"formWiadomosci\"]/div/div/table/tbody/tr/td[2]/table[2]/tbody/tr[1]/td[2]",
			&author, chromedp.NodeVisible,
			chromedp.BySearch),
		)
		if err != nil {
			return nil, err
		}
		var title string
		err = chromedp.Run(ctx, chromedp.Text(
			"//*[@id=\"formWiadomosci\"]/div/div/table/tbody/tr/td[2]/table[2]/tbody/tr[2]/td[2]",
			&title, chromedp.NodeVisible,
			chromedp.BySearch),
		)
		var date time.Time
		date, err = time.Parse("2006-01-02 15:04:05", dateString)
		if err != nil {
			return nil, err
		}
		var content string
		err = chromedp.Run(ctx, chromedp.InnerHTML(`.container-message-content`, &content, chromedp.ByQuery))
		if err != nil {
			return nil, err
		}
		message := Message{Link: link, Content: content, Date: date, Title: title, Author: author, Type: MsgTypeMessage}
		message.GenerateId()
		messages = append(messages, message)
	}
	printLog("Линки обработаны...")
	return messages, nil
}

func GetNews(ctx context.Context) ([]Message, error) {
	const url = "https://synergia.librus.pl/ogloszenia"
	var messages []Message

	var html string
	if err := chromedp.Run(ctx, chromedp.Navigate(url), chromedp.OuterHTML("html", &html)); err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	tableNodes := doc.Find("table")

	var parseErr error
	tableNodes.Each(func(i int, tableNode *goquery.Selection) {
		message := Message{
			Type:    "notification",
			Link:    url,
			Author:  tableNode.Find("th:contains('Dodał')").Next().Text(),
			Title:   tableNode.Find("thead td[colspan='2']").Text(),
			Content: tableNode.Find("th:contains('Treść')").Next().Text(),
			Date: func() time.Time {
				var date time.Time
				date, err = time.Parse("2006-01-02",
					tableNode.Find("th:contains('Data publikacji')").Next().Text())
				if err != nil {
					parseErr = err
					return time.Time{}
				}
				return date
			}(),
			TelegramID: 0,
		}
		message.GenerateId()
		messages = append(messages, message)
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return messages, nil
}

func printLog(text string) {
	if helper.IsDebug() {
		fmt.Println(text)
	}
}

func removeDuplicates(strings []string) []string {
	encountered := map[string]struct{}{}
	var result []string

	for _, str := range strings {
		if _, ok := encountered[str]; !ok {
			encountered[str], result = struct{}{}, append(result, str)
		}
	}

	return result
}
