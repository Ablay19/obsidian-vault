package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"os"
)

var processedHashes = make(map[string]string)

func SaveProcessed(db *sql.DB, hash, category, text string) error {
	_, err := db.Exec(
		`INSERT INTO processed_files (hash, category, extracted_text)
		 VALUES (?, ?, ?)`,
		hash, category, text,
	)
	return err
}
func getFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), nil
}
func IsProcessed(db *sql.DB, hash string) (bool, error) {
	row := db.QueryRow(
		"SELECT 1 FROM processed_files WHERE hash = ? LIMIT 1",
		hash,
	)

	var tmp int
	err := row.Scan(&tmp)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
func isDuplicate(filePath string) bool {
	hash, err := getFileHash(filePath)
	if err != nil {
		return false
	}
	if _, exists := processedHashes[hash]; exists {
		return true
	}
	processedHashes[hash] = filePath
	return false
}
