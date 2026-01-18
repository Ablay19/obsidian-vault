package e2e

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/doppler"
)

// TestEnvironment represents a complete test environment setup
type TestEnvironment struct {
	DopplerManager *doppler.Manager
	OriginalEnv    map[string]string
	CleanupFuncs   []func()
}

// SetupTestEnvironment initializes the test environment with Doppler
func SetupTestEnvironment(t *testing.T, project, config string) *TestEnvironment {
	t.Helper()

	env := &TestEnvironment{
		DopplerManager: doppler.NewManager(project, config),
		OriginalEnv:    make(map[string]string),
		CleanupFuncs:   []func(){},
	}

	// Save original environment
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			env.OriginalEnv[parts[0]] = parts[1]
		}
	}

	// Set up Doppler fallbacks for testing
	env.DopplerManager.WithFallbacks(map[string]string{
		"TEST_DATABASE_URL":  "sqlite://:memory:",
		"TEST_REDIS_ADDR":    "localhost:6379",
		"TEST_TIMEOUT":       "30",
		"TELEGRAM_BOT_TOKEN": "test_token",
		"WHATSAPP_API_KEY":   "test_key",
	})

	// Register cleanup
	t.Cleanup(func() {
		env.Cleanup(t)
	})

	return env
}

// LoadTestSecrets loads test secrets from Doppler or fallbacks
func (env *TestEnvironment) LoadTestSecrets(ctx context.Context, t *testing.T) map[string]string {
	t.Helper()

	secrets, err := env.DopplerManager.LoadAllSecrets(ctx)
	if err != nil {
		t.Logf("Doppler not available, using fallbacks: %v", err)
		// Fallbacks are already set up in the manager
		secrets, err = env.DopplerManager.LoadAllSecrets(ctx)
		if err != nil {
			t.Fatalf("Failed to load test secrets: %v", err)
		}
	}

	return secrets
}

// SetTestEnvironment applies test environment variables
func (env *TestEnvironment) SetTestEnvironment(ctx context.Context, t *testing.T) {
	t.Helper()

	if err := env.DopplerManager.SetEnvironment(ctx); err != nil {
		t.Logf("Doppler environment setup failed, using fallbacks: %v", err)
		// Apply fallbacks manually
		fallbacks := map[string]string{
			"TEST_DATABASE_URL":  "sqlite://:memory:",
			"TEST_REDIS_ADDR":    "localhost:6379",
			"TEST_TIMEOUT":       "30",
			"TELEGRAM_BOT_TOKEN": "test_token",
			"WHATSAPP_API_KEY":   "test_key",
		}

		for k, v := range fallbacks {
			os.Setenv(k, v)
		}
	}
}

// AddCleanup adds a cleanup function to be called at test end
func (env *TestEnvironment) AddCleanup(cleanup func()) {
	env.CleanupFuncs = append(env.CleanupFuncs, cleanup)
}

// Cleanup restores the original environment and runs cleanup functions
func (env *TestEnvironment) Cleanup(t *testing.T) {
	// Run cleanup functions
	for _, cleanup := range env.CleanupFuncs {
		cleanup()
	}

	// Clear test environment variables
	testPrefixes := []string{"TEST_", "DOPPLER_", "TELEGRAM_", "WHATSAPP_", "FACEBOOK_"}
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			shouldRemove := false

			for _, prefix := range testPrefixes {
				if strings.HasPrefix(key, prefix) {
					shouldRemove = true
					break
				}
			}

			if shouldRemove {
				os.Unsetenv(key)
			}
		}
	}

	// Restore original environment
	for key, value := range env.OriginalEnv {
		os.Setenv(key, value)
	}

	// Clear Doppler cache
	env.DopplerManager.ClearCache()
}

// WaitForService waits for a service to be available
func WaitForService(ctx context.Context, check func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := check(); err == nil {
				return nil
			}
		}
	}
}

// MockTransportServer creates a mock transport server for testing
func MockTransportServer(t *testing.T, transportType string) (string, func()) {
	t.Helper()

	// This would create mock servers for different transports
	// For now, return a placeholder
	switch transportType {
	case "whatsapp":
		return "http://localhost:3001", func() {}
	case "telegram":
		return "http://localhost:3002", func() {}
	case "facebook":
		return "http://localhost:3003", func() {}
	default:
		t.Fatalf("Unsupported transport type: %s", transportType)
		return "", nil
	}
}
