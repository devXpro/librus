package parser

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"librus/model"
	"strings"
	"time"
)

func GetNews(ctx context.Context) ([]model.Message, error) {
	const url = "https://synergia.librus.pl/ogloszenia"
	var messages []model.Message

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
		message := model.Message{
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
			UserID: "0",
		}
		message.GenerateId()
		messages = append(messages, message)
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return messages, nil
}
