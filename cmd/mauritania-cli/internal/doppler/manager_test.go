package doppler

import (
	"context"
	"os"
	"testing"
)

func TestManager_LoadSecret_WithFallbacks(t *testing.T) {
	// Clear any existing test environment
	os.Unsetenv("TEST_KEY")

	manager := NewManager("test-project", "test-config")
	manager.WithFallbacks(map[string]string{
		"TEST_KEY": "fallback_value",
	})

	ctx := context.Background()

	// Test fallback loading
	value, err := manager.LoadSecret(ctx, "TEST_KEY")
	if err != nil {
		t.Fatalf("Expected no error loading fallback, got: %v", err)
	}

	if value != "fallback_value" {
		t.Errorf("Expected fallback_value, got: %s", value)
	}
}

func TestManager_LoadSecret_FromEnvironment(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_ENV_KEY", "env_value")
	defer os.Unsetenv("TEST_ENV_KEY")

	manager := NewManager("test-project", "test-config")
	manager.WithFallbacks(map[string]string{
		"TEST_ENV_KEY": "fallback_value",
	})

	ctx := context.Background()

	// Test environment variable loading
	value, err := manager.LoadSecret(ctx, "TEST_ENV_KEY")
	if err != nil {
		t.Fatalf("Expected no error loading from environment, got: %v", err)
	}

	if value != "env_value" {
		t.Errorf("Expected env_value, got: %s", value)
	}
}

func TestManager_LoadAllSecrets(t *testing.T) {
	// Set up test environment
	os.Setenv("TEST_MULTI_1", "value1")
	os.Setenv("TEST_MULTI_2", "value2")
	defer func() {
		os.Unsetenv("TEST_MULTI_1")
		os.Unsetenv("TEST_MULTI_2")
	}()

	manager := NewManager("test-project", "test-config")
	manager.WithFallbacks(map[string]string{
		"TEST_FALLBACK": "fallback_value",
	})

	ctx := context.Background()

	secrets, err := manager.LoadAllSecrets(ctx)
	if err != nil {
		t.Fatalf("Expected no error loading all secrets, got: %v", err)
	}

	if len(secrets) == 0 {
		t.Error("Expected at least some secrets to be loaded")
	}

	if secrets["TEST_MULTI_1"] != "value1" {
		t.Errorf("Expected TEST_MULTI_1=value1, got: %s", secrets["TEST_MULTI_1"])
	}

	if secrets["TEST_FALLBACK"] != "fallback_value" {
		t.Errorf("Expected TEST_FALLBACK=fallback_value, got: %s", secrets["TEST_FALLBACK"])
	}
}

func TestManager_IsAvailable(t *testing.T) {
	manager := NewManager("test-project", "test-config")

	// This should work with our mock/fallback client
	available := manager.IsAvailable(context.Background())
	// We expect this to be false since we don't have real Doppler setup
	if available {
		t.Log("Doppler is unexpectedly available in test environment")
	}
}
