package config

import (
	"os"
	"path/filepath"
)

const (
	// Default paths for attachments
	DefaultAttachmentsBaseDir = "/app/attachments"
	LocalAttachmentsBaseDir   = "./playwright/attachments"

	// gRPC configuration
	DefaultGRPCScraperAddress = "localhost:50051"
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
