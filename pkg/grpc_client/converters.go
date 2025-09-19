package grpc_client

import (
	"librus/model"
	"librus/pkg/config"
	pb "librus/proto"
	"time"
)

// convertPbMessageToModel converts a protobuf Message to model.Message
func convertPbMessageToModel(pbMsg *pb.Message) model.Message {
	if pbMsg == nil {
		return model.Message{}
	}

	// Convert protobuf MessageType to model.MessageType
	var msgType model.MessageType
	switch pbMsg.Type {
	case pb.MessageType_MESSAGE_TYPE_MESSAGE:
		msgType = model.MsgTypeMessage
	case pb.MessageType_MESSAGE_TYPE_NOTIFICATION:
		msgType = model.MsgTypeNotification
	case pb.MessageType_MESSAGE_TYPE_NEWS:
		msgType = model.MsgTypeNews
	default:
		msgType = model.MsgTypeMessage // Default fallback
	}

	// Convert Unix timestamp to time.Time
	date := time.Unix(pbMsg.DateTimestamp, 0)

	// Handle attachments directory
	// If pbMsg.AttachmentsDir contains a UUID, convert it to full path
	var attachmentsDir string
	if pbMsg.AttachmentsDir != "" && pbMsg.AttachmentsDir != "nil" {
		// Check if it's already a full path or just a UUID
		if pbMsg.AttachmentsDir[0] == '/' || pbMsg.AttachmentsDir[0] == '.' {
			// Already a full path
			attachmentsDir = pbMsg.AttachmentsDir
		} else {
			// Assume it's a UUID, convert to full path
			attachmentsDir = config.GetAttachmentsDirPath(pbMsg.AttachmentsDir)
		}
	}

	msg := model.Message{
		Id:             pbMsg.Id,
		Type:           msgType,
		Link:           pbMsg.Link,
		Author:         pbMsg.Author,
		Title:          pbMsg.Title,
		Content:        pbMsg.Content,
		Date:           date,
		AttachmentsDir: attachmentsDir,
		// UserID will be set by the caller
	}

	// Generate ID if not provided
	if msg.Id == "" {
		msg.GenerateId()
	}

	return msg
}

// convertPbMessagesToModel converts a slice of protobuf Messages to model.Messages
func convertPbMessagesToModel(pbMsgs []*pb.Message) []model.Message {
	if len(pbMsgs) == 0 {
		return []model.Message{}
	}

	messages := make([]model.Message, 0, len(pbMsgs))
	for _, pbMsg := range pbMsgs {
		if pbMsg != nil {
			messages = append(messages, convertPbMessageToModel(pbMsg))
		}
	}

	return messages
}

// convertModelMessageToPb converts a model.Message to protobuf Message (if needed for future use)
func convertModelMessageToPb(msg model.Message) *pb.Message {
	// Convert model.MessageType to protobuf MessageType
	var pbType pb.MessageType
	switch msg.Type {
	case model.MsgTypeMessage:
		pbType = pb.MessageType_MESSAGE_TYPE_MESSAGE
	case model.MsgTypeNotification:
		pbType = pb.MessageType_MESSAGE_TYPE_NOTIFICATION
	case model.MsgTypeNews:
		pbType = pb.MessageType_MESSAGE_TYPE_NEWS
	default:
		pbType = pb.MessageType_MESSAGE_TYPE_UNSPECIFIED
	}

	// Convert attachments directory to UUID if it's a full path
	var attachmentsDir string
	if msg.AttachmentsDir != "" && msg.AttachmentsDir != "nil" {
		// Extract UUID from full path if needed
		// For now, just pass as-is
		attachmentsDir = msg.AttachmentsDir
	}

	return &pb.Message{
		Id:             msg.Id,
		Type:           pbType,
		Link:           msg.Link,
		Author:         msg.Author,
		Title:          msg.Title,
		Content:        msg.Content,
		DateTimestamp:  msg.Date.Unix(),
		AttachmentsDir: attachmentsDir,
	}
}
