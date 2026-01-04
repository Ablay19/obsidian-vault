package bot

import (
	"obsidian-automation/internal/database"
	"strings"
)

// getFileHash computes the SHA256 hash of a file.
func SaveProcessed(hash, category, text, summary string, topics []string, questions []string, ai_provider string, userID int64) error {
	_, err := database.DB.Exec(
		`INSERT INTO processed_files (hash, category, extracted_text, summary, topics, questions, ai_provider, user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		hash, category, text, summary, strings.Join(topics, ", "), strings.Join(questions, ", "), ai_provider, userID,
	)
	return err
}
