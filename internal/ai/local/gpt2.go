package local

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tmc/langchaingo/llms"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/models"
	"obsidian-automation/pkg/utils"
)

// GPT2Provider implements local GPT-2 model
type GPT2Provider struct {
	model     llms.Model
	logger    *utils.Logger
	modelPath string
}

// NewGPT2Provider creates a new GPT-2 provider
func NewGPT2Provider(modelPath string, logger *utils.Logger) (*GPT2Provider, error) {
	// Initialize GPT-2 model
	gpt2, err := llms.NewGPT4E(llms.WithModelPath(modelPath))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GPT-2 model: %w", err)
	}

	return &GPT2Provider{
		model:     gpt2,
		logger:    logger,
		modelPath: modelPath,
	}, nil
}

// GenerateCompletion implements AIProvider interface
func (g *GPT2Provider) GenerateCompletion(ctx context.Context, req *ai.RequestModel) (*ai.ResponseModel, error) {
	start := time.Now()

	// Prepare the prompt
	messages := []llms.MessageContent{
		{
			Role:    llms.ChatMessageTypeHuman,
			Content: req.Prompt,
		},
	}

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		systemMsg := llms.MessageContent{
			Role:    llms.ChatMessageTypeSystem,
			Content: req.SystemPrompt,
		}
		messages = append([]llms.MessageContent{systemMsg}, messages...)
	}

	// Generate completion
	content, err := g.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("GPT-2 generation failed: %w", err)
	}

	// Estimate tokens (rough estimation)
	inputTokens := g.estimateTokens(req.Prompt)
	outputTokens := g.estimateTokens(content)

	duration := time.Since(start)
	g.logger.AIRequest("gpt2", "gpt2", inputTokens, 0) // TODO: Get user ID

	// Create response
	response := &ai.ResponseModel{
		Content: content,
		ModelInfo: ai.ModelInfo{
			Name:                 "gpt2-local",
			Latency:              int(duration.Milliseconds()),
			Accuracy:             0.85, // Estimated for local GPT-2
			MaxTokens:            2048,
			RateLimit:            100,
			Concurrency:          1,
			Streaming:            false,
			Enabled:              true,
			Blocked:              false,
			InputCostPerToken:    0,
			OutputCostPerToken:   0,
			MaxInputTokens:       2048,
			MaxOutputTokens:      2048,
			LatencyMsThreshold:   5000,
			AccuracyPctThreshold: 0.8,
			SupportsVision:       false,
		},
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		FinishReason: "stop",
	}

	g.logger.AIResponse("gpt2", "gpt2", outputTokens, duration.Milliseconds(), 0) // TODO: Get user ID

	return response, nil
}

// StreamCompletion implements streaming generation
func (g *GPT2Provider) StreamCompletion(ctx context.Context, req *ai.RequestModel) (<-chan ai.StreamResponse, error) {
	// For local models, we don't implement streaming for now
	return nil, fmt.Errorf("streaming not implemented for local GPT-2 model")
}

// GetModelInfo implements AIProvider interface
func (g *GPT2Provider) GetModelInfo() ai.ModelInfo {
	return ai.ModelInfo{
		Name:                 "gpt2-local",
		Latency:              200,  // Estimated
		Accuracy:             0.85, // Estimated
		MaxTokens:            2048,
		RateLimit:            100,
		Concurrency:          1,
		Streaming:            false,
		Enabled:              true,
		Blocked:              false,
		InputCostPerToken:    0,
		OutputCostPerToken:   0,
		MaxInputTokens:       2048,
		MaxOutputTokens:      2048,
		LatencyMsThreshold:   5000,
		AccuracyPctThreshold: 0.8,
		SupportsVision:       false,
	}
}

// CheckHealth implements AIProvider interface
func (g *GPT2Provider) CheckHealth(ctx context.Context) error {
	// Check if model file exists
	// For now, assume healthy if model was loaded successfully
	g.logger.Info("GPT-2 model health check", "status", "healthy")
	return nil
}

// estimateTokens provides rough token estimation
func (g *GPT2Provider) estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token for English
	return len(text) / 4
}

// Close releases resources
func (g *GPT2Provider) Close() error {
	// No specific cleanup needed for local models
	return nil
}
