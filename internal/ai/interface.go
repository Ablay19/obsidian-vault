package ai

import (
	"context"
)

// AIServiceInterface defines the methods for interacting with the AI service.
// It is used to allow for mock implementations in tests.
type AIServiceInterface interface {
	Chat(ctx context.Context, req *RequestModel, callback func(string)) error
	AnalyzeText(ctx context.Context, text, language string) (*AnalysisResult, error)
	GetActiveProviderName() string
}
