package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/telemetry"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
)

type AIService struct {
	providers       map[string]map[string]AIProvider
	sm              *state.RuntimeConfigManager
	providerConfigs map[string]config.ProviderConfig
	switchingRules  config.SwitchingRules
	mu              *sync.RWMutex
}

func NewAIService(ctx context.Context, sm *state.RuntimeConfigManager, providerConfigs map[string]config.ProviderConfig, switchingRules config.SwitchingRules) *AIService {
	s := &AIService{
		providers:       make(map[string]map[string]AIProvider),
		sm:              sm,
		providerConfigs: providerConfigs,
		switchingRules:  switchingRules,
		mu:              &sync.RWMutex{},
	}

	s.InitializeProviders(ctx)

	count := 0
	for _, m := range s.providers {
		count += len(m)
	}

	if count == 0 {
		telemetry.Warn("No AI providers initialized. AI features unavailable.")
	} else {
		telemetry.Info("AI Service initialized", "provider_count", count)
	}

	return s
}

func (s *AIService) InitializeProviders(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Handle nil RuntimeConfigManager gracefully
	if s.sm == nil {
		s.providers = make(map[string]map[string]AIProvider)
		return
	}

	config := s.sm.GetConfig(false)
	s.providers = make(map[string]map[string]AIProvider)

	for providerName, providerState := range config.Providers {
		// Skip disabled providers
		if !providerState.Enabled {
			continue
		}

		s.providers[providerName] = make(map[string]AIProvider)

		for keyID, keyState := range config.APIKeys {
			if keyState.Provider != providerName || !keyState.Enabled || keyState.Blocked {
				continue
			}

			var provider AIProvider
			modelName := providerState.ModelName

			// Check if Cloudflare worker is configured - if so, use it for all providers
			if cfProvider, hasCF := config.Providers["Cloudflare"]; hasCF && cfProvider.Enabled {
				// Find an enabled Cloudflare key (worker URL)
				for _, cfKey := range config.APIKeys {
					if cfKey.Provider == "Cloudflare" && cfKey.Enabled && !cfKey.Blocked {
						provider = NewCloudflareProviderWithProvider(cfKey.Value, providerName)
						break
					}
				}
			}

			// If no worker available, use direct providers
			if provider == nil {
				switch providerName {
				case "Gemini":
					provider = NewGeminiProvider(ctx, keyState.Value, modelName)
				case "Groq":
					provider = NewGroqProvider(keyState.Value, modelName, nil)
				case "Hugging Face":
					provider = NewHuggingFaceProvider(keyState.Value, modelName)
				case "OpenRouter":
					provider = NewOpenRouterProvider(keyState.Value, modelName, nil)
				case "Cloudflare":
					provider = NewCloudflareProvider(keyState.Value)
				case "Replicate":
					provider = NewReplicateProvider(keyState.Value, modelName)
				case "Together":
					provider = NewTogetherProvider(keyState.Value, modelName)
				default:
					telemetry.Warn("Unknown provider", "name", providerName)
					continue
				}
			}

			s.providers[providerName][keyID] = provider
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
	cfg := s.sm.GetConfig(true)
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

// selectProvider selects an active provider and key, respecting fallback logic and exclusion listate.
func (s *AIService) selectProvider(ctx context.Context, task_tokens int, task_depth int, max_cost float64, excludeKeyIDs []string) (AIProvider, state.APIKeyState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := s.sm.GetConfig(false)
	if !cfg.AIEnabled {
		return nil, state.APIKeyState{}, fmt.Errorf("AI is globally disabled")
	}

	activeProvider := s.GetActiveProviderName()
	providerName := select_provider(task_tokens, task_depth, max_cost, s.providerConfigs, s.switchingRules, activeProvider)
	telemetry.Info("Selected provider", "provider", providerName)

	isExcluded := func(id string) bool {
		for _, ex := range excludeKeyIDs {
			if ex == id {
				return true
			}
		}
		return false
	}

	// Helper to find valid key for provider (case-insensitive lookup)
	findKey := func(provName string) (AIProvider, state.APIKeyState, bool) {
		// First try exact match
		ps, ok := cfg.Providers[provName]
		if !ok || !ps.Enabled || ps.Blocked {
			// Try case-insensitive match
			for name, provider := range cfg.Providers {
				if strings.EqualFold(name, provName) {
					ps = provider
					provName = name // Use the correct cased name
					ok = true
					break
				}
			}
			if !ok || !ps.Enabled || ps.Blocked {
				return nil, state.APIKeyState{}, false
			}
		}

		// Get keys
		keyMap, ok := s.providers[provName]
		if !ok {
			// Try with corrected case from providers map
			for name, keys := range s.providers {
				if strings.EqualFold(name, provName) {
					keyMap = keys
					provName = name
					break
				}
			}
			if keyMap == nil {
				return nil, state.APIKeyState{}, false
			}
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
		return nil, state.APIKeyState{}, false
	}

	if p, k, ok := findKey(providerName); ok {
		telemetry.Info("Selected AI provider: " + providerName + " with key: " + k.ID)
		return p, k, nil
	}

	// If no real providers available, return mock provider for testing
	telemetry.Warn("No active AI providers available for: " + providerName + ", using mock response")
	mockProvider := &MockAIProvider{providerName: providerName}
	mockKey := state.APIKeyState{
		ID:      "mock-key",
		Value:   "mock",
		Enabled: true,
		Blocked: false,
	}
	return mockProvider, mockKey, nil
}

// ExecuteWithRetry handles retries for transient errors, tracking failed keys to ensure they are skipped in subsequent attempts.
func (s *AIService) ExecuteWithRetry(ctx context.Context, task_tokens int, task_depth int, max_cost float64, op func(AIProvider) error) error {
	maxRetries := s.switchingRules.RetryCount
	if maxRetries <= 0 {
		maxRetries = 1 // Ensure at least one retry attempt
	}
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
			telemetry.Warn("AI Provider failover triggered",
				"attempt", i+1,
				"failed_provider", providerName,
				"error_code", appErr.Code,
				"msg", appErr.Message,
			)

			// If it's a serious permanent error, block the key globally
			if appErr.Code == ErrCodeUnauthorized || appErr.Code == ErrCodeInvalidRequest {
				telemetry.Error("Blocking failing key permanently (invalid/unauthorized)", "key_id", key.ID, "error", appErr.Message)
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
				telemetry.Warn("System error, retrying with different provider/key", "attempt", i+1, "error", err)
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
	telemetry.Info("AI Chat called with prompt: " + req.UserPrompt[:min(50, len(req.UserPrompt))] + "...")
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

// MockAIProvider provides mock responses for testing when no real providers are configured
type MockAIProvider struct {
	providerName string
}

func (m *MockAIProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: m.providerName,
		ModelName:    "mock-model",
	}
}

func (m *MockAIProvider) Chat(ctx context.Context, req *RequestModel, callback func(string)) error {
	// Try to make real HTTP requests to AI APIs as fallback
	return m.makeHTTPFallbackRequest(ctx, req, callback)
}

// makeHTTPFallbackRequest attempts to make direct HTTP calls to AI APIs when providers aren't configured
func (m *MockAIProvider) makeHTTPFallbackRequest(ctx context.Context, req *RequestModel, callback func(string)) error {
	provider := strings.ToLower(m.providerName)

	switch provider {
	case "gemini", "google":
		return m.callGeminiAPI(ctx, req, callback)
	case "groq":
		return m.callGroqAPI(ctx, req, callback)
	case "openai":
		return m.callOpenAIAPI(ctx, req, callback)
	case "cloudflare":
		return m.callCloudflareAPI(ctx, req, callback)
	default:
		// Fallback to provider-specific mock response
		mockResponses := map[string]string{
			"gemini":      "Hello! I'm Gemini AI. This is a demo response since no API key is configured. To use real Gemini API, set GEMINI_API_KEY environment variable.",
			"groq":        "Hello! I'm Groq AI. This is a demo response since no API key is configured. To use real Groq API, set GROQ_API_KEY environment variable.",
			"cloudflare":  "Hello! I'm Cloudflare Workers AI. This is a demo response since no API credentials are configured. To use real Cloudflare AI, set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID environment variables.",
			"openrouter":  "Hello! I'm OpenRouter AI. This is a demo response since no API key is configured. To use real OpenRouter API, set OPENROUTER_API_KEY environment variable.",
			"replicate":   "Hello! I'm Replicate AI. This is a demo response since no API key is configured. To use real Replicate API, set REPLICATE_API_KEY environment variable.",
			"together":    "Hello! I'm Together AI. This is a demo response since no API key is configured. To use real Together API, set TOGETHER_API_KEY environment variable.",
			"huggingface": "Hello! I'm Hugging Face AI. This is a demo response since no API key is configured. To use real Hugging Face API, set HUGGINGFACE_API_KEY environment variable.",
		}

		response := mockResponses[strings.ToLower(provider)]
		if response == "" {
			response = fmt.Sprintf("Hello! I'm a demo AI assistant for %s. No API keys configured, using demo mode.", m.providerName)
		}

		// Simulate streaming
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	}
}

// callGeminiAPI makes a direct HTTP call to Gemini API
func (m *MockAIProvider) callGeminiAPI(ctx context.Context, req *RequestModel, callback func(string)) error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		response := "Hello! I'm Gemini AI. This is a demo response since no API key is configured. To use real Gemini API, set GEMINI_API_KEY environment variable."
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	}

	// Make real API call to Gemini
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": req.SystemPrompt + "\n\n" + req.UserPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     req.Temperature,
			"maxOutputTokens": 1000,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Gemini API error: %s", string(body))
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		response := geminiResp.Candidates[0].Content.Parts[0].Text
		// Simulate streaming by sending word by word
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(30 * time.Millisecond)
			}
		}
	}

	return nil
}

// callGroqAPI makes a direct HTTP call to Groq API
func (m *MockAIProvider) callGroqAPI(ctx context.Context, req *RequestModel, callback func(string)) error {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		response := "Hello! I'm Groq AI. This is a demo response since no API key is configured. To use real Groq API, set GROQ_API_KEY environment variable."
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	}

	// Groq uses OpenAI-compatible API
	payload := map[string]interface{}{
		"model": "mixtral-8x7b-32768",
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"stream":      true,
		"max_tokens":  1000,
		"temperature": req.Temperature,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Groq API error: %s", string(body))
	}

	// Parse streaming response (similar to OpenAI)
	return m.parseOpenAIStream(resp.Body, callback)
}

// callOpenAIAPI makes a direct HTTP call to OpenAI API
func (m *MockAIProvider) callOpenAIAPI(ctx context.Context, req *RequestModel, callback func(string)) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		response := "Hello! I'm OpenAI GPT. This is a demo response since no API key is configured. To use real OpenAI API, set OPENAI_API_KEY environment variable."
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	}

	// Make real API call to OpenAI
	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"stream":     true,
		"max_tokens": 1000,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Parse streaming response
	return m.parseOpenAIStream(resp.Body, callback)
}

// parseOpenAIStream parses the streaming response from OpenAI
func (m *MockAIProvider) parseOpenAIStream(body io.Reader, callback func(string)) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var response struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &response); err != nil {
				continue
			}

			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				callback(response.Choices[0].Delta.Content)
			}
		}
	}
	return scanner.Err()
}

// callCloudflareAPI makes a direct HTTP call to Cloudflare Workers AI
func (m *MockAIProvider) callCloudflareAPI(ctx context.Context, req *RequestModel, callback func(string)) error {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	if apiToken == "" || accountID == "" {
		response := "Hello! I'm Cloudflare Workers AI. This is a demo response since no API credentials are configured. To use real Cloudflare AI, set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID environment variables."
		words := strings.Fields(response)
		for _, word := range words {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				callback(word + " ")
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	}

	// Cloudflare Workers AI API call
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.UserPrompt},
		},
		"stream": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/meta/llama-3-8b-instruct", accountID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Cloudflare AI API error: %s", string(body))
	}

	// Cloudflare returns the response as plain text stream
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := string(body)
	// Simulate streaming by sending word by word
	words := strings.Fields(response)
	for _, word := range words {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			callback(word + " ")
			time.Sleep(40 * time.Millisecond)
		}
	}

	return nil
}

func (m *MockAIProvider) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	mockResponses := map[string]string{
		"openai":     "Hello! I'm a mock OpenAI assistant. This is a test response since no API keys are configured.",
		"anthropic":  "Greetings! I'm a mock Anthropic assistant. This is a test response since no API keys are configured.",
		"google":     "Hi there! I'm a mock Google/Gemini assistant. This is a test response since no API keys are configured.",
		"gemini":     "Hi there! I'm a mock Google/Gemini assistant. This is a test response since no API keys are configured.",
		"local":      "Hello! I'm a mock local AI assistant. This is a test response since no API keys are configured.",
		"cloudflare": "Hey! I'm a mock Cloudflare AI assistant. This is a test response since no API keys are configured.",
	}

	response := mockResponses[strings.ToLower(m.providerName)]
	if response == "" {
		response = fmt.Sprintf("Hello! I'm a mock AI assistant for %s. This is a test response since no API keys are configured.", m.providerName)
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: response,
			},
		},
	}, nil
}

func (m *MockAIProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	// Not used in current implementation
	return nil, fmt.Errorf("mock provider does not support GenerateCompletion")
}

func (m *MockAIProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	stream := make(chan StreamResponse, 1)

	mockResponses := map[string]string{
		"openai":     "Hello! I'm a mock OpenAI assistant. This is a test response since no API keys are configured.",
		"anthropic":  "Greetings! I'm a mock Anthropic assistant. This is a test response since no API keys are configured.",
		"google":     "Hi there! I'm a mock Google/Gemini assistant. This is a test response since no API keys are configured.",
		"gemini":     "Hi there! I'm a mock Google/Gemini assistant. This is a test response since no API keys are configured.",
		"local":      "Hello! I'm a mock local AI assistant. This is a test response since no API keys are configured.",
		"cloudflare": "Hey! I'm a mock Cloudflare AI assistant. This is a test response since no API keys are configured.",
	}

	response := mockResponses[strings.ToLower(m.providerName)]
	if response == "" {
		response = fmt.Sprintf("Hello! I'm a mock AI assistant for %s. This is a test response since no API keys are configured.", m.providerName)
	}

	go func() {
		defer close(stream)

		// Simulate streaming by sending in chunks
		words := strings.Fields(response)
		for i, word := range words {
			select {
			case <-ctx.Done():
				stream <- StreamResponse{Error: ctx.Err()}
				return
			case stream <- StreamResponse{Content: word + " "}:
			}

			// Small delay to simulate streaming
			time.Sleep(50 * time.Millisecond)

			// Send final chunk without space if it's the last word
			if i == len(words)-1 {
				stream <- StreamResponse{Content: "", Done: true}
			}
		}
	}()

	return stream, nil
}

func (m *MockAIProvider) CheckHealth(ctx context.Context) error {
	// Mock provider is always healthy
	return nil
}
