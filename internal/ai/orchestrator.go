package ai

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"obsidian-automation/pkg/utils"
)

// TelegramAIProvider wraps existing AI provider interface for Telegram bot
type TelegramAIProvider struct {
	AIProvider
	name      string
	model     string
	available bool
}

// GenerationOptions contains options for AI generation
type GenerationOptions struct {
	Model          string
	MaxTokens      int
	Temperature    float64
	TopP           float64
	Stream         bool
	SystemPrompt   string
	ConversationID int64
}

// GenerationResult contains the result of AI generation
type GenerationResult struct {
	Content      string
	Model        string
	Provider     string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Latency      time.Duration
	FinishReason string
}

// Orchestrator manages AI providers and model selection
type Orchestrator struct {
	providers       map[string]AIProvider
	localModels     map[string]AIProvider
	apiProviders    map[string]AIProvider
	logger          *utils.Logger
	fallbackChain   []string
	currentProvider string
}

// NewOrchestrator creates a new AI orchestrator
func NewOrchestrator(logger *utils.Logger) *Orchestrator {
	return &Orchestrator{
		providers:     make(map[string]AIProvider),
		localModels:   make(map[string]AIProvider),
		apiProviders:  make(map[string]AIProvider),
		logger:        logger,
		fallbackChain: []string{"local", "huggingface", "replicate", "openrouter"},
	}
}

// AddProvider adds an AI provider to orchestrator
func (o *Orchestrator) AddProvider(name string, provider AIProvider) {
	o.providers[name] = provider

	// Categorize provider
	if isLocalProvider(name) {
		o.localModels[name] = provider
	} else {
		o.apiProviders[name] = provider
	}

	modelInfo := provider.GetModelInfo()
	o.logger.Info("Added AI provider", "name", name, "model", modelInfo.Name)
}

// SetDefaultProvider sets the default provider
func (o *Orchestrator) SetDefaultProvider(providerName string) error {
	if _, exists := o.providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	o.currentProvider = providerName
	o.logger.Info("Set default AI provider", "provider", providerName)
	return nil
}

// Generate generates text using the best available provider
func (o *Orchestrator) Generate(ctx context.Context, prompt string, options *GenerationOptions) (*GenerationResult, error) {
	start := time.Now()

	// Convert to RequestModel for existing interface
	req := &RequestModel{
		Prompt:       prompt,
		Model:        options.Model,
		MaxTokens:    options.MaxTokens,
		Temperature:  options.Temperature,
		Stream:       options.Stream,
		SystemPrompt: options.SystemPrompt,
	}

	// Try current provider first
	if o.currentProvider != "" {
		if result, err := o.tryProvider(ctx, o.currentProvider, req); err == nil {
			return o.convertResponse(result, start)
		} else {
			o.logger.Warn("Provider failed, trying fallback", "provider", o.currentProvider, "error", err)
		}
	}

	// Try fallback chain
	for _, providerName := range o.fallbackChain {
		if result, err := o.tryProvider(ctx, providerName, req); err == nil {
			// Update current provider for future requests
			o.currentProvider = providerName
			return o.convertResponse(result, start)
		} else {
			o.logger.Warn("Fallback provider failed", "provider", providerName, "error", err)
		}
	}

	return nil, fmt.Errorf("all AI providers failed")
}

// tryProvider attempts to generate using a specific provider
func (o *Orchestrator) tryProvider(ctx context.Context, providerName string, req *RequestModel) (*ResponseModel, error) {
	provider, exists := o.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", providerName)
	}

	if err := provider.CheckHealth(ctx); err != nil {
		return nil, fmt.Errorf("provider '%s' is not available: %v", providerName, err)
	}

	o.logger.AIRequest(providerName, req.Model, len(req.Prompt), 0) // TODO: Get user ID

	result, err := provider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// convertResponse converts ResponseModel to GenerationResult
func (o *Orchestrator) convertResponse(resp *ResponseModel, start time.Time) (*GenerationResult, error) {
	if resp.Error != nil {
		return nil, resp.Error
	}

	latency := time.Since(start)
	modelInfo := resp.ModelInfo

	o.logger.AIResponse("unknown", modelInfo.Name, resp.OutputTokens, latency.Milliseconds(), 0) // TODO: Get user ID

	return &GenerationResult{
		Content:      resp.Content,
		Model:        modelInfo.Name,
		Provider:     "unknown", // TODO: Track provider
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
		Cost:         0, // TODO: Calculate cost
		Latency:      latency,
		FinishReason: resp.FinishReason,
	}, nil
}

// GetAvailableProviders returns list of available providers
func (o *Orchestrator) GetAvailableProviders() []string {
	var available []string
	for name, provider := range o.providers {
		ctx := context.Background()
		if err := provider.CheckHealth(ctx); err == nil {
			available = append(available, name)
		}
	}
	return available
}

// GetProviderInfo returns information about a provider
func (o *Orchestrator) GetProviderInfo(name string) map[string]interface{} {
	provider, exists := o.providers[name]
	if !exists {
		return nil
	}

	modelInfo := provider.GetModelInfo()
	ctx := context.Background()
	isAvailable := provider.CheckHealth(ctx) == nil

	return map[string]interface{}{
		"name":      name,
		"model":     modelInfo.Name,
		"available": isAvailable,
	}
}

// GetStats returns orchestration statistics
func (o *Orchestrator) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_providers":     len(o.providers),
		"local_providers":     len(o.localModels),
		"api_providers":       len(o.apiProviders),
		"current_provider":    o.currentProvider,
		"available_providers": o.GetAvailableProviders(),
	}

	return stats
}

// NewOrchestrator creates a new AI orchestrator
func NewOrchestrator(logger *utils.Logger) *Orchestrator {
	return &Orchestrator{
		providers:     make(map[string]AIProvider),
		localModels:   make(map[string]AIProvider),
		apiProviders:  make(map[string]AIProvider),
		logger:        logger,
		fallbackChain: []string{"local", "huggingface", "replicate", "openrouter"},
	}
}

// AddProvider adds an AI provider to the orchestrator
func (o *Orchestrator) AddProvider(name string, provider AIProvider) {
	o.providers[name] = provider

	// Categorize provider
	if isLocalProvider(name) {
		o.localModels[name] = provider
	} else {
		o.apiProviders[name] = provider
	}

	o.logger.Info("Added AI provider", "name", name, "model", provider.GetModel())
}

// SetDefaultProvider sets the default provider
func (o *Orchestrator) SetDefaultProvider(providerName string) error {
	if _, exists := o.providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	o.currentProvider = providerName
	o.logger.Info("Set default AI provider", "provider", providerName)
	return nil
}

// Generate generates text using the best available provider
func (o *Orchestrator) Generate(ctx context.Context, prompt string, options *GenerationOptions) (*GenerationResult, error) {
	start := time.Now()

	// Try current provider first
	if o.currentProvider != "" {
		if result, err := o.tryProvider(ctx, o.currentProvider, prompt, options); err == nil {
			return result, nil
		} else {
			o.logger.Warn("Provider failed, trying fallback", "provider", o.currentProvider, "error", err)
		}
	}

	// Try fallback chain
	for _, providerName := range o.fallbackChain {
		if result, err := o.tryProvider(ctx, providerName, prompt, options); err == nil {
			// Update current provider for future requests
			o.currentProvider = providerName
			return result, nil
		} else {
			o.logger.Warn("Fallback provider failed", "provider", providerName, "error", err)
		}
	}

	return nil, fmt.Errorf("all AI providers failed")
}

// tryProvider attempts to generate using a specific provider
func (o *Orchestrator) tryProvider(ctx context.Context, providerName string, prompt string, options *GenerationOptions) (*GenerationResult, error) {
	provider, exists := o.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", providerName)
	}

	if !provider.IsAvailable() {
		return nil, fmt.Errorf("provider '%s' is not available", providerName)
	}

	o.logger.AIRequest(provider.GetName(), provider.GetModel(), len(prompt), 0) // TODO: Get user ID

	result, err := provider.Generate(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	latency := time.Since(start)
	o.logger.AIResponse(provider.GetName(), provider.GetModel(), result.OutputTokens, latency.Milliseconds(), 0) // TODO: Get user ID

	return result, nil
}

// GetAvailableProviders returns list of available providers
func (o *Orchestrator) GetAvailableProviders() []string {
	var available []string
	for name, provider := range o.providers {
		if provider.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}

// GetProviderInfo returns information about a provider
func (o *Orchestrator) GetProviderInfo(name string) map[string]interface{} {
	provider, exists := o.providers[name]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"name":      provider.GetName(),
		"model":     provider.GetModel(),
		"available": provider.IsAvailable(),
	}
}

// GetStats returns orchestration statistics
func (o *Orchestrator) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_providers":     len(o.providers),
		"local_providers":     len(o.localModels),
		"api_providers":       len(o.apiProviders),
		"current_provider":    o.currentProvider,
		"available_providers": o.GetAvailableProviders(),
	}

	return stats
}

// isLocalProvider checks if a provider is local
func isLocalProvider(name string) bool {
	localProviders := []string{"local", "gpt2", "gpt-neo", "codellama", "starcoder"}
	for _, local := range localProviders {
		if name == local {
			return true
		}
	}
	return false
}
