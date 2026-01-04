package bot

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"obsidian-automation/internal/bot/converter"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// BotSink implements pipeline.Sink to save results to Obsidian/Database.
type BotSink struct {
	db         *sql.DB
	bot        *tgbotapi.BotAPI
	gitManager *git.Manager
}

func NewBotSink(db *sql.DB, bot *tgbotapi.BotAPI, gitManager *git.Manager) *BotSink {
	return &BotSink{db: db, bot: bot, gitManager: gitManager}
}

func (s *BotSink) Save(ctx context.Context, job pipeline.Job, result pipeline.Result) error {
	content, ok := result.Output.(ProcessedContent)
	if !ok {
		return fmt.Errorf("invalid output type")
	}

	zap.S().Info("Sink saving result", "job_id", result.JobID, "category", content.Category)

	// 1. Create Note Content
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

	markdownContent := builder.String()

	// 2. Write Markdown to Disk
	noteFilename := fmt.Sprintf("%s_%s.md", time.Now().Format("20060102_150405"), content.Category)
	// Ensure vault/Inbox exists
	os.MkdirAll(filepath.Join("vault", "Inbox"), 0755)
	notePath := filepath.Join("vault", "Inbox", noteFilename)

	if err := os.WriteFile(notePath, []byte(markdownContent), 0644); err != nil {
		return fmt.Errorf("failed to write note file: %w", err)
	}

	// 3. Optional PDF Generation
	var pdfPath string
	if job.OutputFormat == "pdf" {
		pdfFilename := fmt.Sprintf("%s_%s.pdf", time.Now().Format("20060102_150405"), content.Category)
		os.MkdirAll("pdfs", 0755)
		pdfPath = filepath.Join("pdfs", pdfFilename)
		if err := converter.ConvertMarkdownToPDF(markdownContent, pdfPath); err != nil {
			telemetry.ZapLogger.Sugar().Errorw("PDF conversion failed", "error", err)
		}
	}

	// 4. Save to Database
	// Hash the original data from job
	hashBytes := sha256.Sum256(job.Data)
	hash := hex.EncodeToString(hashBytes[:])

	userID, _ := strconv.ParseInt(job.UserContext.UserID, 10, 64)

	// Use package-level function SaveProcessed (from dedup.go)
	if err := SaveProcessed(
		ctx,
		hash,
		filepath.Base(job.FileLocalPath),
		job.FileLocalPath,
		job.ContentType.String(),
		content.Category,
		content.Text,
		content.Summary,
		content.Topics,
		content.Questions,
		content.AIProvider,
		userID,
	); err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to save to DB", "error", err)
	}

	// 5. Organize - this functionality has been removed, manual organization expected for now
	// organizeNote(notePath, content.Category)

	// 6. Git Automation
	if job.GitCommit && s.gitManager != nil {
		go func() {
			commitMsg := fmt.Sprintf("chore: auto commit document %s info about %s", noteFilename, content.Category)
			if err := s.gitManager.SyncAutoCommit(commitMsg); err != nil {
				telemetry.ZapLogger.Sugar().Errorw("Git sync failed", "job_id", job.ID, "error", err)
			} else {
				telemetry.ZapLogger.Sugar().Infow("Git sync successful", "job_id", job.ID)
			}
		}()
	}

	// 7. Notify User
	if s.bot != nil && userID != 0 {
		msg := tgbotapi.NewMessage(userID, fmt.Sprintf("✅ Note '%s' created!", noteFilename))
		s.bot.Send(msg)

		if pdfPath != "" {
			doc := tgbotapi.NewDocument(userID, tgbotapi.FilePath(pdfPath))
			doc.Caption = fmt.Sprintf("📄 PDF Version: %s", noteFilename)
			s.bot.Send(doc)
		}
	}

	return nil
}
