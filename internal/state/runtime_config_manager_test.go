package state

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import for side-effects to use in-memory SQLite
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create the runtime_config table
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

func TestNewRuntimeConfigManager(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager, err := NewRuntimeConfigManager(db)
	if err != nil {
		t.Fatalf("NewRuntimeConfigManager() error = %v", err)
	}
	if manager == nil {
		t.Fatal("NewRuntimeConfigManager() returned nil manager")
	}

	// Check if default AIEnabled is true
	config := manager.GetConfig()
	if !config.AIEnabled {
		t.Error("Expected AIEnabled to be true by default, got false")
	}
}

func TestSetAIEnabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager, _ := NewRuntimeConfigManager(db)

	// Disable AI
	err := manager.SetAIEnabled(false)
	if err != nil {
		t.Fatalf("SetAIEnabled() error = %v", err)
	}

	config := manager.GetConfig()
	if config.AIEnabled {
		t.Error("Expected AIEnabled to be false, got true")
	}

	// Re-enable AI
	err = manager.SetAIEnabled(true)
	if err != nil {
		t.Fatalf("SetAIEnabled() error = %v", err)
	}

	config = manager.GetConfig()
	if !config.AIEnabled {
		t.Error("Expected AIEnabled to be true, got false")
	}
}
