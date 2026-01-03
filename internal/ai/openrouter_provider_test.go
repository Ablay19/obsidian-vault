package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestOpenRouterProvider_GenerateCompletion_Success(t *testing.T) {
	mockResponse := openRouterChatResponse{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: "Hello from OpenRouter!",
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

	provider := NewOpenRouterProvider("fake-key", "gpt-3.5-turbo", client)
	if provider == nil {
		t.Fatal("Expected provider initialization")
	}

	req := &RequestModel{UserPrompt: "Hi"}
	resp, err := provider.GenerateCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateCompletion failed: %v", err)
	}

	if resp.Content != "Hello from OpenRouter!" {
		t.Errorf("Expected 'Hello from OpenRouter!', got '%s'", resp.Content)
	}
}

func TestOpenRouterProvider_GenerateCompletion_Error(t *testing.T) {
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

	provider := NewOpenRouterProvider("fake-key", "gpt-3.5-turbo", client)

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

func TestOpenRouterProvider_StreamCompletion_Success(t *testing.T) {
	// SSE stream format
	streamBody := `data: {"choices": [{"delta": {"content": "Hello"}}]}
data: {"choices": [{"delta": {"content": " World"}}]}
data: [DONE]
`
	client := &http.Client{
		Transport: &MockTransport{
			RoundTripFn: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(streamBody)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	provider := NewOpenRouterProvider("fake-key", "gpt-3.5-turbo", client)
	req := &RequestModel{UserPrompt: "Hi"}

	stream, err := provider.StreamCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("StreamCompletion failed: %v", err)
	}

	var content string
	for chunk := range stream {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}
		content += chunk.Content
	}

	if content != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", content)
	}
}
