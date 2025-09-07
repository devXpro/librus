package config

import (
	"os"
	"strings"
	"testing"
)

func TestGetAttachmentsBaseDir(t *testing.T) {
	// Test with environment variable
	os.Setenv("ATTACHMENTS_BASE_DIR", "/custom/path")
	defer os.Unsetenv("ATTACHMENTS_BASE_DIR")

	result := GetAttachmentsBaseDir()
	if result != "/custom/path" {
		t.Errorf("Expected /custom/path, got %s", result)
	}
}

func TestGetGRPCScraperAddress(t *testing.T) {
	// Test default value
	result := GetGRPCScraperAddress()
	if result != DefaultGRPCScraperAddress {
		t.Errorf("Expected %s, got %s", DefaultGRPCScraperAddress, result)
	}

	// Test with environment variable
	os.Setenv("GRPC_SCRAPER_ADDRESS", "custom:9999")
	defer os.Unsetenv("GRPC_SCRAPER_ADDRESS")

	result = GetGRPCScraperAddress()
	if result != "custom:9999" {
		t.Errorf("Expected custom:9999, got %s", result)
	}
}

func TestGetAttachmentsDirPath(t *testing.T) {
	// Test with empty UUID
	result := GetAttachmentsDirPath("")
	if result != "" {
		t.Errorf("Expected empty string for empty UUID, got %s", result)
	}

	// Test with valid UUID
	uuid := "test-uuid-123"
	result = GetAttachmentsDirPath(uuid)

	// Check that result contains the uuid
	if !strings.Contains(result, uuid) {
		t.Errorf("Expected result to contain UUID %s, got %s", uuid, result)
	}

	// Check that result ends with the uuid
	if !strings.HasSuffix(result, uuid) {
		t.Errorf("Expected result to end with UUID %s, got %s", uuid, result)
	}
}
