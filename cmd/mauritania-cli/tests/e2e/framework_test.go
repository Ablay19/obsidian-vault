package e2e

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSetupTestEnvironment(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, "test-project", "test-config")

	if env == nil {
		t.Fatal("Expected test environment to be created")
	}

	if env.DopplerManager == nil {
		t.Fatal("Expected Doppler manager to be initialized")
	}

	// Test that fallbacks are set
	ctx := context.Background()
	secrets := env.LoadTestSecrets(ctx, t)

	if len(secrets) == 0 {
		t.Error("Expected some secrets to be loaded")
	}

	// Check that test environment variables are applied
	env.SetTestEnvironment(ctx, t)

	if os.Getenv("TEST_DATABASE_URL") != "sqlite://:memory:" {
		t.Errorf("Expected TEST_DATABASE_URL to be set, got: %s", os.Getenv("TEST_DATABASE_URL"))
	}
}

func TestWaitForService_Timeout(t *testing.T) {
	ctx := context.Background()
	check := func() error {
		return nil // Service is immediately available
	}

	err := WaitForService(ctx, check, 1*time.Second)
	if err != nil {
		t.Errorf("Expected no error for immediately available service, got: %v", err)
	}
}

func TestMockTransportServer(t *testing.T) {
	whatsappURL, cleanup := MockTransportServer(t, "whatsapp")
	if whatsappURL != "http://localhost:3001" {
		t.Errorf("Expected whatsapp URL localhost:3001, got: %s", whatsappURL)
	}
	if cleanup == nil {
		t.Error("Expected cleanup function to be returned")
	}

	telegramURL, cleanup := MockTransportServer(t, "telegram")
	if telegramURL != "http://localhost:3002" {
		t.Errorf("Expected telegram URL localhost:3002, got: %s", telegramURL)
	}
	if cleanup == nil {
		t.Error("Expected cleanup function to be returned")
	}
}
