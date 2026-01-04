package bot

import (
	"context"
	"database/sql"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/database/sqlc"
	"strings"
	"time"
)

// SaveProcessed saves a processed file entry to the database using sqlc.
func SaveProcessed(ctx context.Context, hash, fileName, filePath, contentType, category, text, summary string, topics []string, questions []string, ai_provider string, userID int64) error {
	params := sqlc.InsertProcessedFileParams{
		Hash:        hash,
		FileName:    fileName,
		FilePath:    filePath,
		ContentType: sql.NullString{String: contentType, Valid: contentType != ""},
		Status:      "processed",
		Summary:     sql.NullString{String: summary, Valid: summary != ""},
		Topics:      sql.NullString{String: strings.Join(topics, ", "), Valid: len(topics) > 0},
		Questions:   sql.NullString{String: strings.Join(questions, ", "), Valid: len(questions) > 0},
		AiProvider:  sql.NullString{String: ai_provider, Valid: ai_provider != ""},
		UserID:      sql.NullInt64{Int64: userID, Valid: userID != 0},
		CreatedAt:   time.Now(),
		UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
	}
	return database.Client.Queries.InsertProcessedFile(ctx, params)
}
