package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/option"
)

func TestGeminiProvider_GenerateCompletion_Success(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL or headers if needed
		if r.URL.Path != "/v1beta/models/gemini-pro:generateContent" {
			// Note: The path might vary based on client version, but usually follows this pattern
			// ignoring for now, just serving content
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Hello from Gemini!"},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Create Provider with mock client
	ctx := context.Background()
	// google-api-go-client allows overriding endpoint via WithEndpoint
	provider := NewGeminiProvider(ctx, "fake-key", "gemini-pro", option.WithEndpoint(ts.URL), option.WithHTTPClient(ts.Client()))

	if provider == nil {
		t.Fatal("Expected provider to be initialized")
	}

	// Test GenerateCompletion
	req := &RequestModel{
		UserPrompt: "Hello",
	}

	resp, err := provider.GenerateCompletion(ctx, req)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if resp.Content != "Hello from Gemini!" {
		t.Errorf("Expected 'Hello from Gemini!', got '%s'", resp.Content)
	}
}

func TestGeminiProvider_GenerateCompletion_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests) // 429
		w.Write([]byte(`{"error": {"code": 429, "message": "Quota exceeded"}}`))
	}))
	defer ts.Close()

	ctx := context.Background()
	provider := NewGeminiProvider(ctx, "fake-key", "gemini-pro", option.WithEndpoint(ts.URL), option.WithHTTPClient(ts.Client()))

	req := &RequestModel{UserPrompt: "Hello"}
	_, err := provider.GenerateCompletion(ctx, req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}

	if appErr.Code != ErrCodeRateLimit {
		t.Errorf("Expected ErrCodeRateLimit (429), got %d", appErr.Code)
	}
}

func TestGeminiProvider_GetModelInfo(t *testing.T) {
	ctx := context.Background()
	provider := NewGeminiProvider(ctx, "fake-key", "gemini-pro")
	
	info := provider.GetModelInfo()
	if info.ProviderName != "Gemini" {
		t.Errorf("Expected ProviderName 'Gemini', got '%s'", info.ProviderName)
	}
	if info.ModelName != "gemini-pro" {
		t.Errorf("Expected ModelName 'gemini-pro', got '%s'", info.ModelName)
	}
}
