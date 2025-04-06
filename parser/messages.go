package parser

import (
	"context"
	"fmt"
	"librus/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func GetMessages(ctx context.Context) ([]model.Message, error) {
	err := chromedp.Run(ctx, chromedp.Navigate(`https://synergia.librus.pl/wiadomosci`),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		logAction("Navigating to wiadomosci page"),
		chromedp.Sleep(3*time.Second))
	if err != nil {
		return nil, err
	}

	linksCtx, linksCancel := context.WithTimeout(ctx, time.Second)
	defer linksCancel()
	// Get HTML code of the page
	var html string
	if err = chromedp.Run(linksCtx, chromedp.InnerHTML(`html`, &html)); err != nil {
		return nil, err
	}
	printLog("Table received! Loading HTML code into goquery.Document object")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	printLog("Looking for all links in the table with 'decorated' class...")
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
	printLog("Links ready!")
	var messages []model.Message
	printLog("Spin this shit!")
	links = removeDuplicates(links)
	for _, link := range links {
		link = "https://synergia.librus.pl" + link
		if strings.Contains(link, "javascript") {
			continue
		}

		message, err := GetSingleMessage(ctx, link)
		if err != nil {
			printLog(fmt.Sprintf("Error processing message %s: %v", link, err))
			continue // Skip this message and continue with others
		}

		messages = append(messages, message)
	}
	printLog("Links processed...")
	return messages, nil
}

// GetSingleMessage retrieves a single message from the specified URL
// This can be used by other parts of the application to get a message by its direct URL
func GetSingleMessage(ctx context.Context, link string) (model.Message, error) {
	err := chromedp.Run(ctx, chromedp.Navigate(link))
	printLog(link)
	if err != nil {
		return model.Message{}, err
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
		return model.Message{}, err
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
		return model.Message{}, err
	}

	var content string
	err = chromedp.Run(ctx, chromedp.InnerHTML(`.container-message-content`, &content, chromedp.ByQuery))
	if err != nil {
		return model.Message{}, err
	}

	message := model.Message{Link: link, Content: content, Date: date, Title: title, Author: author, Type: model.MsgTypeMessage}
	message.GenerateId()

	// Download attachments if any
	attachmentsDir, err := DownloadAttachments(ctx, "")
	if err != nil {
		// If there's an error downloading attachments, log it but don't interrupt the process
		printLog(fmt.Sprintf("Error downloading attachments: %v", err))
	} else if attachmentsDir != "" {
		// If attachments were successfully downloaded, save the path
		message.AttachmentsDir = attachmentsDir
		printLog(fmt.Sprintf("Downloaded attachments to: %s", attachmentsDir))
	}

	return message, nil
}
