package storage

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/glebarez/go-sqlite"
)

// Database holds the database connection and methods
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databasePath string) (*Database, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign key support
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// RunMigrations runs all database migrations
func (d *Database) RunMigrations() error {
	slog.Info("Running database migrations")

	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id BIGINT UNIQUE NOT NULL,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			language TEXT DEFAULT 'en',
			personality TEXT DEFAULT 'helpful',
			preferences TEXT, -- JSON string for preferences
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Conversations table
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id BIGINT NOT NULL,
			title TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (telegram_id) ON DELETE CASCADE
		)`,
		// Messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			user_id BIGINT NOT NULL,
			content TEXT NOT NULL,
			message_type TEXT DEFAULT 'user', -- 'user', 'bot', 'system'
			model_used TEXT,
			tokens_used INTEGER DEFAULT 0,
			processing_time INTEGER DEFAULT 0, -- milliseconds
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users (telegram_id) ON DELETE CASCADE
		)`,
		// AI Providers table
		`CREATE TABLE IF NOT EXISTS ai_providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			model TEXT NOT NULL,
			provider_type TEXT NOT NULL, -- 'local', 'api'
			config TEXT, -- JSON string for provider-specific config
			enabled BOOLEAN DEFAULT TRUE,
			rate_limit INTEGER DEFAULT 100,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Rate limiting table
		`CREATE TABLE IF NOT EXISTS rate_limits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id BIGINT NOT NULL,
			limit_type TEXT NOT NULL, -- 'hour', 'day'
			request_count INTEGER DEFAULT 0,
			reset_time DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, limit_type)
		)`,
	}

	// Indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id)",
		"CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_rate_limits_user_type ON rate_limits(user_id, limit_type)",
	}

	// Run migrations
	for i, migration := range migrations {
		slog.Debug("Running migration", "index", i+1)
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Run index creation
	for _, index := range indexes {
		if _, err := d.db.Exec(index); err != nil {
			slog.Warn("Failed to create index", "error", err)
		}
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// GetStats returns database statistics
func (d *Database) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count users
	var userCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		slog.Error("Failed to count users", "error", err)
	}
	stats["users"] = userCount

	// Count conversations
	var conversationCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&conversationCount); err != nil {
		slog.Error("Failed to count conversations", "error", err)
	}
	stats["conversations"] = conversationCount

	// Count messages
	var messageCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&messageCount); err != nil {
		slog.Error("Failed to count messages", "error", err)
	}
	stats["messages"] = messageCount

	return stats
}
