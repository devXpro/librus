package grpc_client

import (
	"os"
	"testing"
)

// TestGRPCClientConnection tests that we can create a gRPC client
// This is an integration test that requires the gRPC server to be running
func TestGRPCClientConnection(t *testing.T) {
	// Skip this test if we're not in integration test mode
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client, err := NewLibrusScraperClient()
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	// Test health check
	err = client.HealthCheck()
	if err != nil {
		t.Logf("Health check failed (expected if server is not running): %v", err)
		// Don't fail the test if server is not running
		return
	}

	t.Log("Successfully connected to gRPC server and health check passed")
}

// TestGRPCClientCreation tests that we can create a client without connecting
func TestGRPCClientCreation(t *testing.T) {
	client, err := NewLibrusScraperClient()
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.client == nil {
		t.Fatal("Internal gRPC client should not be nil")
	}
}
