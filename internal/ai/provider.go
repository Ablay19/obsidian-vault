package ai

import "context"

const (
	ModelFlashSearch = "gemini-pro"
	ModelProComplex  = "gemini-1.5-pro-latest"
	ModelImageGen    = "gemini-pro"
)

// AIProvider defines the interface for an AI service.
// This allows for multiple AI providers (e.g., Gemini, Groq) to be used interchangeably.
type AIProvider interface {
	GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error)
	GenerateJSONData(ctx context.Context, text, language string) (string, error)
	ProviderName() string
}
