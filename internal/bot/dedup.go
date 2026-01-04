package bot

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"obsidian-automation/internal/database"
	"strings"
)

// getFileHash computes the SHA256 hash of a file.
func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func SaveProcessed(hash, category, text, summary string, topics []string, questions []string, ai_provider string, userID int64) error {
	_, err := database.DB.Exec(
		`INSERT INTO processed_files (hash, category, extracted_text, summary, topics, questions, ai_provider, user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		hash, category, text, summary, strings.Join(topics, ", "), strings.Join(questions, ", "), ai_provider, userID,
	)
	return err
}