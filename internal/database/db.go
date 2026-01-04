package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"go.uber.org/zap"
)

var DB *sql.DB

func OpenDB() *sql.DB {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url == "" || token == "" {
		fmt.Fprintf(os.Stderr, "FATAL: TURSO_DATABASE_URL or TURSO_AUTH_TOKEN is missing. Please check your environment or .env file.\n")
		zap.S().Error("TURSO_DATABASE_URL or TURSO_AUTH_TOKEN is missing")
		os.Exit(1)
	}

	dsn := url + "?authToken=" + token

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		zap.S().Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	DB = db
	return db
}

// RunMigrations applies database migrations using a custom runner.
func RunMigrations(db *sql.DB) {
	zap.S().Info("Applying database migrations...")

	// Create schema_migrations table if it doesn't exist
	createMigrationsTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
`
	if _, err := db.Exec(createMigrationsTableSQL); err != nil {
		zap.S().Error("Failed to create schema_migrations table", "error", err)
		os.Exit(1)
	}

	// Get all migration files from the migrations directory
	var migrationPaths []string
	files, err := os.ReadDir("./internal/database/migrations")
	if err != nil {
		zap.S().Error("Failed to read migration directory", "error", err)
		os.Exit(1)
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
			zap.S().Error("Failed to check migration status", "migration", migrationName, "error", err)
			os.Exit(1)
		}
		if count > 0 {
			zap.S().Info("Migration already applied, skipping.", "migration", migrationName)
			continue
		}

		// Apply migration
		sqlContent, err := os.ReadFile(path)
		if err != nil {
			zap.S().Error("Failed to read migration file", "migration", migrationName, "error", err)
			os.Exit(1)
		}

		if err := executeSQL(db, string(sqlContent)); err != nil {
			zap.S().Error("Failed to apply migration", "migration", migrationName, "error", err)
			os.Exit(1)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migrationName); err != nil {
			zap.S().Error("Failed to record migration", "migration", migrationName, "error", err)
			os.Exit(1)
		}
		zap.S().Info("Migration applied successfully.", "migration", migrationName)
	}

	zap.S().Info("Database migrations applied successfully.")
}

// columnExists checks if a column exists in a given table.
func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, fmt.Errorf("failed to query table info for %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid      int
			name     string
			ctype    string
			notnull  int
			dflt_val sql.NullString
			pk       int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_val, &pk); err != nil {
			return false, fmt.Errorf("failed to scan table info row: %w", err)
		}
		if strings.EqualFold(name, columnName) {
			return true, nil
		}
	}
	return false, nil
}

// executeSQL executes SQL content, handling ALTER TABLE ADD COLUMN idempotently.
func executeSQL(db *sql.DB, sqlContent string) error {
	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue
		}

		// Remove comments for regex matching
		lines := strings.Split(trimmedStmt, "\n")
		var cleanStmtLines []string
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "--") {
				cleanStmtLines = append(cleanStmtLines, trimmedLine)
			}
		}
		cleanStmt := strings.Join(cleanStmtLines, " ")

		// Simplified regex to just get the table and column name from ADD [COLUMN]
		re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(\S+)\s+ADD\s+(?:COLUMN\s+)?(\S+)`)
		matches := re.FindStringSubmatch(cleanStmt)
		if len(matches) == 3 {
			tableName := matches[1]
			columnName := matches[2]
			exists, err := columnExists(db, tableName, columnName)
			if err != nil {
				return fmt.Errorf("failed to check column existence for %s.%s: %w", tableName, columnName, err)
			}
			if exists {
				continue
			}
		}

		if _, err := db.Exec(trimmedStmt); err != nil {
			return fmt.Errorf("failed to execute SQL statement '%s': %w", trimmedStmt, err)
		}
	}
	return nil
}

// UpsertUser inserts or updates user information in the database.
func UpsertUser(user *tgbotapi.User) error {
	_, err := DB.Exec(`
		INSERT INTO users (id, username, first_name, last_name, language_code)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			language_code = excluded.language_code
	`,
		user.ID, user.UserName, user.FirstName, user.LastName, user.LanguageCode)
	return err
}

// LinkTelegramToEmail associates a Telegram user ID with an existing email-based account.
func LinkTelegramToEmail(telegramID int64, email string) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET telegram_id = ? 
		WHERE email = ?
	`, telegramID, email)
	if err != nil {
		return err
	}
	
	// Also update the record where ID is telegramID if it doesn't have an email
	_, err = DB.Exec(`
		UPDATE users 
		SET email = ? 
		WHERE id = ? AND email IS NULL
	`, email, telegramID)
	
	return err
}