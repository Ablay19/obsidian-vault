package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	st "obsidian-automation/internal/state"
)

// AIService manages multiple AI providers and selects the active one.
type AIService struct {
	providers map[string]map[string]AIProvider
	sm        *st.RuntimeConfigManager
	mu        sync.RWMutex
}

// NewAIService initializes the AI service using the RuntimeConfigManager.
func NewAIService(ctx context.Context, sm *st.RuntimeConfigManager) *AIService {
	s := &AIService{
		providers: make(map[string]map[string]AIProvider),
		sm:        sm,
	}
	s.RefreshProviders(ctx)
	
	// Quick check if any providers loaded
	count := 0
	for _, m := range s.providers {
		count += len(m)
	}
	if count == 0 {
		slog.Warn("No AI providers initialized. AI features unavailable.")
	} else {
		slog.Info("AI Service initialized", "provider_count", count)
	}
	
	return s
}

// RefreshProviders populates the providers map based on the current RuntimeConfig.
func (s *AIService) RefreshProviders(ctx context.Context) {
	config := s.sm.GetConfig()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.providers = make(map[string]map[string]AIProvider)

	for providerName, providerState := range config.Providers {
		if !providerState.Enabled {
			continue
		}

		s.providers[providerName] = make(map[string]AIProvider)
		for keyID, keyState := range config.APIKeys {
			if keyState.Provider == providerName && keyState.Enabled && !keyState.Blocked {
				if keyState.Value == "" {
					continue
				}

				var provider AIProvider
				modelName := providerState.ModelName

				switch providerName {
				case "Gemini":
					provider = NewGeminiProvider(ctx, keyState.Value, modelName)
				case "Groq":
					provider = NewGroqProvider(keyState.Value, modelName)
				case "Hugging Face":
					provider = NewHuggingFaceProvider(keyState.Value, modelName)
				case "OpenRouter":
					provider = NewOpenRouterProvider(keyState.Value, modelName)
				case "onnx", "None", "ONNX":
					// Known but unimplemented or handled elsewhere
					continue
				default:
					slog.Warn("Unknown provider", "name", providerName)
					continue
				}

				if provider != nil {
					s.providers[providerName][keyID] = provider
				}
			}
		}
	}
}

// SetProvider changes the active AI provider preference.
func (s *AIService) SetProvider(providerName string) error {
	return s.sm.SetActiveProvider(providerName)
}

// GetActiveProviderName returns the name of the currently active provider.
func (s *AIService) GetActiveProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.sm.GetConfig()
	if cfg.ActiveProvider != "" {
		return cfg.ActiveProvider
	}
	// Fallback
	for name, ps := range cfg.Providers {
		if ps.Enabled {
			return name
		}
	}
	return "None"
}

// GetAvailableProviders returns a list of available provider names.
func (s *AIService) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.providers))
	for k := range s.providers {
		if len(s.providers[k]) > 0 {
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
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, keyMap := range s.providers {
		for _, p := range keyMap {
			wg.Add(1)
			go func(n string, prov AIProvider) {
				defer wg.Done()
				tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := prov.CheckHealth(tCtx); err == nil {
					mu.Lock()
					found := false
					for _, h := range healthy {
						if h == n {
							found = true
							break
						}
					}
					if !found {
						healthy = append(healthy, n)
					}
					mu.Unlock()
				}
			}(name, p)
			break // Check one key per provider is enough for general "provider health" usually
		}
	}
	wg.Wait()
	return healthy
}

// GetProvidersInfo returns model info.
func (s *AIService) GetProvidersInfo() []ModelInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var infos []ModelInfo
	for _, keyMap := range s.providers {
		for id, p := range keyMap {
			info := p.GetModelInfo()
			info.KeyID = id
			infos = append(infos, info)
		}
	}
	return infos
}

// === Core Logic ===

// selectProvider selects an active provider and key, respecting fallback logic.
func (s *AIService) selectProvider(ctx context.Context) (AIProvider, st.APIKeyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	cfg := s.sm.GetConfig()
	if !cfg.AIEnabled {
		return nil, st.APIKeyState{}, fmt.Errorf("AI is globally disabled")
	}

	preferred := cfg.ActiveProvider
	
	// Helper to find valid key for provider
	findKey := func(provName string) (AIProvider, st.APIKeyState, bool) {
		ps, ok := cfg.Providers[provName]
		if !ok || !ps.Enabled || ps.Blocked {
			return nil, st.APIKeyState{}, false
		}
		
		// Get keys
		keyMap, ok := s.providers[provName]
		if !ok {
			return nil, st.APIKeyState{}, false
		}

		for id, p := range keyMap {
			ks, ok := cfg.APIKeys[id]
			if !ok || !ks.Enabled || ks.Blocked {
				continue
			}
			// Avoid keys recently rate limited? (Simplification: relying on internal state/retry logic mostly)
			return p, ks, true
		}
		return nil, st.APIKeyState{}, false
	}

	// 1. Try preferred
	if preferred != "" {
		if p, k, ok := findKey(preferred); ok {
			return p, k, nil
		}
	}

	// 2. Fallback
	for name := range s.providers {
		if name == preferred {
			continue
		}
		if p, k, ok := findKey(name); ok {
			return p, k, nil
		}
	}

	return nil, st.APIKeyState{}, fmt.Errorf("no active AI providers available")
}

// ExecuteWithRetry handles retries for transient errors.
func (s *AIService) ExecuteWithRetry(ctx context.Context, op func(AIProvider) error) error {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		provider, key, err := s.selectProvider(ctx)
		if err != nil {
			return err
		}

		err = op(provider)
		if err == nil {
			return nil
		}

		// Check if error is an AppError
		if appErr, ok := err.(*AppError); ok {
			// Always block the specific key that failed for this session if it's a serious error
			if appErr.Code == ErrCodeRateLimit || appErr.Code == ErrCodeUnauthorized || appErr.Code == ErrCodeInvalidRequest {
				slog.Warn("Blocking failing key", "key_id", key.ID, "code", appErr.Code)
				s.sm.UpdateKeyUsage(key.ID, appErr.Message, -1)
			}

			if appErr.Retry && i < maxRetries-1 {
				slog.Warn("Transient error, retrying with next available provider/key", "attempt", i+1, "error", err)
				time.Sleep(backoff)
				backoff = time.Duration(math.Min(float64(backoff)*2, float64(30*time.Second)))
				continue
			}
		}

		return err // Non-retryable or max retries reached
	}

	return fmt.Errorf("max retries exceeded")
}

// AnalyzeText generates structured JSON data from text.
func (s *AIService) AnalyzeText(ctx context.Context, text, language string) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object.
Target JSON Structure:
{
  "category": "One of [physics, math, chemistry, admin, general]",
  "topics": ["topic1", "topic2", "topic3"],
  "questions": ["question1", "question2"]
}
Ensure "topics" and "questions" are in %s.
Text:
%s`, language, text)

	req := &RequestModel{
		UserPrompt: prompt,
		JSONMode:   true,
		Temperature: 0.2,
	}

	var resp *ResponseModel
	err := s.ExecuteWithRetry(ctx, func(p AIProvider) error {
		var e error
		resp, e = p.GenerateCompletion(ctx, req)
		return e
	})

	if err != nil {
		return nil, err
	}

	// Clean JSON string if necessary (providers might wrap in markdown blocks despite instructions)
	cleanJSON := strings.TrimSpace(resp.Content)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &result, nil
}

// Chat streams the response for a conversation.
func (s *AIService) Chat(ctx context.Context, req *RequestModel, callback func(string)) error {
	return s.ExecuteWithRetry(ctx, func(p AIProvider) error {
		stream, err := p.StreamCompletion(ctx, req)
		if err != nil {
			return err
		}

		for chunk := range stream {
			if chunk.Error != nil {
				return chunk.Error
			}
			if chunk.Content != "" {
				callback(chunk.Content)
			}
		}
		return nil
	})
}
