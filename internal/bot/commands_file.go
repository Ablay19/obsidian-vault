package bot

import (
	"context"
	"fmt"
	"os"
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
		telemetry.ZapLogger.Sugar().Errorw("Read file error", "error", err)
		return err
	}

	job := pipeline.Job{
		ID:           fmt.Sprintf("%d_%d", message.Chat.ID, time.Now().UnixNano()),
		Source:       "telegram",
		SourceID:     strconv.FormatInt(int64(message.MessageID), 10),
		Data:         fileBytes,
		ContentType:  pipeline.ContentTypeImage,
		ReceivedAt:   time.Now(),
		MaxRetries:   3,
		OutputFormat: outputFormat,
		GitCommit:    gitCommit,
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
