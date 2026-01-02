package bot

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"obsidian-automation/internal/database"
	"os"
	"strings"
)

var processedHashes = make(map[string]string)

func SaveProcessed(hash, category, text, summary string, topics []string, questions []string, ai_provider string) error {
	_, err := database.DB.Exec(
		`INSERT INTO processed_files (hash, category, extracted_text, summary, topics, questions, ai_provider)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hash, category, text, summary, strings.Join(topics, ", "), strings.Join(questions, ", "), ai_provider,
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
func IsProcessed(hash string) (bool, error) {
	row := database.DB.QueryRow(
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
func IsDuplicate(filePath string) bool {
	hash, err := getFileHash(filePath)
	if err != nil {
		return false
	}
	// The `IsProcessed` function will now use the global `database.DB`
	isDup, err := IsProcessed(hash)
	if err != nil {
		// Log the error but proceed as if not a duplicate to avoid blocking
		// if the DB check fails for some reason.
		log.Printf("Error checking if hash is processed: %v", err)
		return false
	}
	if isDup {
		return true
	}

	processedHashes[hash] = filePath
	return false
}
