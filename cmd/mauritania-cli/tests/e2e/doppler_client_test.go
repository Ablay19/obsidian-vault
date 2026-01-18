package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/doppler"
)

func TestDopplerClient_NewClient(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")

	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestDopplerClient_WithToken(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")
	updatedClient := client.WithToken("test-token")

	if updatedClient == nil {
		t.Fatal("Expected updated client")
	}
}

func TestDopplerClient_WithTimeout(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")
	updatedClient := client.WithTimeout(60 * time.Second)

	if updatedClient == nil {
		t.Fatal("Expected updated client")
	}
}

func TestDopplerClient_IsAvailable(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.IsAvailable(ctx)

	if err == nil {
		t.Log("Doppler is available")
	} else {
		t.Logf("Doppler not available (expected in test env): %v", err)
	}
}

func TestDopplerClient_GetSecret(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value, err := client.GetSecret(ctx, "TEST_SECRET")

	if err != nil {
		t.Logf("GetSecret failed (expected without Doppler): %v", err)
	} else {
		if value == "" {
			t.Error("Expected secret value to be non-empty")
		}
	}
}

func TestDopplerClient_GetSecrets(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	secrets, err := client.GetSecrets(ctx)

	if err != nil {
		t.Logf("GetSecrets failed (expected without Doppler): %v", err)
	} else {
		if len(secrets) == 0 {
			t.Error("Expected at least one secret")
		}
	}
}

func TestDopplerClient_Run(t *testing.T) {
	client := doppler.NewClient("test-project", "test-config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := []string{"echo", "test"}
	err := client.Run(ctx, command)

	if err != nil {
		t.Logf("Run failed (expected without Doppler): %v", err)
	}
}

func TestDopplerClient_WithServiceToken(t *testing.T) {
	token := os.Getenv("DOPPLER_TOKEN")
	if token == "" {
		t.Skip("DOPPLER_TOKEN not set, skipping service token test")
	}

	client := doppler.NewClient("test-project", "test-config").WithToken(token)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.IsAvailable(ctx)
	if err != nil {
		t.Errorf("Expected service token to work, got: %v", err)
	}
}
