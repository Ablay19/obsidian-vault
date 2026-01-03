package ai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	st "obsidian-automation/internal/state" // Import the state package
)

// AIService manages multiple AI providers and selects the active one.
type AIService struct {
	// providers maps provider name to a map of keyID to AIProvider instance
	providers map[string]map[string]AIProvider
	sm        *st.RuntimeConfigManager // Reference to the RuntimeConfigManager
	mu        sync.RWMutex
}

// NewAIService initializes the AI service using the RuntimeConfigManager.
func NewAIService(ctx context.Context, sm *st.RuntimeConfigManager) *AIService {
	s := &AIService{
		providers: make(map[string]map[string]AIProvider),
		sm:        sm,
	}

	s.initializeProviders(ctx)

	// Check if any actual providers were initialized
	hasInitializedProviders := false
	for _, keyProviders := range s.providers {
		if len(keyProviders) > 0 {
			hasInitializedProviders = true
			break
		}
	}

	if !hasInitializedProviders {
		slog.Warn("No AI providers could be initialized from RuntimeConfigManager. AI features will be unavailable.")
		return nil
	}

	slog.Info("AI Service initialized.", "available_providers", s.GetAvailableProviders())

	return s
}

// initializeProviders populates the providers map based on the current RuntimeConfig.
func (s *AIService) initializeProviders(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.providers = make(map[string]map[string]AIProvider) // Clear existing

	currentConfig := s.sm.GetConfig()

	for providerName, providerState := range currentConfig.Providers {
		if !providerState.Enabled {
			continue // Skip globally disabled providers
		}

		s.providers[providerName] = make(map[string]AIProvider)
		for keyID, keyState := range currentConfig.APIKeys {
			if keyState.Provider == providerName && keyState.Enabled && !keyState.Blocked {
				if keyState.Value == "" {
					slog.Warn("Skipping provider key due to empty API key.", "provider", providerName, "key_id_partial", truncateString(keyID, 8))
					continue
				}

				var provider AIProvider // This is the interface type
				modelName := providerState.ModelName

				// Use temporary concrete pointers to check for nil correctly
				var tempGeminiProvider *GeminiProvider
				var tempGroqProvider *GroqProvider
				var tempHuggingFaceProvider *HuggingFaceProvider
				var tempOpenRouterProvider *OpenRouterProvider

				switch providerName {
				case "Gemini":
					tempGeminiProvider = NewGeminiProvider(ctx, keyState.Value, modelName)
					if tempGeminiProvider == nil { // Check concrete type directly
						slog.Error("Failed to initialize Gemini provider, skipping.", "key_id_partial", truncateString(keyID, 8))
						continue
					}
					provider = tempGeminiProvider
				case "Groq":
					tempGroqProvider = NewGroqProvider(keyState.Value, modelName)
					if tempGroqProvider == nil { // Check concrete type directly
						slog.Error("Failed to initialize Groq provider, skipping.", "key_id_partial", truncateString(keyID, 8))
						continue
					}
					provider = tempGroqProvider
				case "Hugging Face":
					tempHuggingFaceProvider = NewHuggingFaceProvider(keyState.Value, modelName)
					if tempHuggingFaceProvider == nil { // Check concrete type directly
						slog.Error("Failed to initialize Hugging Face provider, skipping.", "key_id_partial", truncateString(keyID, 8))
						continue
					}
					provider = tempHuggingFaceProvider
				case "OpenRouter":
					tempOpenRouterProvider = NewOpenRouterProvider(keyState.Value, modelName)
					if tempOpenRouterProvider == nil { // Check concrete type directly
						slog.Error("Failed to initialize OpenRouter provider, skipping.", "key_id_partial", truncateString(keyID, 8))
						continue
					}
					provider = tempOpenRouterProvider
				default:
					slog.Warn("Unknown provider type, skipping.", "provider_type", providerName, "key_id", keyID)
					continue
				}

				s.providers[providerName][keyID] = provider // Now 'provider' should genuinely be non-nil if reached here
				slog.Info("Initialized provider.", "provider", providerName, "key_id_partial", truncateString(keyID, 8))
			}
		}
	}
}

// truncateString safely truncates a string for logging purposes.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SetProvider changes the active AI provider preference in RuntimeConfig.
func (s *AIService) SetProvider(providerName string) error {
	if err := s.sm.SetActiveProvider(providerName); err != nil {
		return err
	}
	slog.Info("Switched active AI provider preference.", "provider", providerName)
	return nil
}

// GetActiveProviderName returns the name of the currently active provider based on RuntimeConfigManager.
func (s *AIService) GetActiveProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	currentConfig := s.sm.GetConfig()
	if currentConfig.ActiveProvider != "" {
		return currentConfig.ActiveProvider
	}
	// Fallback to first enabled if none preferred
	for name, ps := range currentConfig.Providers {
		if ps.Enabled {
			return name
		}
	}
	return "None"
}

// GetActiveProvider returns an active AIProvider instance for the currently preferred provider.
// This method should select an appropriate key based on availability and health.
func (s *AIService) GetActiveProvider(ctx context.Context) (AIProvider, st.APIKeyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentConfig := s.sm.GetConfig()
	preferredProviderName := currentConfig.ActiveProvider

	if preferredProviderName == "" || preferredProviderName == "None" {
		// Fallback to first enabled if none preferred
		for name, ps := range currentConfig.Providers {
			if ps.Enabled {
				preferredProviderName = name
				break
			}
		}
	}

	if preferredProviderName == "" {
		return nil, st.APIKeyState{}, fmt.Errorf("no enabled provider found in runtime config")
	}

	return s.selectActiveKeyForProvider(ctx, preferredProviderName)
}

// GetAvailableProviders returns a list of available provider names.
func (s *AIService) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.providers))
	for k := range s.providers {
		if len(s.providers[k]) > 0 { // Only list providers with at least one active key
			keys = append(keys, k)
		}
	}
	return keys
}

// GetHealthyProviders returns a list of provider names that are currently operational.
func (s *AIService) GetHealthyProviders(ctx context.Context) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var healthy []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for providerName, keyProviders := range s.providers {
		for _, provider := range keyProviders {
			wg.Add(1)
			go func(name string, p AIProvider) {
				defer wg.Done()
				// Use a shorter timeout for health checks
				healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				if err := p.CheckHealth(healthCtx); err == nil {
					mu.Lock()
					// Check if already added
					alreadyAdded := false
					for _, h := range healthy {
						if h == name {
							alreadyAdded = true
							break
						}
					}
					if !alreadyAdded {
						healthy = append(healthy, name)
					}
					mu.Unlock()
				} else {
					slog.Warn("Provider health check failed", "provider", name, "error", err)
				}
			}(providerName, provider)
			break // Only need to check one provider (key) per provider type for overall health
		}
	}

	wg.Wait()
	return healthy
}

// GetProvidersInfo returns a list of model information for all available providers and their active keys.
func (s *AIService) GetProvidersInfo() []ModelInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []ModelInfo
	for _, keyProviders := range s.providers {
		for keyID, provider := range keyProviders {
			info := provider.GetModelInfo()
			// Add key-specific info
			if keyState, ok := s.sm.GetConfig().APIKeys[keyID]; ok {
				info.KeyID = keyID
				info.Enabled = keyState.Enabled
				info.Blocked = keyState.Blocked
				info.BlockedReason = keyState.BlockedReason
				info.LastUsedAt = keyState.LastUsedAt
			}
			infos = append(infos, info)
		}
	}
	return infos
}

// Process delegates the call to the active provider.
func (s *AIService) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	provider, keyState, err := s.GetActiveProvider(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active provider: %w", err)
	}

	// Enforce RuntimeConfig checks for the selected key
	if err := s.checkRuntimeConfig(keyState.ID); err != nil {
		return err
	}

	callErr := provider.Process(ctx, w, system, prompt, images)
	s.sm.UpdateKeyUsage(keyState.ID, func() string {
		if callErr != nil {
			return callErr.Error()
		}
		return ""
	}(), -1) // Update key usage regardless of success or failure. Quota not tracked here.

	if callErr != nil {
		return fmt.Errorf("provider '%s' (key: %s) Process failed: %w", provider.GetModelInfo().ProviderName, keyState.ID, callErr)
	}
	return nil
}

// GenerateContent delegates the call to the active provider.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	provider, keyState, err := s.GetActiveProvider(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get active provider: %w", err)
	}

	// Enforce RuntimeConfig checks for the selected key
	if err := s.checkRuntimeConfig(keyState.ID); err != nil {
		return "", err
	}

	content, callErr := provider.GenerateContent(ctx, prompt, imageData, modelType, streamCallback)
	s.sm.UpdateKeyUsage(keyState.ID, func() string {
		if callErr != nil {
			return callErr.Error()
		}
		return ""
	}(), -1)

	if callErr != nil {
		return "", fmt.Errorf("provider '%s' (key: %s) GenerateContent failed: %w", provider.GetModelInfo().ProviderName, keyState.ID, callErr)
	}
	return content, nil
}

// GenerateJSONData delegates the call to the active provider.
func (s *AIService) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	provider, keyState, err := s.GetActiveProvider(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get active provider: %w", err)
	}

	// Enforce RuntimeConfig checks for the selected key
	if err := s.checkRuntimeConfig(keyState.ID); err != nil {
		return "", err
	}

	jsonStr, callErr := provider.GenerateJSONData(ctx, text, language)
	s.sm.UpdateKeyUsage(keyState.ID, func() string {
		if callErr != nil {
			return callErr.Error()
		}
		return ""
	}(), -1)

	if callErr != nil {
		return "", fmt.Errorf("provider '%s' (key: %s) GenerateJSONData failed: %w", provider.GetModelInfo().ProviderName, keyState.ID, callErr)
	}
	return jsonStr, nil
}

// checkRuntimeConfig enforces the rules from the RuntimeConfigManager before allowing an AI call.
func (s *AIService) checkRuntimeConfig(keyID string) error {
	currentConfig := s.sm.GetConfig()

	if !currentConfig.AIEnabled {
		return fmt.Errorf("AI processing is globally disabled by dashboard")
	}

	keyState, ok := currentConfig.APIKeys[keyID]
	if !ok {
		return fmt.Errorf("API key '%s' not found in runtime configuration", keyID)
	}

	// Check provider state for the key's provider
	providerState, ok := currentConfig.Providers[keyState.Provider]
	if !ok {
		return fmt.Errorf("provider '%s' for key '%s' not found in runtime configuration", keyState.Provider, keyID)
	}
	if !providerState.Enabled {
		return fmt.Errorf("AI provider '%s' (for key '%s') is disabled by dashboard", keyState.Provider, keyID)
	}
	if providerState.Paused {
		return fmt.Errorf("AI provider '%s' (for key '%s') is paused by dashboard", keyState.Provider, keyID)
	}
	if providerState.Blocked {
		return fmt.Errorf("AI provider '%s' (for key '%s') is blocked by dashboard: %s", keyState.Provider, keyID, providerState.BlockedReason)
	}

	// Check specific key state
	if !keyState.Enabled {
		return fmt.Errorf("API key '%s' for provider '%s' is disabled by dashboard", keyID, keyState.Provider)
	}
	if keyState.Blocked {
		return fmt.Errorf("API key '%s' for provider '%s' is blocked by dashboard: %s", keyID, keyState.Provider, keyState.BlockedReason)
	}

	// Environment check (TODO: implement this more robustly if different environments have different keys)
	// For now, assume if an API key is selected, it's valid for the active environment.
	// A more robust check might involve tagging keys with environments.

	return nil
}

// selectActiveKeyForProvider selects an active, enabled, unblocked key for a given provider.
// It prioritizes keys that are not blocked by transient errors (like rate limits).
func (s *AIService) selectActiveKeyForProvider(ctx context.Context, providerName string) (AIProvider, st.APIKeyState, error) {
	currentConfig := s.sm.GetConfig()

	// Get all eligible keys for the provider
	var eligibleKeys []st.APIKeyState
	for _, keyState := range currentConfig.APIKeys {
		if keyState.Provider == providerName && keyState.Enabled && !keyState.Blocked {
			eligibleKeys = append(eligibleKeys, keyState)
		}
	}

	if len(eligibleKeys) == 0 {
		return nil, st.APIKeyState{}, fmt.Errorf("no eligible API keys found for provider %s", providerName)
	}

	// Prioritize keys not currently marked with a rate_limit_exceeded error
	for _, keyState := range eligibleKeys {
		if !strings.Contains(keyState.LastError, "rate_limit_exceeded") {
			if provider, ok := s.providers[providerName][keyState.ID]; ok {
				return provider, keyState, nil
			}
		}
	}

	// Fallback to any eligible key if all have rate limit errors or similar
	for _, keyState := range eligibleKeys {
		if provider, ok := s.providers[providerName][keyState.ID]; ok {
			return provider, keyState, nil
		}
	}

	return nil, st.APIKeyState{}, fmt.Errorf("could not select an active provider instance for %s", providerName)
}

// Helper function to get keys from a map.
func getMapKeys(m map[string]AIProvider) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
