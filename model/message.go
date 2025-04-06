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
	msg.Text += "\n_______________________________"

	// Send message
	_, err := bot.Send(msg)
	if err != nil {
		return err
	}

	// Check if there are attachments to send
	if message.AttachmentsDir != "" && message.AttachmentsDir != "nil" {
		// Get list of files in directory
		files, err := os.ReadDir(message.AttachmentsDir)
		if err != nil {
			return fmt.Errorf("error reading attachments directory: %w", err)
		}

		// Send files
		for _, file := range files {
			if file.IsDir() {
				continue // Skip subdirectories
			}

			filePath := filepath.Join(message.AttachmentsDir, file.Name())

			// Create FileBytes with file
			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				// Log error and continue with other files
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}

			// Determine file type by extension
			fileExt := strings.ToLower(filepath.Ext(file.Name()))

			// Send file depending on its type
			var fileMsg tgbotapi.Chattable

			switch fileExt {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				// Send as photo
				photoMsg := tgbotapi.NewPhoto(telegramId, tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				})
				photoMsg.Caption = fmt.Sprintf("📎 %s", file.Name())
				fileMsg = photoMsg

			case ".mp4", ".mov", ".avi", ".mkv":
				// Send as video
				videoMsg := tgbotapi.NewVideo(telegramId, tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				})
				videoMsg.Caption = fmt.Sprintf("📎 %s", file.Name())
				fileMsg = videoMsg

			default:
				// Send as document for all other types
				docMsg := tgbotapi.NewDocument(telegramId, tgbotapi.FileBytes{
					Name:  file.Name(),
					Bytes: fileBytes,
				})
				docMsg.Caption = fmt.Sprintf("📎 %s", file.Name())
				fileMsg = docMsg
			}

			// Send file
			_, err = bot.Send(fileMsg)
			if err != nil {
				fmt.Printf("Error sending file %s: %v\n", file.Name(), err)
				// Continue with other files
			}
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
