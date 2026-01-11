package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/telemetry"
)

type lastCommandHandler struct{}

func (h *lastCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	if state.LastCreatedNote != "" {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Last created note: %s", state.LastCreatedNote)))
		return err
	} else {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No note has been created yet."))
		return err
	}
}

type batchCommandHandler struct{}

func (h *batchCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	pendingFiles := state.GetPendingFiles()

	if len(pendingFiles) == 0 {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No files to process. Send some images or PDFs first."))
		return err
	}

	if len(pendingFiles) == 1 {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Only one file pending. Use /process for single files or send more files for batch processing."))
		return err
	}

	// Send initial status message
	statusMsg, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("🔄 Starting batch processing of %d files...\n⏳ This may take a few minutes.", len(pendingFiles))))
	if err != nil {
		return err
	}

	batchJobID := fmt.Sprintf("batch_%d_%d", message.Chat.ID, time.Now().UnixNano())
	state.SetBatchProcessing(true, batchJobID)

	// Process files in parallel
	processedCount := 0
	failedCount := 0

	for i, pendingFile := range pendingFiles {
		// Update progress
		progressText := fmt.Sprintf("🔄 Processing batch: %d/%d files\n📄 Current: %s",
			i+1, len(pendingFiles), filepath.Base(pendingFile.FilePath))
		cmdCtx.Bot.Request(tgbotapi.NewEditMessageText(message.Chat.ID, statusMsg.MessageID, progressText))

		// Submit to pipeline
		job := pipeline.Job{
			ID:            fmt.Sprintf("%s_file_%d", batchJobID, i),
			Source:        "telegram_batch",
			SourceID:      batchJobID,
			Data:          []byte{}, // Will be read from file
			ContentType:   pipeline.ContentTypeImage,
			FileLocalPath: pendingFile.FilePath,
			ReceivedAt:    time.Now(),
			MaxRetries:    3,
			OutputFormat:  "md",  // Default to markdown
			GitCommit:     false, // Don't auto-commit batches
			UserContext: pipeline.UserContext{
				UserID:   strconv.FormatInt(message.From.ID, 10),
				Language: state.Language,
			},
			Metadata: map[string]interface{}{
				"caption":       pendingFile.Caption,
				"chat_id":       message.Chat.ID,
				"batch_job_id":  batchJobID,
				"batch_index":   i,
				"batch_total":   len(pendingFiles),
				"status_msg_id": statusMsg.MessageID,
			},
		}

		if pendingFile.FileType == "pdf" {
			job.ContentType = pipeline.ContentTypePDF
		}

		if err := cmdCtx.IngestionPipeline.Submit(job); err != nil {
			failedCount++
			telemetry.Error("Failed to submit batch job", "error", err, "batch_id", batchJobID)
		} else {
			processedCount++
		}

		// Small delay to avoid overwhelming the pipeline
		time.Sleep(100 * time.Millisecond)
	}

	// Clear pending files
	state.ClearPendingFiles()
	state.SetBatchProcessing(false, "")

	// Final status update
	finalText := fmt.Sprintf("✅ Batch processing complete!\n📊 Results: %d processed, %d failed", processedCount, failedCount)
	cmdCtx.Bot.Request(tgbotapi.NewEditMessageText(message.Chat.ID, statusMsg.MessageID, finalText))

	return nil
}

type reprocessCommandHandler struct{}

func (h *reprocessCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	if state.LastProcessedFile != "" {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Reprocessing last file..."))
		if err != nil {
			return err
		}
		fileType := "document"
		if strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".jpg") || strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".png") {
			fileType = "image"
		}
		createObsidianNote(ctx, cmdCtx.Bot, cmdCtx.AIService, message, state, state.LastProcessedFile, fileType, 0, "")
		return nil
	} else {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No file to reprocess. Please send a file first."))
		return err
	}
}

type processCommandHandler struct{}

func (h *processCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	if !state.IsStaging {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Nothing to process. Send a file first."))
		return err
	}
	statusMsg, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Submitting to processing pipeline..."))
	if err != nil {
		return err
	}

	args := message.CommandArguments()
	outputFormat := "pdf"
	if strings.Contains(args, "--output md") {
		outputFormat = "md"
	}
	gitCommit := strings.Contains(args, "--commit")

	fileBytes, err := os.ReadFile(state.PendingFile)
	if err != nil {
		cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Failed to read staged file."))
		telemetry.Error("Read file error: " + err.Error())
		return err
	}

	job := pipeline.Job{
		ID:            fmt.Sprintf("%d_%d", message.Chat.ID, time.Now().UnixNano()),
		Source:        "telegram",
		SourceID:      strconv.FormatInt(int64(message.MessageID), 10),
		Data:          fileBytes,
		ContentType:   pipeline.ContentTypeImage,
		FileLocalPath: state.PendingFile,
		ReceivedAt:    time.Now(),
		MaxRetries:    3,
		OutputFormat:  outputFormat,
		GitCommit:     gitCommit,
		UserContext: pipeline.UserContext{
			UserID:   strconv.FormatInt(message.From.ID, 10),
			Language: state.Language,
		},
		Metadata: map[string]interface{}{
			"caption": state.PendingContext,
			"chat_id": message.Chat.ID,
		},
	}

	if state.PendingFileType == "pdf" {
		job.ContentType = pipeline.ContentTypePDF
	}

	if err := cmdCtx.IngestionPipeline.Submit(job); err != nil {
		_, sendErr := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Pipeline full/error: %v", err)))
		return sendErr
	} else {
		cmdCtx.Bot.Request(tgbotapi.NewEditMessageText(message.Chat.ID, statusMsg.MessageID, "Job queued. You will be notified when complete."))
	}

	state.IsStaging = false
	state.PendingFile = ""
	state.PendingContext = ""
	return nil
}

func (h *modeCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Create inline keyboard for processing mode selection
	var keyboard [][]tgbotapi.InlineKeyboardButton

	modes := []struct {
		name        string
		description string
		callback    string
	}{
		{"Fast", "Quick processing, basic analysis", "mode_fast"},
		{"Quality", "Thorough analysis, multiple strategies", "mode_quality"},
		{"Conservative", "Only use proven methods", "mode_conservative"},
		{"Experimental", "Try new processing techniques", "mode_experimental"},
	}

	for _, mode := range modes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s: %s", mode.name, mode.description),
			mode.callback,
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🎯 **Select Processing Mode**\n\nCurrent: %s\n\nChoose how you want files processed:", "Quality"))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err := cmdCtx.Bot.Send(msg)
	return err
}

func (h *botsCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Create inline keyboard for bot instance selection
	var keyboard [][]tgbotapi.InlineKeyboardButton

	botInstances := []struct {
		name     string
		status   string
		callback string
	}{
		{"Main Bot", "Active", "bot_main"},
		{"Test Bot", "Available", "bot_test"},
		{"Dev Bot", "Offline", "bot_dev"},
		{"Backup Bot", "Standby", "bot_backup"},
	}

	for _, bot := range botInstances {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🤖 %s (%s)", bot.name, bot.status),
			bot.callback,
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "🔄 **Select Bot Instance**\n\nChoose which bot instance to interact with:")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err := cmdCtx.Bot.Send(msg)
	return err
}
