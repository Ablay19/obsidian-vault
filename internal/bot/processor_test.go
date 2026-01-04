package bot

import (
	"context"
	"errors"
	"obsidian-automation/internal/ai"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// MockAIService is a mock implementation of the AIService for testing.
type MockAIService struct {
	ChatFunc                  func(ctx context.Context, req *ai.RequestModel, callback func(string)) error
	AnalyzeTextFunc           func(ctx context.Context, text, language string) (*ai.AnalysisResult, error)
	AnalyzeTextWithParamsFunc func(ctx context.Context, text, language string, task_tokens int, task_depth int, max_cost float64) (*ai.AnalysisResult, error)
	GetActiveProviderNameFunc func() string
	SetProviderFunc           func(providerName string) error
	GetAvailableProvidersFunc func() []string
	GetHealthyProvidersFunc   func(ctx context.Context) []string
	GetProvidersInfoFunc      func() []ai.ModelInfo
}

func (m *MockAIService) Chat(ctx context.Context, req *ai.RequestModel, callback func(string)) error {
	if m.ChatFunc != nil {
		return m.ChatFunc(ctx, req, callback)
	}
	return errors.New("ChatFunc not implemented")
}

func (m *MockAIService) AnalyzeText(ctx context.Context, text, language string) (*ai.AnalysisResult, error) {
	if m.AnalyzeTextFunc != nil {
		return m.AnalyzeTextFunc(ctx, text, language)
	}
	return nil, errors.New("AnalyzeTextFunc not implemented")
}

func (m *MockAIService) AnalyzeTextWithParams(ctx context.Context, text, language string, task_tokens int, task_depth int, max_cost float64) (*ai.AnalysisResult, error) {
	if m.AnalyzeTextWithParamsFunc != nil {
		return m.AnalyzeTextWithParamsFunc(ctx, text, language, task_tokens, task_depth, max_cost)
	}
	return nil, errors.New("AnalyzeTextWithParamsFunc not implemented")
}

func (m *MockAIService) GetActiveProviderName() string {
	if m.GetActiveProviderNameFunc != nil {
		return m.GetActiveProviderNameFunc()
	}
	return "mock"
}

func (m *MockAIService) SetProvider(providerName string) error {
	if m.SetProviderFunc != nil {
		return m.SetProviderFunc(providerName)
	}
	return errors.New("SetProviderFunc not implemented")
}

func (m *MockAIService) GetAvailableProviders() []string {
	if m.GetAvailableProvidersFunc != nil {
		return m.GetAvailableProvidersFunc()
	}
	return []string{"mock"}
}

func (m *MockAIService) GetHealthyProviders(ctx context.Context) []string {
	if m.GetHealthyProvidersFunc != nil {
		return m.GetHealthyProvidersFunc(ctx)
	}
	return []string{"mock"}
}

func (m *MockAIService) GetProvidersInfo() []ai.ModelInfo {
	if m.GetProvidersInfoFunc != nil {
		return m.GetProvidersInfoFunc()
	}
	return []ai.ModelInfo{{ProviderName: "mock", ModelName: "mock-model"}}
}

func TestProcessFileWithAI_Success(t *testing.T) {
	// 1. Setup Mock AI Service
	mockAI := &MockAIService{
		GetActiveProviderNameFunc: func() string {
			return "mock-provider"
		},
		ChatFunc: func(ctx context.Context, req *ai.RequestModel, callback func(string)) error {
			callback("This is a summary.")
			return nil
		},
		AnalyzeTextWithParamsFunc: func(ctx context.Context, text, language string, task_tokens int, task_depth int, max_cost float64) (*ai.AnalysisResult, error) {
			return &ai.AnalysisResult{
				Category:  "tech",
				Topics:    []string{"golang", "testing"},
				Questions: []string{"Is this a test?"},
			}, nil
		},
	}

	// 2. Mock execCommand (the package-level variable)
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }() // Restore original after test
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Return a command that echoes "mocked text" to stdout
		cmd := exec.Command("echo", "mocked text") // Use original exec.Command for the echo
		return cmd
	}

	// 3. Create a temporary file
	tmpfile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	filePath := tmpfile.Name()

	// 4. Call the function
	ctx := context.Background()
	updateStatus := func(s string) { t.Logf("UpdateStatus: %s", s) }
	streamCallback := func(s string) { t.Logf("StreamCallback: %s", s) }

	result := processFileWithAI(ctx, filePath, "pdf", mockAI, streamCallback, "english", updateStatus, "")

	// 5. Assertions
	expected := ProcessedContent{
		Text:       "mocked text",
		Category:   "tech",
		Tags:       []string{"tech", "golang", "testing"},
		Confidence: 0.95,
		Language:   "english",
		Summary:    "This is a summary.",
		Topics:     []string{"golang", "testing"},
		Questions:  []string{"Is this a test?"},
		AIProvider: "mock-provider",
	}

	if result.Summary != expected.Summary {
		t.Errorf("Expected Summary '%s', got '%s'", expected.Summary, result.Summary)
	}
	if result.Category != expected.Category {
		t.Errorf("Expected Category '%s', got '%s'", expected.Category, result.Category)
	}
	if !reflect.DeepEqual(result.Topics, expected.Topics) {
		t.Errorf("Expected Topics %v, got %v", expected.Topics, result.Topics)
	}
	if !reflect.DeepEqual(result.Questions, expected.Questions) {
		t.Errorf("Expected Questions %v, got %v", expected.Questions, result.Questions)
	}
	if result.AIProvider != expected.AIProvider {
		t.Errorf("Expected AIProvider '%s', got '%s'", expected.AIProvider, result.AIProvider)
	}
}
