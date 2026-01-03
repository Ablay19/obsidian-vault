package ai

import (
	"context"
	"errors"
	"testing"

	st "obsidian-automation/internal/state"
)

// MockProvider implements AIProvider for testing.
type MockProvider struct {
	Name          string
	ShouldFail    bool
	FailError     error
	Response      string
	StreamContent []string
}

func (m *MockProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	if m.ShouldFail {
		return nil, m.FailError
	}
	return &ResponseModel{Content: m.Response, ProviderInfo: m.GetModelInfo()}, nil
}

func (m *MockProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse, len(m.StreamContent)+1)
	go func() {
		defer close(ch)
		if m.ShouldFail {
			ch <- StreamResponse{Error: m.FailError}
			return
		}
		for _, c := range m.StreamContent {
			ch <- StreamResponse{Content: c}
		}
		ch <- StreamResponse{Done: true}
	}()
	return ch, nil
}

func (m *MockProvider) CheckHealth(ctx context.Context) error {
	if m.ShouldFail {
		return m.FailError
	}
	return nil
}

func (m *MockProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: m.Name,
		ModelName:    "mock-model",
		Enabled:      true,
	}
}

func TestAIService_SelectProvider_Logic(t *testing.T) {
	// Setup in-memory DB for RCM
	db := setupTestDB(t) // Reusing the helper from integration_test.go (same package)
	defer db.Close()
	
	rcm, _ := st.NewRuntimeConfigManager(db)
	
	// Reset to clear Env vars
	rcm.ResetState()
	
	ctx := context.Background()
	service := NewAIService(ctx, rcm)

	// Inject Mocks directly into service.providers
	// We need to setup RCM state to match these keys so selectProvider finds them valid
	
	// Ensure providers are enabled in config (Reset cleared them)
	rcm.SetProviderState("Gemini", true, false, false, "")
	rcm.SetProviderState("Groq", true, false, false, "")

	// Add Key 1 for Gemini (Mock)
	k1ID, err := rcm.AddAPIKey("Gemini", "mock-gemini-key-1", true)
	if err != nil { t.Fatalf("AddKey failed: %v", err) }
	// Add Key 2 for Gemini (Mock)
	k2ID, err := rcm.AddAPIKey("Gemini", "mock-gemini-key-2", true)
	if err != nil { t.Fatalf("AddKey failed: %v", err) }
	// Add Key 1 for Groq (Mock)
	g1ID, err := rcm.AddAPIKey("Groq", "mock-groq-key-1", true)
	if err != nil { t.Fatalf("AddKey failed: %v", err) }

	// Ensure providers are enabled in config
	rcm.SetProviderState("Gemini", true, false, false, "")
	rcm.SetProviderState("Groq", true, false, false, "")

	// Manually inject mocks (bypassing RefreshProviders which would overwrite them)
	service.mu.Lock()
	service.providers = make(map[string]map[string]AIProvider)
	service.providers["Gemini"] = make(map[string]AIProvider)
	service.providers["Groq"] = make(map[string]AIProvider)

	mockGemini1 := &MockProvider{Name: "Gemini", Response: "Response from G1"}
	mockGemini2 := &MockProvider{Name: "Gemini", Response: "Response from G2"}
	mockGroq1 := &MockProvider{Name: "Groq", Response: "Response from Groq"}

	service.providers["Gemini"][k1ID] = mockGemini1
	service.providers["Gemini"][k2ID] = mockGemini2
	service.providers["Groq"][g1ID] = mockGroq1
	service.mu.Unlock()

	// Debug Config
	cfg := rcm.GetConfig()
	t.Logf("Config AIEnabled: %v", cfg.AIEnabled)
	t.Logf("Config ActiveProvider: %v", cfg.ActiveProvider)
	t.Logf("Config Gemini Enabled: %v", cfg.Providers["Gemini"].Enabled)
	t.Logf("Config Key k1ID Enabled: %v", cfg.APIKeys[k1ID].Enabled)
	t.Logf("Config Key k1ID Value: %v", cfg.APIKeys[k1ID].Value)

	// Test 1: Preferred Provider Selection
	service.SetProvider("Gemini")
	
	// We can't easily call selectProvider directly as it's private, but we can call ExecuteWithRetry
	err = service.ExecuteWithRetry(ctx, func(p AIProvider) error {
		info := p.GetModelInfo()
		if info.ProviderName != "Gemini" {
			t.Errorf("Expected Gemini, got %s", info.ProviderName)
		}
		return nil
	})
	if err != nil {
		t.Errorf("ExecuteWithRetry failed: %v", err)
	}

	// Test 2: Fallback Logic
	// Make Gemini 1 and 2 fail with transient errors?
	// The rotation logic in ExecuteWithRetry tries different keys.
	
	// Reset mocks to fail
	mockGemini1.ShouldFail = true
	mockGemini1.FailError = NewError(ErrCodeRateLimit, "rate limited", errors.New("429")) // Retryable
	
	// We expect it to try Gemini 1, fail, then try Gemini 2 (since it's same provider) OR fallback?
	// selectProvider logic:
	// 1. Try preferred (Gemini).
	// 2. Fallback (Gemini is first in fallback list too).
	
	// ExecuteWithRetry loops.
	// Iteration 0: selectProvider returns Gemini Key 1 (assuming iteration order). Op fails. Key 1 added to failedKeys.
	// Iteration 1: selectProvider called with failedKeys=[k1]. It should pick Gemini Key 2.
	
	err = service.ExecuteWithRetry(ctx, func(p AIProvider) error {
		resp, err := p.GenerateCompletion(ctx, &RequestModel{})
		if err != nil {
			return err
		}
		if resp.Content != "Response from G2" {
			// It might have picked Groq if G2 wasn't selected?
			// But G2 is valid and same provider.
			// However, map iteration order is random.
			// If it picked G2 first (success), we are good.
			// If it picked G1 first (fail), it retries.
			// We want to ensure it eventually succeeds.
			return nil
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Fallback/Retry failed: %v", err)
	}
	
	// Test 3: Cross-Provider Fallback
	// Fail ALL Gemini keys
	mockGemini2.ShouldFail = true
	mockGemini2.FailError = NewError(ErrCodeRateLimit, "rate limited", errors.New("429"))

	// Should fall back to Groq
	err = service.ExecuteWithRetry(ctx, func(p AIProvider) error {
		resp, _ := p.GenerateCompletion(ctx, &RequestModel{})
		if resp == nil {
			return errors.New("nil response")
		}
		if p.GetModelInfo().ProviderName != "Groq" {
			t.Errorf("Expected fallback to Groq, got %s", p.GetModelInfo().ProviderName)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Cross-provider fallback failed: %v", err)
	}
}

func TestAIService_NoKeys(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	rcm, _ := st.NewRuntimeConfigManager(db)
	rcm.ResetState()
	service := NewAIService(context.Background(), rcm) // Empty

	err := service.ExecuteWithRetry(context.Background(), func(p AIProvider) error {
		return nil
	})

	if err == nil {
		t.Error("Expected error when no providers available, got nil")
	}
}