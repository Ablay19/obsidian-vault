package bot

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/pipeline"
	"os"
	"testing"
	"time"
)

// MockAIProcessor implements ai.AIServiceInterface for testing
type MockAIProcessor struct {
	analyzeTextFunc           func(ctx context.Context, text, language string) (*ai.AnalysisResult, error)
	chatFunc                  func(ctx context.Context, req *ai.RequestModel, callback func(string)) error
	setProviderFunc           func(providerName string) error
	getActiveProviderNameFunc func() string
	getAvailableProvidersFunc func() []string
}

func (m *MockAIProcessor) AnalyzeText(ctx context.Context, text, language string) (*ai.AnalysisResult, error) {
	if m.analyzeTextFunc != nil {
		return m.analyzeTextFunc(ctx, text, language)
	}
	return &ai.AnalysisResult{}, nil
}

func (m *MockAIProcessor) AnalyzeTextWithParams(ctx context.Context, text, language string, taskTokens int, taskDepth int, maxCost float64) (*ai.AnalysisResult, error) {
	return &ai.AnalysisResult{}, nil
}

func (m *MockAIProcessor) Chat(ctx context.Context, req *ai.RequestModel, callback func(string)) error {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, req, callback)
	}
	return nil
}

func (m *MockAIProcessor) SetProvider(providerName string) error {
	if m.setProviderFunc != nil {
		return m.setProviderFunc(providerName)
	}
	return nil
}

func (m *MockAIProcessor) GetActiveProviderName() string {
	if m.getActiveProviderNameFunc != nil {
		return m.getActiveProviderNameFunc()
	}
	return "mock-provider"
}

func (m *MockAIProcessor) GetAvailableProviders() []string {
	if m.getAvailableProvidersFunc != nil {
		return m.getAvailableProvidersFunc()
	}
	return []string{"mock-provider"}
}

func (m *MockAIProcessor) GetHealthyProviders(ctx context.Context) []string {
	return []string{"mock-provider"}
}

func (m *MockAIProcessor) GetProvidersInfo() []ai.ModelInfo {
	return []ai.ModelInfo{
		{
			ProviderName: "mock-provider",
			ModelName:    "mock-model",
		},
	}
}

func TestBotProcessor_Process_TextContent_Success(t *testing.T) {
	mockAI := &MockAIProcessor{
		analyzeTextFunc: func(ctx context.Context, text, language string) (*ai.AnalysisResult, error) {
			return &ai.AnalysisResult{
				Category:  "document",
				Topics:    []string{"topic1", "topic2"},
				Questions: []string{"question1"},
			}, nil
		},
	}

	processor := NewBotProcessor(mockAI)

	// Create a temporary test file
	tmpFile := "/tmp/test_content.txt"
	err := os.WriteFile(tmpFile, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	job := pipeline.Job{
		ID:            "test-job-1",
		Data:          []byte("test data"),
		FileLocalPath: tmpFile,
		ContentType:   pipeline.ContentTypeText,
		ReceivedAt:    time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   "test-user",
			Language: "English",
		},
		Metadata: map[string]interface{}{
			"caption": "test caption",
		},
	}

	result, err := processor.Process(context.Background(), job)
	if err != nil {
		t.Errorf("Process() error = %v", err)
		return
	}

	if result.Success != true {
		t.Errorf("Process() Success = %v, want %v", result.Success, true)
	}
}

func TestBotProcessor_Process_AnalyzeTextError(t *testing.T) {
	mockAI := &MockAIProcessor{
		analyzeTextFunc: func(ctx context.Context, text, language string) (*ai.AnalysisResult, error) {
			return nil, fmt.Errorf("AI service error")
		},
	}

	processor := NewBotProcessor(mockAI)

	// Create a temporary test file
	tmpFile := "/tmp/test_content.txt"
	err := os.WriteFile(tmpFile, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	job := pipeline.Job{
		ID:            "test-job-1",
		Data:          []byte("test data"),
		FileLocalPath: tmpFile,
		ContentType:   pipeline.ContentTypeText,
		ReceivedAt:    time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   "test-user",
			Language: "English",
		},
	}

	_, err = processor.Process(context.Background(), job)
	if err == nil {
		t.Error("Expected error from Process(), got nil")
	}
}

func TestBotProcessor_Process_ContextCancellation(t *testing.T) {
	mockAI := &MockAIProcessor{
		analyzeTextFunc: func(ctx context.Context, text, language string) (*ai.AnalysisResult, error) {
			// Simulate long-running operation
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return &ai.AnalysisResult{
					Category:  "should not reach here",
					Topics:    []string{},
					Questions: []string{},
				}, nil
			}
		},
	}

	processor := NewBotProcessor(mockAI)

	// Create a temporary test file
	tmpFile := "/tmp/test_content.txt"
	err := os.WriteFile(tmpFile, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	job := pipeline.Job{
		ID:            "test-job-1",
		Data:          []byte("test data"),
		FileLocalPath: tmpFile,
		ContentType:   pipeline.ContentTypeText,
		ReceivedAt:    time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   "test-user",
			Language: "English",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = processor.Process(ctx, job)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}

	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("Expected context cancellation error, got %v", err)
	}
}
