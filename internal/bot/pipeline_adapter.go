package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/pipeline"
)

// BotProcessor implements the pipeline.Processor interface.
type BotProcessor struct {
	aiService *ai.AIService
	tmpDir    string
}

// NewBotProcessor creates a new processor for the bot.
func NewBotProcessor(aiService *ai.AIService) *BotProcessor {
	return &BotProcessor{
		aiService: aiService,
		tmpDir:    os.TempDir(),
	}
}

// Process handles a single job by calling the existing processFileWithAI logic.
func (p *BotProcessor) Process(ctx context.Context, job pipeline.Job) (pipeline.Result, error) {
	// 1. Write Data to Temp File
	// processFileWithAI expects a file path.
	ext := ".tmp"
	fileType := "unknown"

	switch job.ContentType {
	case pipeline.ContentTypeImage:
		ext = ".jpg" // Default assumption
		fileType = "image"
	case pipeline.ContentTypePDF:
		ext = ".pdf"
		fileType = "pdf"
	case pipeline.ContentTypeText:
		ext = ".txt"
		fileType = "text"
	}

	tmpFile := filepath.Join(p.tmpDir, fmt.Sprintf("pipeline_%s%s", job.ID, ext))
	if err := os.WriteFile(tmpFile, job.Data, 0644); err != nil {
		return pipeline.Result{}, fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tmpFile) // Cleanup

	// 2. Prepare callbacks
	// For now, we don't have a live socket to stream back to the user in this pipeline flow,
	// unless we pass a callback function in the Job metadata or UserContext.
	// We'll use a no-op for now or log updates.
	updateStatus := func(s string) {
		slog.Debug("Pipeline Status Update", "job_id", job.ID, "status", s)
	}
	streamCallback := func(chunk string) {
		// potentially accumulate or stream if we had a channel in Job
	}

	additionalContext := ""
	if val, ok := job.Metadata["caption"]; ok {
		additionalContext = val.(string)
	}

	// 3. Call existing logic
	// Note: processFileWithAI is in the same package (bot), so we can call it.
	content := processFileWithAI(
		ctx,
		tmpFile,
		fileType,
		p.aiService,
		streamCallback,
		job.UserContext.Language,
		updateStatus,
		additionalContext,
	)

	// 4. Return Result
	if content.Category == "unprocessed" && len(content.Tags) > 0 && content.Tags[0] == "error" {
		return pipeline.Result{
			JobID:       job.ID,
			Success:     false,
			Error:       fmt.Errorf("processing failed inside processFileWithAI"),
			ProcessedAt: time.Now(),
		}, nil
	}

	return pipeline.Result{
		JobID:       job.ID,
		Success:     true,
		ProcessedAt: time.Now(),
		Output:      content,
	}, nil
}
