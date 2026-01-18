package e2e

import (
	"context"
	"os"
	"testing"

	"obsidian-automation/cmd/mauritania-cli/internal/doppler"
)

func TestEnvManager_NewManager(t *testing.T) {
	manager := SetupTestEnvironment(t, "test-project", "test-config")

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.DopplerManager == nil {
		t.Error("Expected DopplerManager to be initialized")
	}
}

func TestEnvManager_WithFallbacks(t *testing.T) {
	manager := doppler.NewManager("test-project", "test-config")

	fallbacks := map[string]string{
		"TEST_KEY": "test_value",
	}

	updatedEnv := manager.WithFallbacks(fallbacks)

	if updatedEnv == nil {
		t.Fatal("Expected updated manager")
	}
}

func TestEnvManager_LoadSecret(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	manager.WithFallbacks(map[string]string{
		"TEST_SECRET": "fallback_value",
	})

	value, err := manager.LoadSecret(ctx, "TEST_SECRET")

	if err != nil {
		t.Errorf("Expected to load secret with fallback, got error: %v", err)
	}

	if value != "fallback_value" {
		t.Errorf("Expected fallback_value, got: %s", value)
	}
}

func TestEnvManager_LoadSecret_Cache(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	manager.WithFallbacks(map[string]string{
		"CACHE_TEST": "cached_value",
	})

	value1, err1 := manager.LoadSecret(ctx, "CACHE_TEST")
	if err1 != nil {
		t.Fatalf("First load failed: %v", err1)
	}

	value2, err2 := manager.LoadSecret(ctx, "CACHE_TEST")
	if err2 != nil {
		t.Fatalf("Second load failed: %v", err2)
	}

	if value1 != value2 {
		t.Error("Expected cached values to be identical")
	}

	if value1 != "cached_value" {
		t.Errorf("Expected cached_value, got: %s", value1)
	}
}

func TestEnvManager_LoadAllSecrets(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	secrets, err := manager.LoadAllSecrets(ctx)

	if err != nil {
		t.Logf("LoadAllSecrets failed (expected without Doppler): %v", err)
	}

	if len(secrets) == 0 {
		t.Error("Expected at least one secret from environment or fallbacks")
	}
}

func TestEnvManager_SetEnvironment(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	manager.WithFallbacks(map[string]string{
		"ENV_TEST_KEY": "env_test_value",
	})

	os.Unsetenv("ENV_TEST_KEY")

	err := manager.SetEnvironment(ctx)
	if err != nil {
		t.Errorf("SetEnvironment failed: %v", err)
	}

	value := os.Getenv("ENV_TEST_KEY")
	if value != "env_test_value" {
		t.Errorf("Expected env_test_value, got: %s", value)
	}

	os.Unsetenv("ENV_TEST_KEY")
}

func TestEnvManager_ClearCache(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	manager.WithFallbacks(map[string]string{
		"CLEAR_TEST": "test_value",
	})

	value1, _ := manager.LoadSecret(ctx, "CLEAR_TEST")

	manager.ClearCache()

	value2, _ := manager.LoadSecret(ctx, "CLEAR_TEST")

	if value1 != value2 {
		t.Error("Expected values to be same after cache clear")
	}
}

func TestEnvManager_IsAvailable(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	available := manager.IsAvailable(ctx)

	if available {
		t.Log("Doppler is available")
	} else {
		t.Log("Doppler is not available (expected in test env)")
	}
}

func TestEnvManager_LoadSecret_Priority(t *testing.T) {
	ctx := context.Background()
	manager := doppler.NewManager("test-project", "test-config")

	os.Setenv("PRIORITY_TEST", "from_env")
	defer os.Unsetenv("PRIORITY_TEST")

	manager.WithFallbacks(map[string]string{
		"PRIORITY_TEST": "from_fallback",
	})

	value, err := manager.LoadSecret(ctx, "PRIORITY_TEST")

	if err != nil {
		t.Errorf("Failed to load secret: %v", err)
	}

	if value != "from_env" {
		t.Errorf("Expected from_env (higher priority), got: %s", value)
	}
}
