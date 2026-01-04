package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"obsidian-automation/internal/config"
	st "obsidian-automation/internal/state"
)

type AIService struct {
	providers       map[string]map[string]AIProvider
	sm              *st.RuntimeConfigManager
	providerConfigs map[string]config.ProviderConfig
	switchingRules  config.SwitchingRules
	mu              sync.RWMutex
}

func NewAIService(ctx context.Context, sm *st.RuntimeConfigManager, providerConfigs map[string]config.ProviderConfig, switchingRules config.SwitchingRules) *AIService {
	s := &AIService{
		providers:       make(map[string]map[string]AIProvider),
		sm:              sm,
		providerConfigs: providerConfigs,
		switchingRules:  switchingRules,
	}

	s.InitializeProviders(ctx)

	count := 0
	for _, m := range s.providers {
		count += len(m)
	}

	if count == 0 {
		zap.S().Warn("No AI providers initialized. AI features unavailable.")
	} else {
		zap.S().Info("AI Service initialized", "provider_count", count)
	}

	return s
}

func (s *AIService) InitializeProviders(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.sm.GetConfig()
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
					provider = NewGroqProvider(keyState.Value, modelName, nil)
				case "Hugging Face":
					provider = NewHuggingFaceProvider(keyState.Value, modelName)
				case "OpenRouter":
					provider = NewOpenRouterProvider(keyState.Value, modelName, nil)
				default:
					zap.S().Warn("Unknown provider", "name", providerName)
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
	if cfg.ActiveProvider != "" && cfg.ActiveProvider != "None" {
		return cfg.ActiveProvider
	}

	// Fallback to first available
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

// selectProvider selects an active provider and key, respecting fallback logic and exclusion list.
func (s *AIService) selectProvider(ctx context.Context, task_tokens int, task_depth int, max_cost float64, excludeKeyIDs []string) (AIProvider, st.APIKeyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := s.sm.GetConfig()
	if !cfg.AIEnabled {
		return nil, st.APIKeyState{}, fmt.Errorf("AI is globally disabled")
	}

	providerName := select_provider(task_tokens, task_depth, max_cost, s.providerConfigs, s.switchingRules)
	zap.S().Info("Selected provider", "provider", providerName)

	isExcluded := func(id string) bool {
		for _, ex := range excludeKeyIDs {
			if ex == id {
				return true
			}
		}
		return false
	}

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
			if isExcluded(id) {
				continue
			}
			ks, ok := cfg.APIKeys[id]
			if !ok || !ks.Enabled || ks.Blocked {
				continue
			}
			return p, ks, true
		}
		return nil, st.APIKeyState{}, false
	}

	if p, k, ok := findKey(providerName); ok {
		zap.S().Debug("Selected AI provider", "provider", providerName, "key_id", k.ID)
		return p, k, nil
	}

	return nil, st.APIKeyState{}, fmt.Errorf("no active AI providers available for selected provider %s", providerName)
}

// ExecuteWithRetry handles retries for transient errors, tracking failed keys to ensure they are skipped in subsequent attempts.
func (s *AIService) ExecuteWithRetry(ctx context.Context, task_tokens int, task_depth int, max_cost float64, op func(AIProvider) error) error {
	maxRetries := s.switchingRules.RetryCount
	backoff := time.Duration(s.switchingRules.RetryDelayMs) * time.Millisecond
	var failedKeys []string
	var triedProviders []string

	for i := 0; i < maxRetries; i++ {
		provider, key, err := s.selectProvider(ctx, task_tokens, task_depth, max_cost, failedKeys)
		if err != nil {
			return err
		}

		err = op(provider)
		if err == nil {
			return nil
		}

		// Check if error is an AppError
		if appErr, ok := err.(*AppError); ok {
			// Track this key as failed for the current request context
			failedKeys = append(failedKeys, key.ID)

			providerName := provider.GetModelInfo().ProviderName
			triedProviders = append(triedProviders, providerName)

			// Log failover event
			zap.S().Warn("AI Provider failover triggered",
				"attempt", i+1,
				"failed_provider", providerName,
				"error_code", appErr.Code,
				"msg", appErr.Message,
			)

			// If it's a serious permanent error, block the key globally
			if appErr.Code == ErrCodeUnauthorized || appErr.Code == ErrCodeInvalidRequest {
				zap.S().Error("Blocking failing key permanently (invalid/unauthorized)", "key_id", key.ID, "error", appErr.Message)
				s.sm.UpdateKeyUsage(key.ID, appErr.Message, -1)
			}

			if appErr.Retry && i < maxRetries-1 {
				time.Sleep(backoff)
				backoff = time.Duration(math.Min(float64(backoff)*2, float64(10*time.Second)))
				continue
			}
		} else {
			failedKeys = append(failedKeys, key.ID)
			if i < maxRetries-1 {
				zap.S().Warn("System error, retrying with different provider/key", "attempt", i+1, "error", err)
				continue
			}
		}

		return err // Non-retryable or max retries reached
	}

	return fmt.Errorf("max retries exceeded (tried providers: %v)", triedProviders)
}

// AnalyzeText generates structured JSON data from text.
func (s *AIService) AnalyzeText(ctx context.Context, text, language string) (*AnalysisResult, error) {
	return s.AnalyzeTextWithParams(ctx, text, language, len(text), 1, 0.01)
}

// AnalyzeTextWithParams generates structured JSON data from text with additional parameters.
func (s *AIService) AnalyzeTextWithParams(ctx context.Context, text, language string, task_tokens int, task_depth int, max_cost float64) (*AnalysisResult, error) {
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
		UserPrompt:  prompt,
		JSONMode:    true,
		Temperature: 0.2,
	}

	var resp *ResponseModel
	err := s.ExecuteWithRetry(ctx, task_tokens, task_depth, max_cost, func(p AIProvider) error {
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
	task_tokens := len(req.UserPrompt)
	task_depth := 1  // Simple chat is depth 1
	max_cost := 0.01 // Default max cost for a chat

	return s.ExecuteWithRetry(ctx, task_tokens, task_depth, max_cost, func(p AIProvider) error {
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
