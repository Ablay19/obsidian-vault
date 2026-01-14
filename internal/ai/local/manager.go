package local

import (
	"context"
	"fmt"
	"time"

	"obsidian-automation/internal/ai"
	"obsidian-automation/pkg/utils"
)

// ModelManager manages local AI models
type ModelManager struct {
	models map[string]ai.AIProvider
	logger *utils.Logger
}

// NewModelManager creates a new model manager
func NewModelManager(logger *utils.Logger) *ModelManager {
	return &ModelManager{
		models: make(map[string]ai.AIProvider),
		logger: logger,
	}
}

// AddModel adds a local model to the manager
func (mm *ModelManager) AddModel(name string, model ai.AIProvider) {
	mm.models[name] = model
	mm.logger.Info("Added local AI model", "name", name, "model", "local")
}

// GetBestModel returns the best available local model
func (mm *ModelManager) GetBestModel() (string, ai.AIProvider) {
	// For now, return first available model
	for name, model := range mm.models {
		if mm.isModelAvailable(model) {
			return name, model
		}
	}

	// Default fallback
	return "", nil
}

// isModelAvailable checks if a model is available
func (mm *ModelManager) isModelAvailable(model ai.AIProvider) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple health check - in a real implementation,
	// this would check model file existence and basic functionality
	err := model.CheckHealth(ctx)
	return err == nil
}

// GetModel returns a specific model by name
func (mm *ModelManager) GetModel(name string) (ai.AIProvider, error) {
	model, exists := mm.models[name]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found", name)
	}

	return model, nil
}

// ListModels returns all available models
func (mm *ModelManager) ListModels() []string {
	var available []string
	for name := range mm.models {
		available = append(available, name)
	}
	return available
}

// GetStats returns model manager statistics
func (mm *ModelManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_models":     len(mm.models),
		"available_models": mm.ListModels(),
	}
}

// Close releases resources
func (mm *ModelManager) Close() error {
	// Clean up resources if needed
	mm.logger.Info("Model manager closed")
	return nil
}
