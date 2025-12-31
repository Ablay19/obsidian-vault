package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

const (
	ModelFlashSearch = "gemini-1.5-flash"
	ModelProComplex  = "gemini-1.5-pro"
	ModelImageGen    = "gemini-1.5-flash"
)

// AIService manages multiple AI providers and selects the active one.
type AIService struct {
	providers      map[string]AIProvider
	activeProvider AIProvider
	mu             sync.RWMutex
}

// NewAIService initializes all available AI providers.
func NewAIService(ctx context.Context) *AIService {
	providers := make(map[string]AIProvider)

	// Initialize Gemini provider
	geminiProvider := NewGeminiProvider(ctx)
	if geminiProvider != nil {
		providers[geminiProvider.ProviderName()] = geminiProvider
	}

	// Initialize Groq provider
	groqProvider := NewGroqProvider(ctx)
	if groqProvider != nil {
		providers[groqProvider.ProviderName()] = groqProvider
	}

	if len(providers) == 0 {
		log.Println("No AI providers could be initialized. AI features will be unavailable.")
		return nil
	}

	// Set default provider (e.g., Gemini)
	activeProvider, ok := providers["Gemini"]
	if !ok {
		// Fallback to the first available provider if Gemini isn't there
		for _, p := range providers {
			activeProvider = p
			break
		}
	}

	log.Printf("AI Service initialized. Available providers: %v. Active provider: %s", getMapKeys(providers), activeProvider.ProviderName())

	return &AIService{
		providers:      providers,
		activeProvider: activeProvider,
	}
}

// SetProvider changes the active AI provider.
func (s *AIService) SetProvider(providerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("provider '%s' not found or not configured", providerName)
	}
	s.activeProvider = provider
	log.Printf("Switched AI provider to %s", providerName)
	return nil
}

// GetActiveProviderName returns the name of the currently active provider.
func (s *AIService) GetActiveProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.activeProvider == nil {
		return "None"
	}
	return s.activeProvider.ProviderName()
}

// GetAvailableProviders returns a list of available provider names.
func (s *AIService) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return getMapKeys(s.providers)
}

// GenerateContent delegates the call to the active provider.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	s.mu.RLock()
	provider := s.activeProvider
	s.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("no active AI provider")
	}
	return provider.GenerateContent(ctx, prompt, imageData, modelType, streamCallback)
}

// GenerateJSONData delegates the call to the active provider.
func (s *AIService) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	s.mu.RLock()
	provider := s.activeProvider
	s.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("no active AI provider")
	}
	return provider.GenerateJSONData(ctx, text, language)
}

// Helper function to get keys from a map.
func getMapKeys(m map[string]AIProvider) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
