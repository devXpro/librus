package telegram

import (
	"sort"
	"time"

	"librus/model"
	"librus/mongo"
	"librus/pkg/config"
	"librus/pkg/grpc_client"
	"librus/pkg/logger"
	"librus/telegram/channel"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func checkNewLibrusMessagesPeriodically(bot *tgbotapi.BotAPI) {
	// Create gRPC client once and reuse it
	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		logger.ErrorWithError("Failed to create gRPC client", err)
		return
	}
	defer client.Close()

	// Get configurable check interval
	checkInterval := config.GetMessageCheckInterval()
	logger.Info("Message check interval configured", zap.Duration("interval", checkInterval))

	for {
		select {
		case <-time.After(checkInterval):
			logger.Info("Starting periodic message check")
		case <-channel.UpdateNow:
			logger.Info("Starting forced message update")
		}
		accounts := mongo.GetLibrusAccountsFromDatabase()
		for _, account := range accounts {
			// Use gRPC GetAllUpdates to get both messages and news in one call
			msgs, news, err := client.GetAllUpdates(account.Login, account.Password)
			if err != nil {
				logger.ErrorWithError("Failed to get updates for account", err,
					zap.String("login", account.Login),
				)
				continue
			}

			// Combine messages and news
			allMsgs := append(msgs, news...)
			if len(allMsgs) == 0 {
				continue
			}
			allMsgs = addLibrusLoginToMessages(allMsgs, account.Login)

			allMsgs, err = mongo.AddMessagesToDatabase(allMsgs, account.Login)
			if err != nil {
				logger.ErrorWithError("Failed to add messages to database", err,
					zap.String("login", account.Login),
				)
				continue
			}

			sort.Slice(allMsgs, func(i, j int) bool {
				return allMsgs[i].Date.Before(allMsgs[j].Date)
			})

			// Get all telegram users for this Librus account
			telegramUsers, err := mongo.GetTelegramUsersByLibrusLogin(account.Login)
			if err != nil {
				logger.ErrorWithError("Failed to get telegram users for account", err,
					zap.String("login", account.Login),
				)
				continue
			}

			// Send messages to each telegram user
			for _, message := range allMsgs {
				for _, telegramUser := range telegramUsers {
					// Check if message was already sent to this user
					if mongo.IsMessageSentToUser(telegramUser.Id, message.Id) {
						continue
					}

					// Translate message if user has language preference
					translatedMessage := message
					if telegramUser.Language != "" {
						translatedMessage.Translate(telegramUser.Language)
					}

					// Send message
					err = translatedMessage.Send(bot, telegramUser.TelegramID)
					if err != nil {
						logger.ErrorWithError("Error sending message to user", err,
							zap.String("user_id", telegramUser.Id),
							zap.Int64("telegram_id", telegramUser.TelegramID),
						)
						continue
					}

					// Mark message as sent
					err = mongo.MarkMessageAsSent(telegramUser.Id, message.Id)
					if err != nil {
						logger.ErrorWithError("Error marking message as sent", err,
							zap.String("user_id", telegramUser.Id),
							zap.String("message_id", message.Id),
						)
					}
				}

				// Clean up attachments directory after sending to all users
				if err := message.CleanupAttachments(); err != nil {
					logger.ErrorWithError("Error cleaning up attachments", err,
						zap.String("message_id", message.Id),
						zap.String("attachments_dir", message.AttachmentsDir),
					)
				}
			}
		}
	}
}

func addLibrusLoginToMessages(msgs []model.Message, login string) []model.Message {
	var result []model.Message
	for _, msg := range msgs {
		msg.LibrusLogin = login
		result = append(result, msg)
	}
	return result
}
