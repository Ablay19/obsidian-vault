package database

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
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

// ApplySchemaAndMigrations reads SQL files and executes their contents against the database.
func ApplySchemaAndMigrations(db *sql.DB) {
	log.Println("Applying database schema and migrations...")

	// Define paths relative to the project root
	schemaPath := "./internal/database/schema.sql"
	migration1Path := "./internal/database/migrations/001_create_chat_history.sql"
	migration2Path := "./internal/database/migrations/002_add_ai_fields_to_processed_files.sql"
	migration3Path := "./internal/database/migrations/003_create_users_table.sql"

	// Read and execute schema.sql
	schemaSQL, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}
	executeSQL(db, string(schemaSQL))

	// Read and execute migrations
	migration1SQL, err := ioutil.ReadFile(migration1Path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", migration1Path, err)
	}
	executeSQL(db, string(migration1SQL))

	migration2SQL, err := ioutil.ReadFile(migration2Path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", migration2Path, err)
	}
	executeSQL(db, string(migration2SQL))

	migration3SQL, err := ioutil.ReadFile(migration3Path)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", migration3Path, err)
	}
	executeSQL(db, string(migration3SQL))

	log.Println("Database schema and migrations applied successfully.")
}

func executeSQL(db *sql.DB, sqlContent string) {
	// Split SQL content by semicolons, ignoring empty statements
	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue
		}
		_, err := db.Exec(trimmedStmt)
		if err != nil {
			log.Fatalf("Failed to execute SQL statement: %s\nError: %v", trimmedStmt, err)
		}
	}
}

func UpsertUser(user *tgbotapi.User) error {
	_, err := DB.Exec(`
		INSERT INTO users (id, username, first_name, last_name, language_code)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			language_code = excluded.language_code
	`, user.ID, user.UserName, user.FirstName, user.LastName, user.LanguageCode)
	return err
}
