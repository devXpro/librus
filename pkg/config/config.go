package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// Default paths for attachments
	DefaultAttachmentsBaseDir = "/app/attachments"
	LocalAttachmentsBaseDir   = "./playwright/attachments"

	// gRPC configuration
	DefaultGRPCScraperAddress = "localhost:50051"

	// Message processing configuration
	DefaultMessageCheckInterval = 30 * time.Minute
)

// GetAttachmentsBaseDir returns the base directory for attachments
// Uses environment variable or defaults based on environment
func GetAttachmentsBaseDir() string {
	if dir := os.Getenv("ATTACHMENTS_BASE_DIR"); dir != "" {
		return dir
	}

	// Check if we're running in Docker (common indicator)
	if _, err := os.Stat("/app"); err == nil {
		return DefaultAttachmentsBaseDir
	}

	// Local development
	return LocalAttachmentsBaseDir
}

// GetMessageCheckInterval returns the message check interval
// Uses environment variable or default value
func GetMessageCheckInterval() time.Duration {
	intervalStr := os.Getenv("MESSAGE_CHECK_INTERVAL")
	if intervalStr == "" {
		return DefaultMessageCheckInterval
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		// If parsing fails, return default
		return DefaultMessageCheckInterval
	}

	return interval
}

// GetGRPCScraperAddress returns the gRPC scraper service address
func GetGRPCScraperAddress() string {
	if addr := os.Getenv("GRPC_SCRAPER_ADDRESS"); addr != "" {
		return addr
	}
	return DefaultGRPCScraperAddress
}

// GetAttachmentsDirPath returns the full path to an attachments directory by UUID
func GetAttachmentsDirPath(uuid string) string {
	if uuid == "" {
		return ""
	}
	return filepath.Join(GetAttachmentsBaseDir(), uuid)
}
