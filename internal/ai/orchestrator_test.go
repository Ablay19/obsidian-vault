package ai

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ResponseModel), args.Error(1)
}

func (m *MockAIProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(<-chan StreamResponse), args.Error(1)
}

func (m *MockAIProvider) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAIProvider) GetModelInfo() ModelInfo {
	args := m.Called()
	return args.Get(0).(ModelInfo)
}

type MockSessionStore struct {
	mock.Mock
}

func (m *MockSessionStore) Store(ctx context.Context, session interface{}) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionStore) Get(ctx context.Context, sessionID string) (interface{}, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0), args.Error(1)
}

func (m *MockSessionStore) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionStore) Cleanup(ctx context.Context, expiredBefore time.Time) error {
	args := m.Called(ctx, expiredBefore)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) AIRequest(provider, model string, inputTokens int, userID string) {
	m.Called(provider, model, inputTokens, userID)
}

func (m *MockLogger) AIResponse(provider, model string, outputTokens int, duration int64, userID string) {
	m.Called(provider, model, outputTokens, duration, userID)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func setupTestOrchestrator(t *testing.T) (*Orchestrator, *MockLogger) {
	mockLogger := &MockLogger{}
	orc := NewOrchestrator(mockLogger)
	return orc, mockLogger
}

func TestNewOrchestrator(t *testing.T) {
	mockLogger := &MockLogger{}
	orc := NewOrchestrator(mockLogger)

	assert.NotNil(t, orc)
	assert.NotNil(t, orc.providers)
	assert.NotNil(t, orc.localModels)
	assert.NotNil(t, orc.apiProviders)
	assert.Equal(t, mockLogger, orc.logger)
	assert.Contains(t, orc.fallbackChain, "local")
	assert.Contains(t, orc.fallbackChain, "huggingface")
}

func TestOrchestrator_AddProvider(t *testing.T) {
	orc, mockLogger := setupTestOrchestrator(t)
	mockProvider := &MockAIProvider{}

	// Mock GetModelInfo
	mockProvider.On("GetModelInfo").Return(ModelInfo{
		ProviderName: "test_provider",
		ModelName:    "test_model",
	})

	// Mock logger
	mockLogger.On("Info", "Added AI provider", "name", "test_provider", "model", "test_model")

	orc.AddProvider("test_provider", mockProvider)

	assert.Contains(t, orc.providers, "test_provider")
	assert.Contains(t, orc.apiProviders, "test_provider") // Since it's not local

	mockProvider.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestOrchestrator_AddProvider_Local(t *testing.T) {
	orc, mockLogger := setupTestOrchestrator(t)
	mockProvider := &MockAIProvider{}

	mockProvider.On("GetModelInfo").Return(ModelInfo{
		ProviderName: "local",
		ModelName:    "local_model",
	})

	mockLogger.On("Info", "Added AI provider", "name", "local", "model", "local_model")

	orc.AddProvider("local", mockProvider)

	assert.Contains(t, orc.providers, "local")
	assert.Contains(t, orc.localModels, "local")

	mockProvider.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestOrchestrator_SetDefaultProvider(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)
	mockProvider := &MockAIProvider{}
	mockProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "test", ModelName: "test"})

	orc.AddProvider("test_provider", mockProvider)

	err := orc.SetDefaultProvider("test_provider")
	assert.NoError(t, err)
	assert.Equal(t, "test_provider", orc.currentProvider)
}

func TestOrchestrator_SetDefaultProvider_NotFound(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	err := orc.SetDefaultProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider 'nonexistent' not found")
}

func TestOrchestrator_Generate_Success(t *testing.T) {
	orc, mockLogger := setupTestOrchestrator(t)
	mockProvider := &MockAIProvider{}

	// Setup provider
	mockProvider.On("GetModelInfo").Return(ModelInfo{
		ProviderName: "test_provider",
		ModelName:    "test_model",
	})
	mockProvider.On("CheckHealth", mock.Anything).Return(nil)

	expectedResponse := &ResponseModel{
		Content: "Test response",
		Usage: TokenUsage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
		ProviderInfo: ModelInfo{
			ProviderName: "test_provider",
			ModelName:    "test_model",
		},
	}

	mockProvider.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("*ai.RequestModel")).Return(expectedResponse, nil)

	// Setup logger
	mockLogger.On("AIRequest", "test_provider", "test_model", mock.AnythingOfType("int"), "test_user")
	mockLogger.On("AIResponse", "test_provider", "test_model", 20, mock.AnythingOfType("int64"), "test_user")

	orc.AddProvider("test_provider", mockProvider)
	orc.SetDefaultProvider("test_provider")

	options := &GenerationOptions{
		Model:     "test_model",
		UserID:    "test_user",
		MaxTokens: 100,
	}

	result, err := orc.Generate(context.Background(), "Test prompt", options)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test response", result.Content)
	assert.Equal(t, "test_provider", result.Provider)
	assert.Equal(t, "test_model", result.Model)
	assert.Equal(t, "test_user", result.UserID)
	assert.Equal(t, 10, result.InputTokens)
	assert.Equal(t, 20, result.OutputTokens)
	assert.True(t, result.Cost > 0) // Should have calculated cost

	mockProvider.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestOrchestrator_Generate_Fallback(t *testing.T) {
	orc, mockLogger := setupTestOrchestrator(t)

	// Failing provider
	failingProvider := &MockAIProvider{}
	failingProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "failing", ModelName: "failing"})
	failingProvider.On("CheckHealth", mock.Anything).Return(errors.New("health check failed"))

	// Successful fallback provider
	successProvider := &MockAIProvider{}
	successProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "success", ModelName: "success"})
	successProvider.On("CheckHealth", mock.Anything).Return(nil)
	successProvider.On("GenerateCompletion", mock.Anything, mock.AnythingOfType("*ai.RequestModel")).Return(&ResponseModel{
		Content:      "Fallback response",
		Usage:        TokenUsage{InputTokens: 5, OutputTokens: 10},
		ProviderInfo: ModelInfo{ProviderName: "success", ModelName: "success"},
	}, nil)

	mockLogger.On("AIRequest", "success", "success", mock.AnythingOfType("int"), "")
	mockLogger.On("AIResponse", "success", "success", 10, mock.AnythingOfType("int64"), "")

	orc.AddProvider("failing", failingProvider)
	orc.AddProvider("success", successProvider)
	orc.fallbackChain = []string{"failing", "success"}

	options := &GenerationOptions{Model: "test"}
	result, err := orc.Generate(context.Background(), "Test prompt", options)

	assert.NoError(t, err)
	assert.Equal(t, "Fallback response", result.Content)
	assert.Equal(t, "success", result.Provider)
	assert.Equal(t, "success", orc.currentProvider) // Should update to successful provider

	failingProvider.AssertExpectations(t)
	successProvider.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestOrchestrator_Generate_AllFail(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	failingProvider := &MockAIProvider{}
	failingProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "failing", ModelName: "failing"})
	failingProvider.On("CheckHealth", mock.Anything).Return(errors.New("health check failed"))

	orc.AddProvider("failing", failingProvider)
	orc.fallbackChain = []string{"failing"}

	options := &GenerationOptions{Model: "test"}
	result, err := orc.Generate(context.Background(), "Test prompt", options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "all AI providers failed")

	failingProvider.AssertExpectations(t)
}

func TestOrchestrator_GetAvailableProviders(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	healthyProvider := &MockAIProvider{}
	healthyProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "healthy", ModelName: "healthy"})
	healthyProvider.On("CheckHealth", mock.Anything).Return(nil)

	unhealthyProvider := &MockAIProvider{}
	unhealthyProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "unhealthy", ModelName: "unhealthy"})
	unhealthyProvider.On("CheckHealth", mock.Anything).Return(errors.New("unhealthy"))

	orc.AddProvider("healthy", healthyProvider)
	orc.AddProvider("unhealthy", unhealthyProvider)

	available := orc.GetAvailableProviders()

	assert.Contains(t, available, "healthy")
	assert.NotContains(t, available, "unhealthy")

	healthyProvider.AssertExpectations(t)
	unhealthyProvider.AssertExpectations(t)
}

func TestOrchestrator_GetProviderInfo(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	mockProvider := &MockAIProvider{}
	mockProvider.On("GetModelInfo").Return(ModelInfo{
		ProviderName: "test",
		ModelName:    "test_model",
	})
	mockProvider.On("CheckHealth", mock.Anything).Return(nil)

	orc.AddProvider("test", mockProvider)

	info := orc.GetProviderInfo("test")

	assert.NotNil(t, info)
	assert.Equal(t, "test", info["name"])
	assert.Equal(t, "test_model", info["model"])
	assert.Equal(t, true, info["available"])

	mockProvider.AssertExpectations(t)
}

func TestOrchestrator_GetProviderInfo_NotFound(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	info := orc.GetProviderInfo("nonexistent")
	assert.Nil(t, info)
}

func TestOrchestrator_GetStats(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	mockProvider := &MockAIProvider{}
	mockProvider.On("GetModelInfo").Return(ModelInfo{ProviderName: "test", ModelName: "test"})

	orc.AddProvider("test", mockProvider)

	stats := orc.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats["total_providers"])
	assert.Equal(t, 0, stats["local_providers"]) // test is not local
	assert.Equal(t, 1, stats["api_providers"])
	assert.Empty(t, stats["current_provider"])
	assert.Contains(t, stats["available_providers"], "test")

	mockProvider.AssertExpectations(t)
}

func TestIsLocalProvider(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"local", true},
		{"gpt2", true},
		{"gpt-neo", true},
		{"codellama", true},
		{"starcoder", true},
		{"openai", false},
		{"anthropic", false},
		{"huggingface", false},
		{"replicate", false},
		{"", false},
	}

	for _, test := range tests {
		result := isLocalProvider(test.name)
		assert.Equal(t, test.expected, result, "Failed for provider: %s", test.name)
	}
}

func TestCalculateCost(t *testing.T) {
	orc, _ := setupTestOrchestrator(t)

	tests := []struct {
		provider     string
		inputTokens  int
		outputTokens int
		expectedCost float64
	}{
		{"openai", 1000, 1000, 1.5},    // $0.001 * 1000 + $0.003 * 1000 = $1.5
		{"anthropic", 1000, 1000, 1.7}, // $0.011 * 1000 + $0.032 * 1000 = $1.7
		{"gemini", 1000, 1000, 0.9},    // $0.005 * 1000 + $0.015 * 1000 = $0.9
		{"groq", 1000, 1000, 0.7},      // $0.005 * 1000 + $0.01 * 1000 = $0.7
		{"deepseek", 1000, 1000, 0.38}, // $0.0014 * 1000 + $0.0028 * 1000 = $0.38
		{"local", 1000, 1000, 0.0},     // Free
		{"unknown", 1000, 1000, 1.5},   // Defaults to OpenAI pricing
	}

	for _, test := range tests {
		cost := orc.calculateCost(test.provider, test.inputTokens, test.outputTokens)
		assert.Equal(t, test.expectedCost, cost, "Failed for provider: %s", test.provider)
	}
}
