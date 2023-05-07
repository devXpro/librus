package parser

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

func AnswerMessage(link string, text string, ctx context.Context) error {
	err := chromedp.Run(ctx, chromedp.Navigate(link),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		logAction(fmt.Sprintf("Navigation in %s", link)),
		chromedp.Sleep(2*time.Second),
		logAction("Click on Odpowiedz"),
		chromedp.Click(`input[type="button"][value="Odpowiedz"]`, chromedp.ByQuery),
		chromedp.WaitReady(`#tresc_wiadomosci`, chromedp.ByID),
		chromedp.Evaluate(`
        let textarea = document.getElementById('tresc_wiadomosci');
        let originalText = textarea.value;
        let newText = '`+text+`';
        textarea.value = newText + originalText;
    `, nil),
		chromedp.Sleep(1*time.Second),
		logAction("Click on Wyślij"),
		chromedp.Click(`input[type="submit"][name="wyslij"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return err
	}
	return nil
}
