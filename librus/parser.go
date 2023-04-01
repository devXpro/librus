package librus

import (
	"context"
	"errors"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"librus/helper"
	"log"
	"strings"
	"time"
)

type Message struct {
	Link       string    `bson:"link"`
	Author     string    `bson:"author"`
	Title      string    `bson:"title"`
	Content    string    `bson:"content"`
	Date       time.Time `bson:"date"`
	TelegramID int64     `bson:"telegram_id"`
}

func Login(login string, password string) (context.Context, context.CancelFunc, error) {
	headless := true

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	options := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134"),
		chromedp.WindowSize(1920, 1080),
	}

	if headless {
		options = append(options, chromedp.Headless)
	}

	actx, _ := chromedp.NewExecAllocator(ctx, options...)
	ctx, cancel = chromedp.NewContext(actx)

	logAction := func(name string) chromedp.Action {
		return chromedp.ActionFunc(func(ctx context.Context) error {
			if helper.IsDebug() {
				log.Printf("Операция: %s", name)
			}
			return nil
		})
	}

	// Операции для авторизации
	ops := []chromedp.Action{
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
		chromedp.Sleep(2 * time.Second),
	}

	if err := chromedp.Run(ctx, ops...); err != nil {
		return nil, cancel, err
	}
	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		return nil, cancel, err
	}
	if currentURL == "https://portal.librus.pl/rodzina/synergia/loguj" {
		cancel()
		return nil, nil, errors.New("invalid Login")
	}
	return ctx, cancel, nil
}

func GetMessages(ctx context.Context) ([]Message, error) {
	err := chromedp.Run(ctx, chromedp.Navigate(`https://synergia.librus.pl/wiadomosci`))
	chromedp.Sleep(3 * time.Second)
	if err != nil {
		return nil, err
	}
	var links []*cdp.Node
	tableCtx, cancelTable := context.WithTimeout(ctx, time.Second)
	defer cancelTable()
	//err = chromedp.Run(tableCtx, chromedp.Nodes(`table.decorated td[style="font-weight: bold;"] a`, &links, chromedp.ByQueryAll))
	err = chromedp.Run(tableCtx, chromedp.Nodes(`table.decorated td a`, &links, chromedp.ByQueryAll))

	if err != nil {
		return nil, err
	}
	var messages []Message
	for _, link := range links {
		href := "https://synergia.librus.pl" + link.AttributeValue("href")
		if strings.Contains(href, "javascript") {
			continue
		}
		err = chromedp.Run(ctx, chromedp.Navigate(href))
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
		messages = append(messages, Message{Link: href, Content: content, Date: date, Title: title, Author: author})
	}
	return messages, nil
}