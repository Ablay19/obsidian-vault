package ai

import "context"

// MockGeminiProvider is a mock implementation of the AIProvider interface for testing.
type MockGeminiProvider struct{}

func (p *MockGeminiProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	return "Gemini response", nil
}

func (p *MockGeminiProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	return `{"category": "gemini"}`, nil
}

func (p *MockGeminiProvider) ProviderName() string {
	return "Gemini"
}

// MockGroqProvider is a mock implementation of the AIProvider interface for testing.
type MockGroqProvider struct{}

func (p *MockGroqProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	return "Groq response", nil
}

func (p *MockGroqProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	return `{"category": "groq"}`, nil
}

func (p *MockGroqProvider) ProviderName() string {
	return "Groq"
}
