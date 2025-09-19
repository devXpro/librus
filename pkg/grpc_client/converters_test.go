package grpc_client

import (
	"librus/model"
	pb "librus/proto"
	"testing"
	"time"
)

func TestConvertPbMessageToModel(t *testing.T) {
	// Test with nil message
	result := convertPbMessageToModel(nil)
	if result.Id != "" {
		t.Errorf("Expected empty message for nil input, got %+v", result)
	}

	// Test with valid protobuf message
	now := time.Now()
	pbMsg := &pb.Message{
		Id:             "test-id",
		Type:           pb.MessageType_MESSAGE_TYPE_MESSAGE,
		Link:           "https://test.com",
		Author:         "Test Author",
		Title:          "Test Title",
		Content:        "Test Content",
		DateTimestamp:  now.Unix(),
		AttachmentsDir: "test-uuid-123",
	}

	result = convertPbMessageToModel(pbMsg)

	if result.Id != "test-id" {
		t.Errorf("Expected Id 'test-id', got %s", result.Id)
	}

	if result.Type != model.MsgTypeMessage {
		t.Errorf("Expected Type %s, got %s", model.MsgTypeMessage, result.Type)
	}

	if result.Link != "https://test.com" {
		t.Errorf("Expected Link 'https://test.com', got %s", result.Link)
	}

	if result.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got %s", result.Author)
	}

	if result.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got %s", result.Title)
	}

	if result.Content != "Test Content" {
		t.Errorf("Expected Content 'Test Content', got %s", result.Content)
	}

	// Check date conversion (allow 1 second tolerance)
	if abs(result.Date.Unix()-now.Unix()) > 1 {
		t.Errorf("Expected Date around %v, got %v", now, result.Date)
	}

	// Check attachments directory conversion
	if result.AttachmentsDir == "" {
		t.Errorf("Expected AttachmentsDir to be converted to full path, got empty string")
	}
}

func TestConvertPbMessagesToModel(t *testing.T) {
	// Test with empty slice
	result := convertPbMessagesToModel([]*pb.Message{})
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %d messages", len(result))
	}

	// Test with nil slice
	result = convertPbMessagesToModel(nil)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for nil input, got %d messages", len(result))
	}

	// Test with valid messages
	pbMsgs := []*pb.Message{
		{
			Id:            "msg1",
			Type:          pb.MessageType_MESSAGE_TYPE_MESSAGE,
			Title:         "Message 1",
			DateTimestamp: time.Now().Unix(),
		},
		{
			Id:            "msg2",
			Type:          pb.MessageType_MESSAGE_TYPE_NOTIFICATION,
			Title:         "Message 2",
			DateTimestamp: time.Now().Unix(),
		},
	}

	result = convertPbMessagesToModel(pbMsgs)

	if len(result) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(result))
	}

	if result[0].Id != "msg1" {
		t.Errorf("Expected first message Id 'msg1', got %s", result[0].Id)
	}

	if result[1].Id != "msg2" {
		t.Errorf("Expected second message Id 'msg2', got %s", result[1].Id)
	}
}

func TestConvertModelMessageToPb(t *testing.T) {
	now := time.Now()
	modelMsg := model.Message{
		Id:             "test-id",
		Type:           model.MsgTypeMessage,
		Link:           "https://test.com",
		Author:         "Test Author",
		Title:          "Test Title",
		Content:        "Test Content",
		Date:           now,
		AttachmentsDir: "/path/to/attachments",
	}

	result := convertModelMessageToPb(modelMsg)

	if result.Id != "test-id" {
		t.Errorf("Expected Id 'test-id', got %s", result.Id)
	}

	if result.Type != pb.MessageType_MESSAGE_TYPE_MESSAGE {
		t.Errorf("Expected Type MESSAGE_TYPE_MESSAGE, got %v", result.Type)
	}

	if result.DateTimestamp != now.Unix() {
		t.Errorf("Expected DateTimestamp %d, got %d", now.Unix(), result.DateTimestamp)
	}
}

// Helper function to calculate absolute difference
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
