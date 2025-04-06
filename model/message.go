package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"librus/translator"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageType string

const (
	MsgTypeMessage      MessageType = "message"
	MsgTypeNotification MessageType = "notification"
	MsgTypeNews         MessageType = "news"
)

type Message struct {
	Id             string      `bson:"_id"`
	Type           MessageType `bson:"type"`
	Link           string      `bson:"link"`
	Author         string      `bson:"author"`
	Title          string      `bson:"title"`
	Content        string      `bson:"content"`
	Date           time.Time   `bson:"date"`
	UserID         string      `bson:"user_id"`
	AttachmentsDir string      `bson:"attachments_dir,omitempty"` // Path to directory with attachments
}

func (message *Message) GenerateId() {
	stringToHash := ""
	if message.Type == MsgTypeNotification {
		stringToHash = message.Title + message.Content + message.Date.Format("2006-01-02")
	} else if message.Type == MsgTypeMessage {
		stringToHash = message.Link
	} else {
		panic("Wrong message type")
	}
	h := md5.New()
	h.Write([]byte(stringToHash))
	message.Id = hex.EncodeToString(h.Sum(nil))
}
func (message *Message) Translate(lang string) {
	translation, err := translator.TranslateText(lang, message.Title)
	if err == nil {
		message.Title = translation
	}

	translation, err = translator.TranslateText(lang, message.Content)
	if err == nil {
		message.Content = translation
	}

	translation, err = translator.TranslateText(lang, message.Author)
	if err == nil {
		message.Author = translation
	}
}

func (message *Message) Send(bot *tgbotapi.BotAPI, telegramId int64) error {
	// Check if there are attachments to prepare
	var photos []interface{}
	var videos []interface{}
	var documents []tgbotapi.FileBytes
	var fileNames []string

	if message.AttachmentsDir != "" && message.AttachmentsDir != "nil" {
		// Get list of files in directory
		files, err := os.ReadDir(message.AttachmentsDir)
		if err != nil {
			return fmt.Errorf("error reading attachments directory: %w", err)
		}

		// Process files and categorize them
		for _, file := range files {
			if file.IsDir() {
				continue // Skip subdirectories
			}

			filePath := filepath.Join(message.AttachmentsDir, file.Name())
			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}

			// Determine file type by extension
			fileExt := strings.ToLower(filepath.Ext(file.Name()))
			fileNames = append(fileNames, file.Name())

			switch fileExt {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				// Add to photos group
				photos = append(photos, tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				}))
			case ".mp4", ".mov", ".avi", ".mkv":
				// Add to videos group
				videos = append(videos, tgbotapi.NewInputMediaVideo(tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				}))
			default:
				// Add to documents list
				documents = append(documents, tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				})
			}
		}
	}

	// Prepare message text with attachment info if files exist
	var attachmentInfo string
	if len(fileNames) > 0 {
		attachmentInfo = "\n\n📎 Attachments: " + strings.Join(fileNames, ", ")
	}

	// Create the text message
	msg := tgbotapi.NewMessage(telegramId, "")
	msg.ParseMode = tgbotapi.ModeHTML

	// Add information about message type
	var typeIcon string
	switch message.Type {
	case MsgTypeMessage:
		typeIcon = "📩"
	case MsgTypeNotification:
		typeIcon = "🔔"
	default:
		typeIcon = ""
	}
	msg.Text += fmt.Sprintf("%s <b><a href=\"%s\">%s</a></b>\n\n", typeIcon, message.Link, message.Title)

	// Add information about the author
	msg.Text += fmt.Sprintf("👤 <i>From: %s</i>\n\n", message.Author)

	// Add main text
	msg.Text += "📝 " + replaceBrTags(message.Content) + "\n\n"

	// Add date and time
	msg.Text += fmt.Sprintf("📅 %s", message.Date.Format("02.01.2006 15:04"))

	// Add attachment info if available
	msg.Text += attachmentInfo

	msg.Text += "\n_______________________________"

	// Send text message
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}

	// Send photo group if any
	if len(photos) > 0 {
		mediaGroup := tgbotapi.MediaGroupConfig{
			ChatID: telegramId,
			Media:  photos,
		}
		_, err = bot.SendMediaGroup(mediaGroup)
		if err != nil {
			fmt.Printf("Error sending photo group: %v\n", err)
		}
	}

	// Send video group if any
	if len(videos) > 0 {
		mediaGroup := tgbotapi.MediaGroupConfig{
			ChatID: telegramId,
			Media:  videos,
		}
		_, err = bot.SendMediaGroup(mediaGroup)
		if err != nil {
			fmt.Printf("Error sending video group: %v\n", err)
		}
	}

	// Send documents separately if any
	for _, doc := range documents {
		docMsg := tgbotapi.NewDocument(telegramId, doc)
		_, err = bot.Send(docMsg)
		if err != nil {
			fmt.Printf("Error sending document %s: %v\n", doc.Name, err)
		}
	}

	return nil
}

func replaceBrTags(s string) string {
	return strings.ReplaceAll(s, "<br>", "")
}

// CleanupAttachments removes the attachments directory if it exists
func (message *Message) CleanupAttachments() error {
	if message.AttachmentsDir != "" && message.AttachmentsDir != "nil" {
		err := os.RemoveAll(message.AttachmentsDir)
		if err != nil {
			return fmt.Errorf("error removing attachments directory: %w", err)
		}
		message.AttachmentsDir = ""
	}
	return nil
}
