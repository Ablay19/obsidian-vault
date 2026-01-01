package database

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"strings"

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
	migrationPath := "./internal/database/migrations/001_create_chat_history.sql"

	// Read and execute schema.sql
	schemaSQL, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}
	executeSQL(db, string(schemaSQL))

	// Read and execute 001_create_chat_history.sql
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read 001_create_chat_history.sql: %v", err)
	}
	executeSQL(db, string(migrationSQL))

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
