package ai

import (
	"context"
	"fmt"
)

// Orchestrator manages AI providers and local models
type Orchestrator struct {
	providers   map[string]AIProvider
	localModels map[string]ModelInfo
	logger      Logger
}

// AIProvider interface for AI services
type AIProvider interface {
	GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error)
	StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error)
	CheckHealth(ctx context.Context) error
	GetModelInfo() ModelInfo
}

// Logger interface
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewOrchestrator creates a new AI orchestrator
func NewOrchestrator(logger Logger) *Orchestrator {
	return &Orchestrator{
		providers:   make(map[string]AIProvider),
		localModels: make(map[string]ModelInfo),
		logger:      logger,
	}
}

// AddProvider adds an AI provider
func (o *Orchestrator) AddProvider(name string, provider AIProvider) {
	o.providers[name] = provider
}

// GetProvider gets an AI provider by name
func (o *Orchestrator) GetProvider(name string) (AIProvider, error) {
	provider, ok := o.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}
