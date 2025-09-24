package mongo

import (
	"testing"
	"time"

	"librus/model"
	"librus/mongo/client"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewsStatusFunctions tests the news status tracking functions
func TestNewsStatusFunctions(t *testing.T) {
	// Skip if no MongoDB connection
	if client.Db == nil {
		t.Skip("MongoDB not available for testing")
	}

	// Create test data
	telegramUserID := primitive.NewObjectID().Hex()
	newsID := "test_news_id_123"

	// Test 1: Initially news should not be sent
	if IsNewsSentToUser(telegramUserID, newsID) {
		t.Error("News should not be marked as sent initially")
	}

	// Test 2: Mark news as sent
	err := MarkNewsAsSent(telegramUserID, newsID)
	if err != nil {
		t.Fatalf("Failed to mark news as sent: %v", err)
	}

	// Test 3: Now news should be marked as sent
	if !IsNewsSentToUser(telegramUserID, newsID) {
		t.Error("News should be marked as sent after MarkNewsAsSent")
	}

	// Test 4: Test type-aware functions with notification type
	messageID := "test_message_id_456"
	
	// Should use news status for notifications
	if IsMessageSentToUserByType(telegramUserID, messageID, model.MsgTypeNotification) {
		t.Error("Notification should not be marked as sent initially")
	}

	err = MarkMessageAsSentByType(telegramUserID, messageID, model.MsgTypeNotification)
	if err != nil {
		t.Fatalf("Failed to mark notification as sent: %v", err)
	}

	if !IsMessageSentToUserByType(telegramUserID, messageID, model.MsgTypeNotification) {
		t.Error("Notification should be marked as sent after MarkMessageAsSentByType")
	}

	// Test 5: Test type-aware functions with regular message type
	regularMessageID := "test_regular_message_789"
	
	// Should use regular message status for messages
	if IsMessageSentToUserByType(telegramUserID, regularMessageID, model.MsgTypeMessage) {
		t.Error("Regular message should not be marked as sent initially")
	}

	err = MarkMessageAsSentByType(telegramUserID, regularMessageID, model.MsgTypeMessage)
	if err != nil {
		t.Fatalf("Failed to mark regular message as sent: %v", err)
	}

	if !IsMessageSentToUserByType(telegramUserID, regularMessageID, model.MsgTypeMessage) {
		t.Error("Regular message should be marked as sent after MarkMessageAsSentByType")
	}

	// Cleanup - remove test data
	cleanupTestData(telegramUserID, newsID, messageID, regularMessageID)
}

func cleanupTestData(telegramUserID, newsID, messageID, regularMessageID string) {
	// Clean up news status
	client.Db.Collection("user_news_status").DeleteMany(nil, map[string]interface{}{
		"telegram_user_id": telegramUserID,
	})
	
	// Clean up message status
	client.Db.Collection("user_message_status").DeleteMany(nil, map[string]interface{}{
		"telegram_user_id": telegramUserID,
	})
}

// TestMessageIDGeneration tests that news generate the same ID for same content
func TestMessageIDGeneration(t *testing.T) {
	// Create two identical news messages
	news1 := model.Message{
		Type:    model.MsgTypeNotification,
		Title:   "Test News Title",
		Content: "Test news content",
		Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	news2 := model.Message{
		Type:    model.MsgTypeNotification,
		Title:   "Test News Title",
		Content: "Test news content",
		Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	// Generate IDs
	news1.GenerateId()
	news2.GenerateId()

	// They should have the same ID
	if news1.Id != news2.Id {
		t.Errorf("Identical news should have the same ID. Got %s and %s", news1.Id, news2.Id)
	}

	// Create a different news message
	news3 := model.Message{
		Type:    model.MsgTypeNotification,
		Title:   "Different News Title",
		Content: "Test news content",
		Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	news3.GenerateId()

	// It should have a different ID
	if news1.Id == news3.Id {
		t.Error("Different news should have different IDs")
	}
}
