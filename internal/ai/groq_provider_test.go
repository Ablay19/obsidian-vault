package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// MockTransport allows mocking HTTP responses
type MockTransport struct {
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.RoundTripFn(req)
}

func TestGroqProvider_GenerateCompletion_Success(t *testing.T) {
	mockResponse := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": "Hello from Groq!",
				},
			},
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	client := &http.Client{
		Transport: &MockTransport{
			RoundTripFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(respBytes)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	provider := NewGroqProvider("fake-key", "llama3-70b", client)
	if provider == nil {
		t.Fatal("Expected provider initialization")
	}

	req := &RequestModel{UserPrompt: "Hi"}
	resp, err := provider.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if resp.Content != "Hello from Groq!" {
		t.Errorf("Expected 'Hello from Groq!', got '%s'", resp.Content)
	}
}

func TestGroqProvider_GenerateCompletion_Error(t *testing.T) {
	client := &http.Client{
		Transport: &MockTransport{
			RoundTripFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 401,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": {"message": "Invalid API key"}}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	provider := NewGroqProvider("fake-key", "llama3-70b", client)

	req := &RequestModel{UserPrompt: "Hi"}
	_, err := provider.GenerateCompletion(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("Expected AppError, got %T", err)
	}

	if appErr.Code != ErrCodeUnauthorized {
		t.Errorf("Expected ErrCodeUnauthorized (401), got %d", appErr.Code)
	}
}
