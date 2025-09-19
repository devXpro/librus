package telegram

import (
	"crypto/tls"
	"fmt"
	"librus/helper"
	"librus/model"
	"librus/telegram/handler/callback"
	"librus/telegram/handler/command"
	"librus/telegram/handler/message"
	"librus/telegram/handler/state"
	"librus/telegram/keyboard"
	"librus/telegram/router"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Start() {
	bot := createBot()
	go checkNewLibrusMessagesPeriodically(bot)

	// Create router and register handlers
	r := setupRouter()

	u := createUpdateConfig()
	for update := range bot.GetUpdatesChan(u) {
		// Skip updates we can't handle
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		logReceivedUpdate(update)

		// Handle update with router
		if err := r.HandleUpdate(bot, update); err != nil {
			log.Printf("Error handling update: %v", err)
		}
	}

	select {}
}

// setupRouter creates and configures the router with all handlers
func setupRouter() *router.Router {
	r := router.NewRouter()

	// Register command handlers
	r.RegisterCommand("start", &command.StartHandler{})

	// Register callback handlers
	r.RegisterCallback(keyboard.CallbackLanguageSelect, &callback.LanguageHandler{})
	r.RegisterCallback(keyboard.CallbackLogin, &callback.AuthHandler{})
	r.RegisterCallback(keyboard.CallbackCheckMessages, &callback.MenuHandler{})
	r.RegisterCallback(keyboard.CallbackGetByURL, &callback.MenuHandler{})
	r.RegisterCallback(keyboard.CallbackSettings, &callback.MenuHandler{})
	r.RegisterCallback(keyboard.CallbackHelp, &callback.MenuHandler{})
	r.RegisterCallback(keyboard.CallbackBackToMenu, &callback.MenuHandler{})
	r.RegisterCallback(keyboard.CallbackChangeLanguage, &callback.SettingsHandler{})
	r.RegisterCallback(keyboard.CallbackResetAccount, &callback.SettingsHandler{})
	r.RegisterCallback(keyboard.CallbackConfirmYes, &callback.ConfirmationHandler{})
	r.RegisterCallback(keyboard.CallbackConfirmNo, &callback.ConfirmationHandler{})
	r.RegisterCallback(keyboard.CallbackCancel, &callback.ConfirmationHandler{})

	// Register state handlers
	r.RegisterState(model.StateAwaitingLogin, &state.LoginHandler{})
	r.RegisterState(model.StateAwaitingPassword, &state.LoginHandler{})
	r.RegisterState(model.StateAwaitingURL, &state.URLHandler{})

	// Register message handlers (order matters - first match wins)
	r.RegisterMessage(&message.AdminHandler{})    // Handle admin commands first
	r.RegisterMessage(&message.ReplyHandler{})    // Handle replies second
	r.RegisterMessage(&message.FallbackHandler{}) // Fallback for everything else

	return r
}

func createBot() *tgbotapi.BotAPI {
	httpClient := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	bot, err := tgbotapi.NewBotAPIWithClient(
		helper.GetEnv("TELEGRAM_TOKEN", "token"),
		"https://api.telegram.org/bot%s/%s",
		httpClient,
	)
	if err != nil {
		log.Fatal(err)
	}

	return bot
}

func createUpdateConfig() tgbotapi.UpdateConfig {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 3
	return u
}

func logReceivedUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		fmt.Printf("Received message: %s (from: %d)\n", update.Message.Text, update.Message.From.ID)
		if update.Message.ReplyToMessage != nil {
			fmt.Printf("ReplyToMessage: %s\n", update.Message.ReplyToMessage.Text)
		}
	} else if update.CallbackQuery != nil {
		fmt.Printf("Received callback: %s (from: %d)\n", update.CallbackQuery.Data, update.CallbackQuery.From.ID)
	}
}
