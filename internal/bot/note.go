package bot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"obsidian-automation/internal/ai"
)

// createObsidianNote orchestrates the whole process of creating an Obsidian note.
func createObsidianNote(ctx context.Context, bot Bot, aiService ai.AIServiceInterface, message *tgbotapi.Message, state *UserState, filePath, fileType string, messageID int, additionalContext string) {
	updateStatus := func(status string) {
		if messageID != 0 {
			bot.Send(tgbotapi.NewEditMessageText(message.Chat.ID, messageID, status))
		}
	}

	streamCallback := func(chunk string) {
		// This could be used to stream the response to the user in real-time
	}

	content := processFileWithAI(ctx, filePath, fileType, aiService, streamCallback, state.Language, updateStatus, additionalContext)

	if content.Category == "unprocessed" || content.Category == "error" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Could not process the file."))
		return
	}

	// Create note content
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# %s\n\n", time.Now().Format("2006-01-02_15-04-05")))
	builder.WriteString(fmt.Sprintf("**Category:** %s\n", content.Category))
	builder.WriteString(fmt.Sprintf("**AI Provider:** %s\n", content.AIProvider))
	builder.WriteString(fmt.Sprintf("**Tags:** #%s\n\n", strings.Join(content.Tags, " #")))

	if content.Summary != "" {
		builder.WriteString("## Summary\n")
		builder.WriteString(content.Summary + "\n\n")
	}
	if len(content.Topics) > 0 {
		builder.WriteString("## Key Topics\n")
		for _, topic := range content.Topics {
			builder.WriteString(fmt.Sprintf("- %s\n", topic))
		}
		builder.WriteString("\n")
	}
	if len(content.Questions) > 0 {
		builder.WriteString("## Review Questions\n")
		for _, q := range content.Questions {
			builder.WriteString(fmt.Sprintf("- %s\n", q))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("## Extracted Text\n")
	builder.WriteString("```\n")
	builder.WriteString(content.Text)
	builder.WriteString("\n```\n")

	// Save the note
	noteFilename := fmt.Sprintf("%s_%s.md", time.Now().Format("20060102_150405"), content.Category)
	notePath := filepath.Join("vault", "Inbox", noteFilename)
	err := os.WriteFile(notePath, []byte(builder.String()), 0644)
	if err != nil {
		zap.S().Error("Error writing note file", "error", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error saving the note."))
		return
	}

	// Save to database
	hash, err := getFileHash(filePath)
	if err != nil {
		zap.S().Error("Error getting file hash", "error", err)
	} else {
		err := SaveProcessed(hash, content.Category, content.Text, content.Summary, content.Topics, content.Questions, content.AIProvider, message.From.ID)
		if err != nil {
			zap.S().Error("Error saving processed file to DB", "error", err)
		}
	}

	// Organize the note
	organizeNote(notePath, content.Category)

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Note '%s' created successfully!", noteFilename)))
	state.LastCreatedNote = noteFilename
	state.LastProcessedFile = filePath
}

func organizeNote(filePath, category string) {
	if category == "" || category == "general" {
		return // No organization needed
	}

	// Ensure category directory exists
	dirPath := filepath.Join("vault", category)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0755)
	}

	// Move the file
	newPath := filepath.Join(dirPath, filepath.Base(filePath))
	os.Rename(filePath, newPath)
}

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
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
