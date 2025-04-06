package parser

import (
	"context"
	"errors"
	"time"

	"github.com/chromedp/chromedp"
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

	// Operations for authorization
	ops := []chromedp.Action{
		chromedp.EmulateViewport(2000, 2000),
		chromedp.Navigate(`https://portal.librus.pl/rodzina`),
		logAction("Navigating to https://portal.librus.pl/rodzina"),
		chromedp.WaitVisible(`.modal-button__primary`, chromedp.ByQuery),
		logAction("Waiting for .modal-button__primary element"),
		chromedp.Click(`.modal-button__primary`, chromedp.ByQuery),
		logAction("Clicking on .modal-button__primary element"),
		chromedp.Sleep(3 * time.Second),
		logAction("Delay for 3 seconds"),
		chromedp.WaitVisible(`a.btn-synergia-top[href="#"]`, chromedp.ByQuery),
		logAction("Waiting for a.btn-synergia-top[href=\"#\"] element"),
		chromedp.Click(`a[href="#"]`, chromedp.ByQuery),
		logAction("Clicking on a[href=\"#\"] element"),
		chromedp.WaitVisible(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		logAction("Waiting for a[href=\"/rodzina/synergia/loguj\"] element"),
		chromedp.Click(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		logAction("Clicking on a[href=\"/rodzina/synergia/loguj\"] element"),
		chromedp.Sleep(3 * time.Second),
		chromedp.Click(`#loginLabel`),
		logAction("Setting focus on login field"),
		chromedp.SendKeys(`#Login`, login),
		chromedp.Sleep(1 * time.Second),
		logAction("Entering login"),
		chromedp.Click(`#passwordLabel`),
		logAction("Setting focus on password field"),
		chromedp.SendKeys(`#Pass`, password),
		chromedp.Sleep(1 * time.Second),
		logAction("Entering password"),
		chromedp.Click(`#LoginBtn`),
		logAction("Clicking login button"),
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
