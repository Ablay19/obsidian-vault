package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall" // Added for syscall.Signal
	"time"

	"obsidian-automation/internal/database/sqlc" // sqlc generated code
	"obsidian-automation/internal/telemetry"     // Use new structured logger

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	// _ "modernc.org/sqlite" // Commented out - using libsql driver instead
)

// DBClient combines the raw *sql.DB and the sqlc-generated Queries.
type DBClient struct {
	DB      *sql.DB
	Queries *sqlc.Queries
}

var Client *DBClient

// OpenDB initializes the database connection and sqlc queries.
func OpenDB() *DBClient {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url == "" {
		telemetry.ZapLogger.Sugar().Fatalw("TURSO_DATABASE_URL is missing", "url", url)
	}

	var db *sql.DB
	var err error

	// Handle file:// URLs for local SQLite
	if url == "file:./test.db" || (url == "file:./dev.db" || url == "file:./obsidian.db") {
		db, err = sql.Open("sqlite", url)
	} else {
		// Handle remote Turso URLs
		if token == "" {
			telemetry.ZapLogger.Sugar().Fatalw("TURSO_AUTH_TOKEN is missing for remote database", "url", url)
		}
		dsn := url + "?authToken=" + token
		db, err = sql.Open("libsql", dsn)
	}
	if err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Failed to open database", "error", err)
	}

	Client = &DBClient{
		DB:      db,
		Queries: sqlc.New(db),
	}
	return Client
}

// RunMigrations applies database migrations using a custom runner.
func RunMigrations(db *sql.DB) {
	telemetry.ZapLogger.Sugar().Info("Applying database migrations...")

	// Create schema_migrations table if it doesn't exist
	createMigrationsTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
`
	if _, err := db.Exec(createMigrationsTableSQL); err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Failed to create schema_migrations table", "error", err)
	}

	// Get all migration files from the migrations directory
	var migrationPaths []string
	files, err := os.ReadDir("./internal/database/migrations")
	if err != nil {
		telemetry.ZapLogger.Sugar().Warnw("Migration directory not found, skipping migrations", "error", err)
		telemetry.ZapLogger.Sugar().Info("Database migrations skipped (no migration files found)")
		return
	}

	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".sql") {
			migrationPaths = append(migrationPaths, filepath.Join("./internal/database/migrations", fileInfo.Name()))
		}
	}

	// Sort migration paths to ensure correct application order
	sort.Strings(migrationPaths)

	for _, path := range migrationPaths {
		migrationName := filepath.Base(path)

		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", migrationName).Scan(&count)
		if err != nil {
			telemetry.ZapLogger.Sugar().Fatalw("Failed to check migration status", "migration", migrationName, "error", err)
		}
		if count > 0 {
			telemetry.ZapLogger.Sugar().Infow("Migration already applied, skipping.", "migration", migrationName)
			continue
		}

		// Apply migration
		sqlContent, err := os.ReadFile(path)
		if err != nil {
			telemetry.ZapLogger.Sugar().Fatalw("Failed to read migration file", "migration", migrationName, "error", err)
		}

		// Execute each statement in the migration file
		for _, stmt := range strings.Split(string(sqlContent), ";") {
			trimmedStmt := strings.TrimSpace(stmt)
			if trimmedStmt == "" {
				continue
			}
			if _, err := db.Exec(trimmedStmt); err != nil {
				// Handle specific SQLite errors for "duplicate column" if it's an ALTER TABLE ADD COLUMN
				if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") && strings.Contains(strings.ToLower(trimmedStmt), "alter table") && strings.Contains(strings.ToLower(trimmedStmt), "add column") {
					telemetry.ZapLogger.Sugar().Warnw("Skipping ALTER TABLE ADD COLUMN due to duplicate column name (likely already applied)", "migration", migrationName, "statement", trimmedStmt, "error", err)
				} else {
					telemetry.ZapLogger.Sugar().Fatalw("Failed to execute SQL statement", "migration", migrationName, "statement", trimmedStmt, "error", err)
				}
			}
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migrationName); err != nil {
			telemetry.ZapLogger.Sugar().Fatalw("Failed to record migration", "migration", migrationName, "error", err)
		}
		telemetry.ZapLogger.Sugar().Infow("Migration applied successfully.", "migration", migrationName)
	}

	telemetry.ZapLogger.Sugar().Info("Database migrations applied successfully.")
}

const HEARTBEAT_THRESHOLD = 30 * time.Second

// CheckExistingInstance checks if another bot instance is running and cleans up stale instances.
func CheckExistingInstance(ctx context.Context) error {
	var pid int64
	var startedAt time.Time

	// Use a raw query to fetch the instance details
	row := Client.DB.QueryRowContext(ctx, "SELECT pid, started_at FROM instances WHERE id = 1")
	err := row.Scan(&pid, &startedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No instance found, safe to start
		}
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// If the heartbeat is older than the threshold, consider the instance stale.
	if time.Since(startedAt) > HEARTBEAT_THRESHOLD {
		telemetry.ZapLogger.Sugar().Warnw("Stale instance found, removing and starting new instance", "pid", pid, "last_heartbeat", startedAt)
		if err := Client.Queries.DeleteInstance(ctx); err != nil {
			telemetry.ZapLogger.Sugar().Errorw("Failed to delete stale instance", "pid", pid, "error", err)
			return fmt.Errorf("failed to delete stale instance with pid %d: %w", pid, err)
		}
		return nil // Stale instance removed, safe to start
	}

	// The instance is not stale, check if the process is running.
	// This is a secondary check; the primary one is the heartbeat.
	// In a container environment, this PID check can be misleading, but we keep it as a fallback.
	if isProcessRunning(int(pid)) {
		return fmt.Errorf("instance with PID %d already running and has a recent heartbeat", pid)
	}

	// If the process is not running but the heartbeat is recent, something is wrong.
	// This could happen if the bot crashed and the container is still up.
	// We'll treat it as a stale instance and remove it.
	telemetry.ZapLogger.Sugar().Warnw("Instance with recent heartbeat but no running process found, removing stale instance", "pid", pid, "last_heartbeat", startedAt)
	if err := Client.Queries.DeleteInstance(ctx); err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to delete stale instance", "pid", pid, "error", err)
		return fmt.Errorf("failed to delete stale instance with pid %d: %w", pid, err)
	}

	return nil
}

// AddInstance records the current bot instance's PID and start time.
func AddInstance(ctx context.Context, pid int) error {
	return Client.Queries.AddInstance(ctx, sqlc.AddInstanceParams{
		Pid:       int64(pid),
		StartedAt: time.Now(),
	})
}

// RemoveInstance deletes the current bot instance record.
func RemoveInstance(ctx context.Context) error {
	return Client.Queries.DeleteInstance(ctx)
}

// UpdateInstanceHeartbeat updates the timestamp of the running instance.
func UpdateInstanceHeartbeat(ctx context.Context) error {
	return Client.Queries.UpdateInstanceHeartbeat(ctx, time.Now())
}

// SaveMessage saves a chat message to the database.
func SaveMessage(ctx context.Context, userID, chatID int64, messageID int, direction, contentType, textContent, filePath string) error {
	return Client.Queries.SaveChatMessage(ctx, sqlc.SaveChatMessageParams{
		UserID:      userID,
		ChatID:      chatID,
		MessageID:   int64(messageID),
		Direction:   direction,
		ContentType: contentType,
		TextContent: sql.NullString{String: textContent, Valid: textContent != ""},
		FilePath:    sql.NullString{String: filePath, Valid: filePath != ""},
		CreatedAt:   time.Now(),
	})
}

// GetChatHistory retrieves chat messages for a user.
func GetChatHistory(ctx context.Context, userID int64, limit int) ([]sqlc.ChatHistory, error) {
	if userID == 0 { // Global history (e.g., for dashboard, or if a global user_id is intended)
		return Client.Queries.ListChatMessagesGlobal(ctx, int64(limit))
	}
	return Client.Queries.ListChatMessages(ctx, sqlc.ListChatMessagesParams{
		UserID: userID,
		Limit:  int64(limit),
	})
}

// UpsertUser inserts or updates user information in the database.
func UpsertUser(ctx context.Context, user *tgbotapi.User) error {
	return Client.Queries.UpsertUser(ctx, sqlc.UpsertUserParams{
		ID:           user.ID,
		Username:     sql.NullString{String: user.UserName, Valid: user.UserName != ""},
		FirstName:    sql.NullString{String: user.FirstName, Valid: user.FirstName != ""},
		LastName:     sql.NullString{String: user.LastName, Valid: user.LastName != ""},
		LanguageCode: sql.NullString{String: user.LanguageCode, Valid: user.LanguageCode != ""},
	})
}

// LinkTelegramToEmail associates a Telegram user ID with an existing email-based account.
func LinkTelegramToEmail(ctx context.Context, telegramID int64, email string) error {
	tx, err := Client.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := Client.Queries.WithTx(tx)

	// Update user by email to add telegram_id
	if err := qtx.LinkTelegramToEmailByEmail(ctx, sqlc.LinkTelegramToEmailByEmailParams{
		TelegramID: sql.NullInt64{Int64: telegramID, Valid: true},
		Email:      sql.NullString{String: email, Valid: email != ""}, // email also needs to be NullString
	}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to link telegram by email: %w", err)
	}

	// Also update the record where ID is telegramID if it doesn't have an email
	if err := qtx.LinkTelegramToEmailByID(ctx, sqlc.LinkTelegramToEmailByIDParams{
		Email: sql.NullString{String: email, Valid: email != ""},
		ID:    telegramID, // Assuming telegramID is used as the user's ID in the users table for telegram-originated users
	}); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to link telegram by ID: %w", err)
	}

	return tx.Commit()
}

// isProcessRunning checks if a process with the given PID is currently running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 to a process checks its existence without killing it.
	// It returns an error if the process does not exist.
	return process.Signal(syscall.Signal(0)) == nil // Changed to syscall.Signal
}
