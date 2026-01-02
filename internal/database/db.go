package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *sql.DB

func OpenDB() *sql.DB {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url == "" || token == "" {
		log.Fatal("TURSO_DATABASE_URL or TURSO_AUTH_TOKEN is missing")
	}

	dsn := url + "?authToken=" + token

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	return db
}

// RunMigrations applies database migrations using a custom runner.
func RunMigrations(db *sql.DB) {
	log.Println("Applying database migrations...")

	// Create schema_migrations table if it doesn't exist
	createMigrationsTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createMigrationsTableSQL); err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Get all migration files from the migrations directory
	var migrationPaths []string
	files, err := ioutil.ReadDir("./internal/database/migrations")
	if err != nil {
		log.Fatalf("Failed to read migration directory: %v", err)
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
			log.Fatalf("Failed to check migration status for %s: %v", migrationName, err)
		}
		if count > 0 {
			log.Printf("Migration %s already applied, skipping.", migrationName)
			continue
		}

		// Apply migration
		sqlContent, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", migrationName, err)
		}

		if err := executeSQL(db, string(sqlContent)); err != nil {
			log.Fatalf("Failed to apply migration %s: %v", migrationName, err)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migrationName); err != nil {
			log.Fatalf("Failed to record migration %s: %v", migrationName, err)
		}
		log.Printf("Migration %s applied successfully.", migrationName)
	}

	log.Println("Database migrations applied successfully.")
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
		if name == columnName {
			return true, nil
		}
	}
	return false, nil
}

// executeSQL executes SQL content, handling ALTER TABLE ADD COLUMN idempotently.
func executeSQL(db *sql.DB, sqlContent string) error {
	alterColumnRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(\S+)\s+ADD\s+COLUMN\s+(\S+)\s+.*`)

	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue
		}

		matches := alterColumnRegex.FindStringSubmatch(trimmedStmt)
		if len(matches) == 3 {
			tableName := matches[1]
			columnName := matches[2]
			exists, err := columnExists(db, tableName, columnName)
			if err != nil {
				return fmt.Errorf("failed to check column existence for %s.%s: %w", tableName, columnName, err)
			}
			if exists {
				log.Printf("Column '%s' already exists in table '%s', skipping ALTER TABLE ADD COLUMN statement.", columnName, tableName)
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
