package ai

import (
	"context"
	"database/sql"
	"testing"

	st "obsidian-automation/internal/state"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE runtime_config (
			id INTEGER PRIMARY KEY,
			config_data BLOB,
			updated_at DATETIME
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create runtime_config table: %v", err)
	}

	return db
}

func TestAIIntegration_RefreshProviders(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rcm, err := st.NewRuntimeConfigManager(db)
	if err != nil {
		t.Fatalf("Failed to create RCM: %v", err)
	}

	// Clear any Env-loaded keys for clean test
	rcm.ResetState()

	ctx := context.Background()
	aiService := NewAIService(ctx, rcm)

	// Initially, no providers should be available (since no keys)
	available := aiService.GetAvailableProviders()
	if len(available) != 0 {
		t.Errorf("Expected 0 available providers, got %v", available)
	}

	// Add an API key for Gemini
	_, err = rcm.AddAPIKey("Gemini", "test-api-key", true)
	if err != nil {
		t.Fatalf("Failed to add API key: %v", err)
	}

	// Ensure the provider is enabled (defaults to disabled if no env keys found during init)
	err = rcm.SetProviderState("Gemini", true, false, false, "")
	if err != nil {
		t.Fatalf("Failed to enable Gemini provider: %v", err)
	}

	// Refresh AIService
	aiService.RefreshProviders(ctx)

	// Now Gemini should be available
	available = aiService.GetAvailableProviders()
	found := false
	for _, p := range available {
		if p == "Gemini" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Gemini to be available, got %v", available)
	}

	// Disable Gemini provider state
	err = rcm.SetProviderState("Gemini", false, false, false, "")
	if err != nil {
		t.Fatalf("Failed to disable Gemini: %v", err)
	}

	aiService.RefreshProviders(ctx)

	// Gemini should no longer be available
	available = aiService.GetAvailableProviders()
	for _, p := range available {
		if p == "Gemini" {
			t.Errorf("Gemini should not be available after disabling")
		}
	}
}
