package grpc_client

import (
	"context"
	"fmt"
	"librus/model"
	"librus/pkg/config"
	pb "librus/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LibrusScraperClient wraps the gRPC client and provides methods that return model.Message
type LibrusScraperClient struct {
	conn   *grpc.ClientConn
	client pb.LibrusScraperClient
}

// NewLibrusScraperClient creates a new gRPC client
func NewLibrusScraperClient() (*LibrusScraperClient, error) {
	address := config.GetGRPCScraperAddress()

	// Set up connection with insecure credentials for now
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", address, err)
	}

	client := pb.NewLibrusScraperClient(conn)

	return &LibrusScraperClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *LibrusScraperClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateLogin validates user credentials (replaces parser.Login)
func (c *LibrusScraperClient) ValidateLogin(login, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.ValidateLoginRequest{
		Login:    login,
		Password: password,
	}

	_, err := c.client.ValidateLogin(ctx, req)
	if err != nil {
		return fmt.Errorf("login validation failed: %w", err)
	}

	return nil
}

// GetMessages retrieves messages (replaces parser.GetMessages)
func (c *LibrusScraperClient) GetMessages(login, password string) ([]model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := &pb.GetMessagesRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.client.GetMessages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return convertPbMessagesToModel(resp.Messages), nil
}

// GetNews retrieves news/announcements (replaces parser.GetNews)
func (c *LibrusScraperClient) GetNews(login, password string) ([]model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := &pb.GetNewsRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.client.GetNews(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	return convertPbMessagesToModel(resp.News), nil
}

// GetSingleMessage retrieves a single message by URL (replaces parser.GetSingleMessage)
func (c *LibrusScraperClient) GetSingleMessage(login, password, messageURL string) (model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := &pb.GetSingleMessageRequest{
		Login:      login,
		Password:   password,
		MessageUrl: messageURL,
	}

	resp, err := c.client.GetSingleMessage(ctx, req)
	if err != nil {
		return model.Message{}, fmt.Errorf("failed to get single message: %w", err)
	}

	if resp.Message == nil {
		return model.Message{}, fmt.Errorf("received nil message from server")
	}

	return convertPbMessageToModel(resp.Message), nil
}

// AnswerMessage sends a reply to a message (replaces parser.AnswerMessage)
func (c *LibrusScraperClient) AnswerMessage(login, password, messageURL, answerText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := &pb.AnswerMessageRequest{
		Login:      login,
		Password:   password,
		MessageUrl: messageURL,
		AnswerText: answerText,
	}

	_, err := c.client.AnswerMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to answer message: %w", err)
	}

	return nil
}

// GetAllUpdates retrieves both messages and news in one call
func (c *LibrusScraperClient) GetAllUpdates(login, password string) ([]model.Message, []model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	req := &pb.GetAllUpdatesRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.client.GetAllUpdates(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get all updates: %w", err)
	}

	messages := convertPbMessagesToModel(resp.Messages)
	news := convertPbMessagesToModel(resp.News)

	return messages, news, nil
}

// HealthCheck checks if the gRPC service is healthy
func (c *LibrusScraperClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.HealthCheckRequest{}

	resp, err := c.client.HealthCheck(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !resp.Healthy {
		return fmt.Errorf("service is not healthy: %s", resp.Status)
	}

	return nil
}
