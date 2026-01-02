package ai

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	cfg "obsidian-automation/internal/config"
)

// AIService manages multiple AI providers and selects the active one.
type AIService struct {
	providers      map[string]AIProvider
	ActiveProvider AIProvider
	mu             sync.RWMutex
}

// NewAIService initializes the AI service with provided providers.
// If no providers are given, it attempts to initialize Gemini and Groq providers from environment variables.
func NewAIService(ctx context.Context, appConfig *cfg.Config, initialProviders ...AIProvider) *AIService {
	providers := make(map[string]AIProvider)

	if len(initialProviders) > 0 {
		for _, p := range initialProviders {
			providers[p.GetModelInfo().ProviderName] = p
		}
	} else {
		// Existing logic to initialize Gemini and Groq from env vars
		// Initialize Gemini provider
		geminiProvider := NewGeminiProvider(ctx)
		if geminiProvider != nil {
			providers[geminiProvider.GetModelInfo().ProviderName] = geminiProvider
		}

		// Initialize Groq provider
		groqProvider := NewGroqProvider(ctx)
		if groqProvider != nil {
			providers[groqProvider.GetModelInfo().ProviderName] = groqProvider
		}

		// Initialize ONNX provider
		if appConfig.Providers.ONNX.ModelPath != "" {
			onnxProvider, err := NewONNXProvider(appConfig.Providers.ONNX.ModelPath)
			if err != nil {
				log.Printf("Warning: Failed to initialize ONNX provider: %v", err)
			} else {
				log.Println("✅ ONNX provider initialized successfully")
				providers["ONNX"] = onnxProvider
			}
		}

	}

	if len(providers) == 0 {
		log.Println("No AI providers could be initialized. AI features will be unavailable.")
		return nil
	}

	// Set default provider (e.g., Gemini)
	ActiveProvider, ok := providers["Gemini"]
	if !ok {
		// Fallback to the first available provider if Gemini isn't there
		for _, p := range providers {
			ActiveProvider = p
			break
		}
	}

	log.Printf("AI Service initialized. Available providers: %v. Active provider: %s", getMapKeys(providers), ActiveProvider.GetModelInfo().ProviderName)

	return &AIService{
		providers:      providers,
		ActiveProvider: ActiveProvider,
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
	s.ActiveProvider = provider
	log.Printf("Switched AI provider to %s", providerName)
	return nil
}

// GetActiveProviderName returns the name of the currently active provider.
func (s *AIService) GetActiveProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ActiveProvider == nil {
		return "None"
	}
	return s.ActiveProvider.GetModelInfo().ProviderName
}

// GetActiveProvider returns the active AI provider.
func (s *AIService) GetActiveProvider() AIProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ActiveProvider
}

// GetAvailableProviders returns a list of available provider names.
func (s *AIService) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return getMapKeys(s.providers)
}

// GetProvidersInfo returns a list of model information for all available providers.
func (s *AIService) GetProvidersInfo() []ModelInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []ModelInfo
	for _, p := range s.providers {
		infos = append(infos, p.GetModelInfo())
	}
	return infos
}

// Process delegates the call to the active provider.
func (s *AIService) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	s.mu.RLock()
	provider := s.ActiveProvider
	s.mu.RUnlock()

	if provider == nil {
		return fmt.Errorf("no active AI provider")
	}
	return provider.Process(ctx, w, system, prompt, images)
}

// GenerateContent delegates the call to the active provider.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	s.mu.RLock()
	provider := s.ActiveProvider
	s.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("no active AI provider")
	}
	return provider.GenerateContent(ctx, prompt, imageData, modelType, streamCallback)
}

// GenerateJSONData delegates the call to the active provider.
func (s *AIService) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	s.mu.RLock()
	provider := s.ActiveProvider
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
