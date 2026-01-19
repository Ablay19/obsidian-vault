package ai

import "time"

// GenerationOptions represents AI generation parameters
type GenerationOptions struct {
	Model        string  `json:"model"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	TopP         float64 `json:"top_p"`
	Stream       bool    `json:"stream"`
	SystemPrompt string  `json:"system_prompt"`
	UserID       string  `json:"user_id"`
}

// GenerationResult represents the result of AI generation
type GenerationResult struct {
	Content      string        `json:"content"`
	Model        string        `json:"model"`
	Provider     string        `json:"provider"`
	UserID       string        `json:"user_id"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	Cost         float64       `json:"cost"`
	Latency      time.Duration `json:"latency"`
	FinishReason string        `json:"finish_reason"`
}

// RequestModel represents an AI request
type RequestModel struct {
	Prompt    string                 `json:"prompt"`
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	Options   map[string]interface{} `json:"options"`
}

// ResponseModel represents an AI response
type ResponseModel struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Tokens  int    `json:"tokens"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// ModelInfo represents information about an AI model
type ModelInfo struct {
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	MaxTokens      int    `json:"max_tokens"`
	SupportsStream bool   `json:"supports_stream"`
}
