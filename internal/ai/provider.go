package ai

import (
	"context"
	"time"
)

// AIProvider defines the interface for AI services.
type AIProvider interface {
	// GenerateCompletion sends a request to the AI service and returns a complete response.
	GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error)

	// StreamCompletion streams the response from the AI service.
	StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error)

	// GetModelInfo returns information about the model.
	GetModelInfo() ModelInfo

	// CheckHealth verifies if the provider is currently operational.
	CheckHealth(ctx context.Context) error
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
