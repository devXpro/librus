package main

import (
	"context"
	"errors"
	"time"

	"github.com/chromedp/chromedp"
)

// TestLogin is a special version of Login that doesn't use remote allocator
// It's used only for testing purposes
func TestLogin(login string, password string) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	// Always use local Chrome/Chromium with headless mode
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134"),
		chromedp.WindowSize(2000, 2000),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	actx, actxCancel := chromedp.NewExecAllocator(ctx, options...)
	// Create a new context
	ctx, cancel = chromedp.NewContext(actx)

	// Ensure cleanup
	oldCancel := cancel
	cancel = func() {
		oldCancel()
		actxCancel()
	}

	// Operations for authorization
	ops := []chromedp.Action{
		chromedp.EmulateViewport(2000, 2000),
		chromedp.Navigate(`https://portal.librus.pl/rodzina`),
		chromedp.WaitVisible(`.modal-button__primary`, chromedp.ByQuery),
		chromedp.Click(`.modal-button__primary`, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second),
		chromedp.WaitVisible(`a.btn-synergia-top[href="#"]`, chromedp.ByQuery),
		chromedp.Click(`a[href="#"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		chromedp.Click(`a[href="/rodzina/synergia/loguj"]`, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second),
		chromedp.Click(`#loginLabel`),
		chromedp.SendKeys(`#Login`, login),
		chromedp.Sleep(1 * time.Second),
		chromedp.Click(`#passwordLabel`),
		chromedp.SendKeys(`#Pass`, password),
		chromedp.Sleep(1 * time.Second),
		chromedp.Click(`#LoginBtn`),
		chromedp.Sleep(3 * time.Second),
	}

	if err := chromedp.Run(ctx, ops...); err != nil {
		cancel()
		return nil, nil, err
	}

	var currentURL string
	if err := chromedp.Run(ctx, chromedp.Location(&currentURL)); err != nil {
		cancel()
		return nil, nil, err
	}

	if currentURL == "https://portal.librus.pl/rodzina/synergia/loguj" {
		cancel()
		return nil, nil, errors.New("invalid Login")
	}

	return ctx, cancel, nil
}
