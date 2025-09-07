package main

import (
	"testing"
	"time"

	pb "librus/proto"
)

func TestProtobufGeneration(t *testing.T) {
	// Test creating a protobuf Message
	msg := &pb.Message{
		Id:             "test-id",
		Type:           pb.MessageType_MESSAGE_TYPE_MESSAGE,
		Link:           "https://test.com",
		Author:         "Test Author",
		Title:          "Test Title",
		Content:        "Test Content",
		DateTimestamp:  time.Now().Unix(),
		AttachmentsDir: "/tmp/attachments",
	}

	if msg.Id != "test-id" {
		t.Errorf("Expected Id to be 'test-id', got %s", msg.Id)
	}

	if msg.Type != pb.MessageType_MESSAGE_TYPE_MESSAGE {
		t.Errorf("Expected Type to be MESSAGE_TYPE_MESSAGE, got %v", msg.Type)
	}

	// Test creating a GetMessagesRequest
	req := &pb.GetMessagesRequest{
		Login:    "test-login",
		Password: "test-password",
	}

	if req.Login != "test-login" {
		t.Errorf("Expected Login to be 'test-login', got %s", req.Login)
	}
}

func TestMessageTypeEnum(t *testing.T) {
	// Test enum values
	if pb.MessageType_MESSAGE_TYPE_MESSAGE.String() != "MESSAGE_TYPE_MESSAGE" {
		t.Errorf("Expected MESSAGE_TYPE_MESSAGE string representation")
	}

	if pb.MessageType_MESSAGE_TYPE_NOTIFICATION.String() != "MESSAGE_TYPE_NOTIFICATION" {
		t.Errorf("Expected MESSAGE_TYPE_NOTIFICATION string representation")
	}
}
