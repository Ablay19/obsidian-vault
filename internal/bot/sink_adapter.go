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
	"obsidian-automation/internal/telemetry"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createProgressBar creates a visual progress bar
func createProgressBar(percent int) string {
	const barLength = 10
	filled := percent * barLength / 100
	empty := barLength - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %d%%", bar, percent)
}

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

	telemetry.Info("Sink saving result", "job_id", result.JobID, "category", content.Category)

	// Check if this is part of a batch job
	isBatchJob := false
	var batchJobID string
	var batchIndex, batchTotal int
	var statusMsgID int
	var chatID int64

	if job.Metadata != nil {
		if bid, ok := job.Metadata["batch_job_id"].(string); ok && bid != "" {
			isBatchJob = true
			batchJobID = bid
		}
		if idx, ok := job.Metadata["batch_index"].(int); ok {
			batchIndex = idx
		}
		if total, ok := job.Metadata["batch_total"].(int); ok {
			batchTotal = total
		}
		if msgID, ok := job.Metadata["status_msg_id"].(int); ok {
			statusMsgID = msgID
		}
		if cid, ok := job.Metadata["chat_id"].(int64); ok {
			chatID = cid
		}
	}

	// Track all processing steps - only send success if EVERYTHING works
	var processingErrors []string
	var notePath, pdfPath string

	// Step 1: Create Note Content
	var builder strings.Builder
	title := fmt.Sprintf("Document Analysis - %s", content.Category)
	if len(content.Topics) > 0 {
		title = content.Topics[0] // Use first topic as title
	}
	builder.WriteString(fmt.Sprintf("# %s\n\n", title))
	builder.WriteString(fmt.Sprintf("**Processed:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
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

	// Step 2: Write Markdown to Disk (CRITICAL)
	noteFilename := fmt.Sprintf("%s_%s.md", time.Now().Format("20060102_150405"), content.Category)
	os.MkdirAll(filepath.Join("vault", "Inbox"), 0755)
	notePath = filepath.Join("vault", "Inbox", noteFilename)

	if err := os.WriteFile(notePath, []byte(markdownContent), 0644); err != nil {
		processingErrors = append(processingErrors, fmt.Sprintf("Failed to write note file: %v", err))
		return fmt.Errorf("failed to write note file: %w", err)
	}

	// Step 3: Optional PDF Generation (can fail gracefully)
	if job.OutputFormat == "pdf" {
		pdfFilename := fmt.Sprintf("%s_%s.pdf", time.Now().Format("20060102_150405"), content.Category)
		os.MkdirAll("pdfs", 0755)
		pdfPath = filepath.Join("pdfs", pdfFilename)
		if err := converter.ConvertMarkdownToPDF(markdownContent, pdfPath); err != nil {
			telemetry.Warn("PDF conversion failed, continuing without PDF: " + err.Error())
			pdfPath = "" // Clear path so we don't try to send PDF
		}
	}

	// Step 4: Save to Database (non-blocking, can fail gracefully)
	hashBytes := sha256.Sum256(job.Data)
	hash := hex.EncodeToString(hashBytes[:])
	userID, _ := strconv.ParseInt(job.UserContext.UserID, 10, 64)

	go func() {
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
			telemetry.Error("Failed to save to DB (non-critical): " + err.Error())
		} else {
			telemetry.Info("File processing saved to database successfully")
		}
	}()

	// Step 5: Git Automation (non-blocking, can fail gracefully)
	if job.GitCommit && s.gitManager != nil {
		go func() {
			commitMsg := fmt.Sprintf("chore: auto commit document %s info about %s", noteFilename, content.Category)
			if err := s.gitManager.SyncAutoCommit(commitMsg); err != nil {
				telemetry.Error("Git sync failed (non-critical): " + err.Error())
			} else {
				telemetry.Info("Git sync successful for job: " + job.ID)
			}
		}()
	}

	// Step 6: Notify User ONLY if note creation succeeded (CRITICAL)
	if s.bot != nil && userID != 0 && len(processingErrors) == 0 {
		// For batch jobs, update progress instead of sending individual messages
		if isBatchJob && statusMsgID != 0 {
			progressPercent := int(float64(batchIndex+1) / float64(batchTotal) * 100)
			progressBar := createProgressBar(progressPercent)
			progressText := fmt.Sprintf("🔄 Processing batch: %d/%d files (%d%%)\n%s\n✅ %s - Done!",
				batchIndex+1, batchTotal, progressPercent, progressBar, noteFilename)
			s.bot.Request(tgbotapi.NewEditMessageText(chatID, statusMsgID, progressText))
		} else {
			// Send individual success message for non-batch jobs
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("✅ Note '%s' created successfully!", noteFilename))
			if _, err := s.bot.Send(msg); err != nil {
				telemetry.Error("Failed to send success message: " + err.Error())
			}

			// Send PDF only if it was successfully created
			if pdfPath != "" {
				doc := tgbotapi.NewDocument(userID, tgbotapi.FilePath(pdfPath))
				doc.Caption = fmt.Sprintf("📄 PDF Version: %s", noteFilename)
				if _, err := s.bot.Send(doc); err != nil {
					telemetry.Warn("Failed to send PDF (non-critical): " + err.Error())
				}
			}
		}

		telemetry.Info("User notification sent successfully", "user_id", userID, "note", noteFilename, "is_batch", isBatchJob, "batch_id", batchJobID)
	} else if len(processingErrors) > 0 {
		// Send failure message if critical steps failed
		errorMsg := fmt.Sprintf("❌ Processing failed: %s", processingErrors[0])

		// For batch jobs, update progress with failure
		if isBatchJob && statusMsgID != 0 {
			progressPercent := int(float64(batchIndex+1) / float64(batchTotal) * 100)
			progressBar := createProgressBar(progressPercent)
			progressText := fmt.Sprintf("🔄 Processing batch: %d/%d files (%d%%)\n%s\n❌ %s - Failed: %s",
				batchIndex+1, batchTotal, progressPercent, progressBar, filepath.Base(job.FileLocalPath), processingErrors[0])
			s.bot.Request(tgbotapi.NewEditMessageText(chatID, statusMsgID, progressText))
		} else if s.bot != nil && userID != 0 {
			msg := tgbotapi.NewMessage(userID, errorMsg)
			s.bot.Send(msg)
		}

		return fmt.Errorf("processing failed: %v", processingErrors)
	}

	return nil
}
