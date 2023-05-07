package parser

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"librus/model"
	"strings"
	"time"
)

func GetMessages(ctx context.Context) ([]model.Message, error) {
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
	// For always all messages
	//itemsSelector := "table.decorated td a"
	itemsSelector := "table.decorated td[style=\"font-weight: bold;\"] a"
	doc.Find(itemsSelector).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok {
			links = append(links, href)
		}
	})
	printLog("линки готовы!")
	var messages []model.Message
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
		message := model.Message{Link: link, Content: content, Date: date, Title: title, Author: author, Type: model.MsgTypeMessage}
		message.GenerateId()
		messages = append(messages, message)
	}
	printLog("Линки обработаны...")
	return messages, nil
}
