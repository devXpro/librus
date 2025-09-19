package message

import (
	"errors"
	"log"
	"strings"

	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/telegram/localization"
	"librus/telegram/router"
	"librus/translator"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ReplyHandler handles replies to messages
type ReplyHandler struct{}

// Handle processes reply messages
func (h *ReplyHandler) Handle(ctx *router.Context) error {
	// Check if this is a reply to a message
	if ctx.Update.Original.Message == nil || ctx.Update.Original.Message.ReplyToMessage == nil {
		return nil // Not a reply, let other handlers process it
	}

	// Check if telegram user is authenticated and has Librus account
	if ctx.User == nil || ctx.User.State != model.StateAuthenticated || ctx.User.LibrusLogin == "" {
		return ctx.SendMessage(localization.MsgPleaseLogin)
	}

	// Get Librus account
	librusAccount, err := mongo.FindLibrusAccount(ctx.User.LibrusLogin)
	if err != nil || librusAccount == nil {
		return ctx.SendMessage(localization.MsgPleaseLogin)
	}

	replyToMessage := ctx.Update.Original.Message.ReplyToMessage
	link, err := h.extractLink(replyToMessage)
	if err != nil {
		return ctx.SendMessage(localization.MsgCannotReply)
	}

	// Check if it's a message we can reply to
	if !strings.Contains(link, "wiadomosci") {
		return ctx.SendMessage(localization.MsgCannotReply)
	}

	// Show processing message
	err = ctx.SendMessage(localization.MsgProcessing)
	if err != nil {
		log.Printf("Error sending processing message: %v", err)
	}

	// Create gRPC client
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		return ctx.SendMessage(localization.MsgServiceError)
	}
	defer client.Close()

	// Prepare text for sending
	text := ctx.Update.Data
	if ctx.User.Language != "" {
		// Translate to Polish for sending to Librus
		text, err = translator.TranslateText("pl", text)
		if err != nil {
			log.Printf("Translation error: %v", err)
			return ctx.SendMessage(localization.MsgSomethingWrong)
		}
	}

	// Send reply
	err = client.AnswerMessage(librusAccount.Login, librusAccount.Password, link, text)
	if err != nil {
		log.Printf("Failed to answer message: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	return ctx.SendMessage(localization.MsgMessageSent)
}

// extractLink extracts URL from message entities
func (h *ReplyHandler) extractLink(message *tgbotapi.Message) (string, error) {
	for _, entity := range message.Entities {
		if entity.Type == "text_link" {
			return entity.URL, nil
		}
	}
	return "", errors.New("no link found")
}
