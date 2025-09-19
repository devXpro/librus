package state

import (
	"fmt"
	"log"
	"strings"

	"librus/model"
	"librus/mongo"
	"librus/pkg/grpc_client"
	"librus/telegram/keyboard"
	"librus/telegram/localization"
	"librus/telegram/router"
)

// LoginHandler handles login state messages
type LoginHandler struct{}

// Handle processes messages when user is in login state
func (h *LoginHandler) Handle(ctx *router.Context) error {
	switch ctx.User.State {
	case model.StateAwaitingLogin:
		return h.handleLoginInput(ctx)
	case model.StateAwaitingPassword:
		return h.handlePasswordInput(ctx)
	default:
		log.Printf("LoginHandler called with invalid state: %s", ctx.User.State)
		return nil
	}
}

// handleLoginInput processes login input
func (h *LoginHandler) handleLoginInput(ctx *router.Context) error {
	login := strings.TrimSpace(ctx.Update.Data)

	if login == "" {
		return ctx.SendMessage(localization.MsgEnterLogin)
	}

	// Store login temporarily (we'll validate it with password)
	// For now, we'll store it in a temporary field or use a different approach
	// Let's update the user with the login and move to password state

	// Update user login
	err := mongo.UpdateUserFieldByTelegramID(ctx.Update.ChatID, "login", login)
	if err != nil {
		log.Printf("Error updating user login: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	// Update state to awaiting password
	err = mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAwaitingPassword)
	if err != nil {
		log.Printf("Error updating user state: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	// Ask for password
	cancelKeyboard := keyboard.CancelKeyboard(ctx.Localization)
	return ctx.SendMessageWithKeyboard(localization.MsgEnterPassword, cancelKeyboard)
}

// handlePasswordInput processes password input and validates credentials
func (h *LoginHandler) handlePasswordInput(ctx *router.Context) error {
	password := strings.TrimSpace(ctx.Update.Data)

	if password == "" {
		return ctx.SendMessage(localization.MsgEnterPassword)
	}

	// Get current user to get the login
	user, err := mongo.FindUserByTelegramID(ctx.Update.ChatID)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		return ctx.SendMessage(localization.MsgSomethingWrong)
	}

	if user.Login == "" {
		// Something went wrong, restart login process
		err = mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAwaitingLogin)
		if err != nil {
			log.Printf("Error updating user state: %v", err)
		}
		return ctx.SendMessage(localization.MsgEnterLogin)
	}

	// Show processing message
	err = ctx.SendMessage(localization.MsgProcessing)
	if err != nil {
		log.Printf("Error sending processing message: %v", err)
	}

	// Validate credentials with gRPC
	if h.validateCredentials(user.Login, password) {
		// Success - update user with password and set authenticated state
		err = mongo.UpdateUserFieldByTelegramID(ctx.Update.ChatID, "password", password)
		if err != nil {
			log.Printf("Error updating user password: %v", err)
			return ctx.SendMessage(localization.MsgSomethingWrong)
		}

		err = mongo.UpdateUserStateByTelegramID(ctx.Update.ChatID, model.StateAuthenticated)
		if err != nil {
			log.Printf("Error updating user state: %v", err)
			return ctx.SendMessage(localization.MsgSomethingWrong)
		}

		// Show success message
		err = ctx.SendMessage(localization.MsgLoginSuccess)
		if err != nil {
			return err
		}

		// Show main menu
		mainMenu := keyboard.MainMenuKeyboard(ctx.Localization)
		return ctx.SendMessageWithKeyboard(localization.MsgMainMenu, mainMenu)

	} else {
		// Failed - ask to try again
		cancelKeyboard := keyboard.CancelKeyboard(ctx.Localization)
		return ctx.SendMessageWithKeyboard(localization.MsgInvalidCredentials, cancelKeyboard)
	}
}

// validateCredentials validates login credentials using gRPC
func (h *LoginHandler) validateCredentials(login, password string) bool {
	if login == "" || password == "" {
		return false
	}

	client, err := grpc_client.NewLibrusScraperClient()
	if err != nil {
		fmt.Printf("Failed to create gRPC client: %v\n", err)
		return false
	}
	defer client.Close()

	err = client.ValidateLogin(login, password)
	if err != nil {
		fmt.Printf("Login validation failed: %v\n", err)
		return false
	}

	return true
}
