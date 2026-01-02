package ai

import (
	"context"
	"io"
	"time"
)

// AIProvider defines the interface for AI services.
type AIProvider interface {
	// Process sends a request to the AI service and returns a stream of responses.
	Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error
	// GenerateContent streams a human-readable response from AI.
	GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error)
	// GenerateJSONData gets structured data in JSON format from AI.
	GenerateJSONData(ctx context.Context, text, language string) (string, error)
	// GetModelInfo returns information about the model.
	GetModelInfo() ModelInfo
}

// ModelInfo holds information about an AI model.
type ModelInfo struct {
	ProviderName  string    `json:"provider_name"`
	ModelName     string    `json:"model_name"`
	KeyID         string    `json:"key_id,omitempty"`
	Enabled       bool      `json:"enabled"`
	Blocked       bool      `json:"blocked"`
	BlockedReason string    `json:"blocked_reason,omitempty"`
	LastUsedAt    time.Time `json:"last_used_at,omitempty"`
}