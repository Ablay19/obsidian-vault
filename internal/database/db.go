package database

import (
	"database/sql"
	"log"
	"os"

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
