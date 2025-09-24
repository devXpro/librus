package model

import (
	"time"
)

// UserState represents the current state of user interaction
type UserState string

const (
	StateLanguageSelection UserState = "language_selection"
	StateAwaitingLogin     UserState = "awaiting_login"
	StateAwaitingPassword  UserState = "awaiting_password"
	StateAuthenticated     UserState = "authenticated"
	StateAwaitingURL       UserState = "awaiting_url"
)

// LibrusAccount represents a Librus school account
type LibrusAccount struct {
	Login    string `bson:"_id"`      // Login is unique, so we use it as _id
	Password string `bson:"password"` // Encrypted password
}

// TelegramUser represents a Telegram user with individual settings
type TelegramUser struct {
	Id           string    `bson:"_id"`            // Auto-generated ID
	TelegramID   int64     `bson:"telegram_id"`    // Telegram chat ID (unique)
	LibrusLogin  string    `bson:"librus_login"`   // Reference to LibrusAccount._id
	Language     string    `bson:"language"`       // User's language preference
	State        UserState `bson:"state"`          // Current interaction state
	CreatedAt    time.Time `bson:"created_at"`     // When user was created
	LastActiveAt time.Time `bson:"last_active_at"` // Last interaction time
}

// UserMessageStatus tracks which messages were sent to which users
type UserMessageStatus struct {
	Id             string    `bson:"_id"`              // Auto-generated ID
	TelegramUserID string    `bson:"telegram_user_id"` // Reference to TelegramUser._id
	MessageID      string    `bson:"message_id"`       // Reference to Message._id
	SentAt         time.Time `bson:"sent_at"`          // When message was sent
}

// UserNewsStatus tracks which news were sent to which users
// This is separate from UserMessageStatus to handle news properly
// since news have the same ID for all users in the same school
type UserNewsStatus struct {
	Id             string    `bson:"_id"`              // Auto-generated ID
	TelegramUserID string    `bson:"telegram_user_id"` // Reference to TelegramUser._id
	NewsID         string    `bson:"news_id"`          // Reference to Message._id (for news only)
	SentAt         time.Time `bson:"sent_at"`          // When news was sent
}
