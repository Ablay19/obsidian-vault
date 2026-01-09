#!/bin/bash

echo "🔧 Running database migrations..."
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
    echo "✅ Environment variables loaded from .env"
else
    echo "❌ .env file not found"
    exit 1
fi

# Create a simple migration runner
cat > run_migrations.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "sort"
    "strings"
    "path/filepath"

    _ "github.com/tursodatabase/libsql-client-go/libsql"
    _ "modernc.org/sqlite"
)

func main() {
    url := os.Getenv("TURSO_DATABASE_URL")
    token := os.Getenv("TURSO_AUTH_TOKEN")

    if url == "" {
        log.Fatal("TURSO_DATABASE_URL is missing")
    }

    var db *sql.DB
    var err error

    // Handle file:// URLs for local SQLite
    if url == "file:./test.db" || url == "file:./dev.db" || url == "file:./obsidian.db" {
        db, err = sql.Open("sqlite", url)
    } else {
        // Handle remote Turso URLs
        if token == "" {
            log.Fatal("TURSO_AUTH_TOKEN is missing for remote database")
        }
        dsn := url + "?authToken=" + token
        db, err = sql.Open("libsql", dsn)
    }
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    // Test connection
    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }
    fmt.Println("✅ Database connection established")

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
    fmt.Println("✅ Schema migrations table ready")

    // Get all migration files
    migrationDir := "./internal/database/migrations"
    files, err := os.ReadDir(migrationDir)
    if err != nil {
        log.Fatalf("Failed to read migration directory: %v", err)
    }

    var migrationPaths []string
    for _, fileInfo := range files {
        if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".sql") && strings.Contains(fileInfo.Name(), ".up.sql") {
            migrationPaths = append(migrationPaths, filepath.Join(migrationDir, fileInfo.Name()))
        }
    }

    sort.Strings(migrationPaths)
    fmt.Printf("📝 Found %d migration files\n", len(migrationPaths))

    // Run migrations
    for _, path := range migrationPaths {
        migrationName := filepath.Base(path)

        // Check if migration already applied
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", migrationName).Scan(&count)
        if err != nil {
            log.Fatalf("Failed to check migration status: %v", err)
        }
        if count > 0 {
            fmt.Printf("⏭️  Migration already applied: %s\n", migrationName)
            continue
        }

        fmt.Printf("🔄 Applying migration: %s\n", migrationName)

        // Read and apply migration
        sqlContent, err := os.ReadFile(path)
        if err != nil {
            log.Fatalf("Failed to read migration file %s: %v", migrationName, err)
        }

        // Execute each statement
        for _, stmt := range strings.Split(string(sqlContent), ";") {
            trimmedStmt := strings.TrimSpace(stmt)
            if trimmedStmt == "" {
                continue
            }
            if _, err := db.Exec(trimmedStmt); err != nil {
                log.Printf("⚠️  Warning executing statement in %s: %v\nStatement: %s", migrationName, err, trimmedStmt)
            }
        }

        // Record migration as applied
        if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", migrationName); err != nil {
            log.Fatalf("Failed to record migration %s: %v", migrationName, err)
        }

        fmt.Printf("✅ Migration applied: %s\n", migrationName)
    }

    fmt.Println("🎉 All migrations completed successfully!")
}
EOF

echo "📦 Building migration runner..."
go build -o migrate run_migrations.go

echo "🚀 Running migrations..."
./migrate

echo "🧹 Cleaning up..."
rm -f run_migrations.go migrate

echo "✅ Migration process completed!"
