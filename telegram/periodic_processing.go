package telegram

import (
	"sort"
	"time"

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

		logger.Debug("Getting Librus accounts from database")
		accounts := mongo.GetLibrusAccountsFromDatabase()
		logger.Debug("Retrieved accounts", zap.Int("count", len(accounts)))

		for _, account := range accounts {
			logger.Debug("Processing account", zap.String("login", account.Login))
			// Use gRPC GetAllUpdates to get both messages and news in one call
			logger.Debug("Calling gRPC GetAllUpdates", zap.String("login", account.Login))
			msgs, news, err := client.GetAllUpdates(account.Login, account.Password)
			if err != nil {
				logger.ErrorWithError("Failed to get updates for account", err,
					zap.String("login", account.Login),
				)
				continue
			}
			logger.Debug("Received updates from gRPC",
				zap.String("login", account.Login),
				zap.Int("messages_count", len(msgs)),
				zap.Int("news_count", len(news)),
			)

			// Combine messages and news
			allMsgs := append(msgs, news...)
			if len(allMsgs) == 0 {
				logger.Debug("No new messages for account", zap.String("login", account.Login))
				continue
			}
			logger.Debug("Total messages to process",
				zap.String("login", account.Login),
				zap.Int("total_count", len(allMsgs)),
			)
			logger.Debug("Adding messages to database",
				zap.String("login", account.Login),
				zap.Int("messages_count", len(allMsgs)),
			)
			err = mongo.AddMessagesToDatabase(allMsgs)
			if err != nil {
				logger.ErrorWithError("Failed to add messages to database", err,
					zap.String("login", account.Login),
				)
				continue
			}
			logger.Debug("Messages added to database",
				zap.String("login", account.Login),
				zap.Int("messages_count", len(allMsgs)),
			)

			sort.Slice(allMsgs, func(i, j int) bool {
				return allMsgs[i].Date.Before(allMsgs[j].Date)
			})

			// Get all telegram users for this Librus account
			logger.Debug("Getting telegram users for account", zap.String("login", account.Login))
			telegramUsers, err := mongo.GetTelegramUsersByLibrusLogin(account.Login)
			if err != nil {
				logger.ErrorWithError("Failed to get telegram users for account", err,
					zap.String("login", account.Login),
				)
				continue
			}
			logger.Debug("Retrieved telegram users",
				zap.String("login", account.Login),
				zap.Int("users_count", len(telegramUsers)),
			)

			// Send messages to each telegram user
			logger.Debug("Starting to send messages to users",
				zap.String("login", account.Login),
				zap.Int("messages_count", len(allMsgs)),
				zap.Int("users_count", len(telegramUsers)),
			)
			for _, message := range allMsgs {
				for _, telegramUser := range telegramUsers {
					// Check if message was already sent to this user
					if mongo.IsMessageSentToUser(telegramUser.Id, message.Id) {
						logger.Debug("Message already sent to user, skipping",
							zap.String("user_id", telegramUser.Id),
							zap.String("message_id", message.Id),
							zap.String("message_type", string(message.Type)),
						)
						continue
					}

					logger.Debug("Sending message to user",
						zap.String("user_id", telegramUser.Id),
						zap.String("message_id", message.Id),
						zap.String("message_title", message.Title),
						zap.String("message_type", string(message.Type)),
					)

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
							zap.String("message_type", string(message.Type)),
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
			logger.Debug("Finished processing account", zap.String("login", account.Login))
		}
		logger.Info("Finished processing all accounts, waiting for next interval")
	}
}
