package parser

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"time"
)

func Login(login string, password string) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

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
