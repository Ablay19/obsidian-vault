package ai

import (
	"context"
	"fmt"
	"obsidian-automation/internal/state"
	"testing"
	"time"
)

func TestAIService_GetActiveProviderName(t *testing.T) {
	tests := []struct {
		name     string
		config   *state.RuntimeConfig
		expected string
	}{
		{
			name: "Active provider set",
			config: &state.RuntimeConfig{
				ActiveProvider: "gemini",
				Providers: map[string]state.ProviderState{
					"gemini": {Enabled: true},
				},
			},
			expected: "gemini",
		},
		{
			name: "No active provider, fallback to first enabled",
			config: &state.RuntimeConfig{
				ActiveProvider: "None",
				Providers: map[string]state.ProviderState{
					"gemini": {Enabled: true},
					"groq":   {Enabled: true},
				},
			},
			expected: "gemini",
		},
		{
			name: "No providers enabled",
			config: &state.RuntimeConfig{
				ActiveProvider: "None",
				Providers: map[string]state.ProviderState{
					"gemini": {Enabled: false},
					"groq":   {Enabled: false},
				},
			},
			expected: "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRCM := &MockRuntimeConfigManager{config: tt.config}
			aiService := NewAIService(context.Background(), mockRCM, map[string]config.ProviderConfig{}, config.SwitchingRules{})

			result := aiService.GetActiveProviderName()
			if result != tt.expected {
				t.Errorf("GetActiveProviderName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAIService_GetAvailableProviders(t *testing.T) {
	mockRCM := &MockRuntimeConfigManager{
		config: &state.RuntimeConfig{
			Providers: map[string]state.ProviderState{
				"gemini":     {Enabled: true},
				"groq":       {Enabled: true},
				"openrouter": {Enabled: false},
			},
		},
	}

	// Simulate initialized providers
	aiService := &AIService{
		providers: map[string]map[string]AIProvider{
			"gemini": {
				"gemini-pro": &MockAIProvider{},
			},
			"groq": {
				"llama-3": &MockAIProvider{},
			},
		},
		sm: mockRCM,
	}

	available := aiService.GetAvailableProviders()

	expectedProviders := []string{"gemini", "groq"}

	if len(available) != len(expectedProviders) {
		t.Errorf("GetAvailableProviders() count = %v, want %v", len(available), len(expectedProviders))
	}

	for _, expected := range expectedProviders {
		found := false
		for _, actual := range available {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetAvailableProviders() missing provider %v", expected)
		}
	}
}

func TestAIService_GetHealthyProviders(t *testing.T) {
	tests := []struct {
		name           string
		providerHealth map[string]bool
		expected       []string
	}{
		{
			name: "All providers healthy",
			providerHealth: map[string]bool{
				"gemini": true,
				"groq":   true,
			},
			expected: []string{"gemini", "groq"},
		},
		{
			name: "One provider unhealthy",
			providerHealth: map[string]bool{
				"gemini": true,
				"groq":   false,
			},
			expected: []string{"gemini"},
		},
		{
			name: "All providers unhealthy",
			providerHealth: map[string]bool{
				"gemini": false,
				"groq":   false,
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRCM := &MockRuntimeConfigManager{
				config: &state.RuntimeConfig{
					Providers: map[string]state.ProviderState{
						"gemini": {Enabled: true},
						"groq":   {Enabled: true},
					},
				},
			}

			// Create mock providers with health check
			providers := make(map[string]map[string]AIProvider)
			for providerName, healthy := range tt.providerHealth {
				mockProvider := &MockAIProvider{
					healthy: healthy,
				}
				providers[providerName] = map[string]AIProvider{
					"model": mockProvider,
				}
			}

			aiService := &AIService{
				providers: providers,
				sm:        mockRCM,
			}

			healthyProviders := aiService.GetHealthyProviders(context.Background())

			if len(healthyProviders) != len(tt.expected) {
				t.Errorf("GetHealthyProviders() count = %v, want %v", len(healthyProviders), len(tt.expected))
			}

			for _, expected := range tt.expected {
				found := false
				for _, actual := range healthyProviders {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetHealthyProviders() missing healthy provider %v", expected)
				}
			}
		})
	}
}

func TestAIService_SetProvider(t *testing.T) {
	tests := []struct {
		name          string
		providerName  string
		expectedError bool
	}{
		{
			name:          "Set valid provider",
			providerName:  "gemini",
			expectedError: false,
		},
		{
			name:          "Set invalid provider",
			providerName:  "invalid",
			expectedError: true,
		},
		{
			name:          "Set empty provider",
			providerName:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRCM := &MockRuntimeConfigManager{
				config: &state.RuntimeConfig{
					Providers: map[string]state.ProviderState{
						"gemini": {Enabled: true},
						"groq":   {Enabled: true},
					},
				},
			}

			aiService := &AIService{
				providers: map[string]map[string]AIProvider{
					"gemini": {
						"gemini-pro": &MockAIProvider{},
					},
				},
				sm: mockRCM,
			}

			err := aiService.SetProvider(tt.providerName)

			if tt.expectedError && err == nil {
				t.Error("SetProvider() expected error, got nil")
			} else if !tt.expectedError && err != nil {
				t.Errorf("SetProvider() unexpected error = %v", err)
			} else if !tt.expectedError {
				// Check that config was updated
				if mockRCM.config.ActiveProvider != tt.providerName {
					t.Errorf("SetProvider() active provider = %v, want %v", mockRCM.config.ActiveProvider, tt.providerName)
				}
			}
		})
	}
}

func TestAIService_ContextTimeout(t *testing.T) {
	mockRCM := &MockRuntimeConfigManager{
		config: &state.RuntimeConfig{
			Providers: map[string]state.ProviderState{
				"gemini": {Enabled: true},
			},
		},
	}

	// Create mock provider that times out
	mockProvider := &MockAIProvider{
		delay: 100 * time.Millisecond,
	}

	aiService := &AIService{
		providers: map[string]map[string]AIProvider{
			"gemini": {
				"gemini-pro": mockProvider,
			},
		},
		sm: mockRCM,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := &RequestModel{
		UserPrompt: "test prompt",
	}

	var callbackCalled bool
	err := aiService.Chat(ctx, req, func(chunk string) {
		callbackCalled = true
	})

	if err == nil {
		t.Error("Chat() expected timeout error, got nil")
	}

	if callbackCalled {
		t.Error("Chat() callback should not be called on timeout")
	}
}

// Mock implementations for testing

type MockRuntimeConfigManager struct {
	config *state.RuntimeConfig
}

func (m *MockRuntimeConfigManager) GetConfig() *state.RuntimeConfig {
	return m.config
}

func (m *MockRuntimeConfigManager) UpdateConfig(config *state.RuntimeConfig) error {
	m.config = config
	return nil
}

type MockAIProvider struct {
	healthy bool
	delay   time.Duration
	error   error
}

func (m *MockAIProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
			// Continue
		}
	}

	if m.error != nil {
		return nil, m.error
	}

	return &ResponseModel{
		Content: "mock response",
		Usage: TokenUsage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
		ProviderInfo: ModelInfo{
			ProviderName: "mock",
			ModelName:    "mock-model",
		},
	}, nil
}

func (m *MockAIProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse, 1)

	go func() {
		defer close(ch)

		if m.delay > 0 {
			select {
			case <-ctx.Done():
				ch <- StreamResponse{Error: ctx.Err()}
				return
			case <-time.After(m.delay):
				// Continue
			}
		}

		if m.error != nil {
			ch <- StreamResponse{Error: m.error}
			return
		}

		ch <- StreamResponse{Content: "mock response"}
		ch <- StreamResponse{Done: true}
	}()

	return ch, nil
}

func (m *MockAIProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "mock",
		ModelName:    "mock-model",
		Enabled:      m.healthy,
	}
}

func (m *MockAIProvider) CheckHealth(ctx context.Context) error {
	if !m.healthy {
		return fmt.Errorf("mock provider unhealthy")
	}
	return nil
}
