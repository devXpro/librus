package mongo

import (
	"testing"
	"time"

	"librus/model"
	"librus/mongo/client"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestMessageDeliveryFunctions tests the unified message delivery tracking functions
func TestMessageDeliveryFunctions(t *testing.T) {
	// Skip if no MongoDB connection
	if client.Db == nil {
		t.Skip("MongoDB not available for testing")
	}

	// Create test data
	telegramUserID := primitive.NewObjectID().Hex()
	messageID := "test_message_id_123"

	// Test 1: Initially message should not be sent
	if IsMessageSentToUser(telegramUserID, messageID) {
		t.Error("Message should not be marked as sent initially")
	}

	// Test 2: Mark message as sent
	err := MarkMessageAsSent(telegramUserID, messageID)
	if err != nil {
		t.Fatalf("Failed to mark message as sent: %v", err)
	}

	// Test 3: Now message should be marked as sent
	if !IsMessageSentToUser(telegramUserID, messageID) {
		t.Error("Message should be marked as sent after MarkMessageAsSent")
	}

	// Test 4: Test with different message (news)
	newsID := "test_news_id_456"

	if IsMessageSentToUser(telegramUserID, newsID) {
		t.Error("News should not be marked as sent initially")
	}

	err = MarkMessageAsSent(telegramUserID, newsID)
	if err != nil {
		t.Fatalf("Failed to mark news as sent: %v", err)
	}

	if !IsMessageSentToUser(telegramUserID, newsID) {
		t.Error("News should be marked as sent after MarkMessageAsSent")
	}

	// Cleanup - remove test data
	cleanupTestData(telegramUserID, messageID, newsID)
}

func cleanupTestData(telegramUserID string, messageIDs ...string) {
	// Clean up message delivery records
	client.Db.Collection("user_message_delivery").DeleteMany(nil, map[string]interface{}{
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
