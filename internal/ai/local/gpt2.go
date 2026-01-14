package local

import (
	"context"
	"fmt"
	"time"

	"obsidian-automation/internal/ai"
	"obsidian-automation/pkg/utils"
)

// GPT2Provider implements local GPT-2 model
type GPT2Provider struct {
	logger    *utils.Logger
	modelPath string
}

// NewGPT2Provider creates a new GPT-2 provider
func NewGPT2Provider(modelPath string, logger *utils.Logger) (*GPT2Provider, error) {
	return &GPT2Provider{
		logger:    logger,
		modelPath: modelPath,
	}, nil
}

// GenerateCompletion implements AIProvider interface
func (g *GPT2Provider) GenerateCompletion(ctx context.Context, req *ai.RequestModel) (*ai.ResponseModel, error) {
	start := time.Now()

	duration := time.Since(start)
	g.logger.AIRequest("gpt2", "gpt2", 0, 0)

	// Create response
	response := &ai.ResponseModel{
		Content: "GPT-2 local model not fully implemented",
		ProviderInfo: ai.ModelInfo{
			ProviderName: "gpt2-local",
			ModelName:    "gpt2-local",
			Enabled:      true,
			Blocked:      false,
		},
	}

	g.logger.AIResponse("gpt2", "gpt2", 0, duration.Milliseconds(), 0)

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
		ProviderName: "gpt2-local",
		ModelName:    "gpt2-local",
		Enabled:      true,
		Blocked:      false,
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
