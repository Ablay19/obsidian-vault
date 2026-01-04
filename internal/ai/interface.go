package ai

import (
	"context"
)

// AIServiceInterface defines the methods for interacting with the AI service.
// It is used to allow for mock implementations in tests.
type AIServiceInterface interface {
	Chat(ctx context.Context, req *RequestModel, callback func(string)) error
	AnalyzeText(ctx context.Context, text, language string) (*AnalysisResult, error)
	AnalyzeTextWithParams(ctx context.Context, text, language string, task_tokens int, task_depth int, max_cost float64) (*AnalysisResult, error)
	GetActiveProviderName() string
	SetProvider(providerName string) error
	GetAvailableProviders() []string
	GetHealthyProviders(ctx context.Context) []string
	GetProvidersInfo() []ModelInfo
}
