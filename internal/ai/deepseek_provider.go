package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"obsidian-automation/internal/telemetry"
	"time"
)

type DeepSeekProvider struct {
	apiKey   string
	model    string
	endpoint string
}

func NewDeepSeekProvider(apiKey, model string) AIProvider {
	return &DeepSeekProvider{
		apiKey:   apiKey,
		model:    model,
		endpoint: "https://api.deepseek.com/v1/chat/completions",
	}
}

// GenerateCompletion sends a request to DeepSeek and returns a complete response
func (p *DeepSeekProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	telemetry.Info("DeepSeek GenerateCompletion called", "model", p.model, "prompt_length", len(req.UserPrompt))

	// Prepare DeepSeek request (compatible with OpenAI API format)
	openaiReq := struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
		Stream      bool    `json:"stream"`
	}{
		Model: p.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: req.UserPrompt},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false, // Use non-streaming for simplicity
	}

	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DeepSeek API error: %d", resp.StatusCode)
	}

	var deepseekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in DeepSeek response")
	}

	response := &ResponseModel{
		Content: deepseekResp.Choices[0].Message.Content,
		Usage: TokenUsage{
			InputTokens:  deepseekResp.Usage.PromptTokens,
			OutputTokens: deepseekResp.Usage.CompletionTokens,
			TotalTokens:  deepseekResp.Usage.TotalTokens,
		},
		ProviderInfo: ModelInfo{
			ProviderName: "DeepSeek",
			ModelName:    p.model,
			Enabled:      true,
		},
	}

	telemetry.Info("DeepSeek GenerateCompletion completed", "response_length", len(response.Content), "total_tokens", response.Usage.TotalTokens)
	return response, nil
}

// StreamCompletion streams the response from DeepSeek
func (p *DeepSeekProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	telemetry.Info("DeepSeek StreamCompletion called", "model", p.model)

	streamChan := make(chan StreamResponse, 10)

	go func() {
		defer close(streamChan)

		// For simplicity, we'll do a single completion and send it as one chunk
		// In a production implementation, you'd implement proper streaming
		response, err := p.GenerateCompletion(ctx, req)
		if err != nil {
			streamChan <- StreamResponse{Error: err}
			return
		}

		// Send the content
		streamChan <- StreamResponse{
			Content: response.Content,
			Done:    false,
		}

		// Send final done signal
		streamChan <- StreamResponse{
			Content: "",
			Done:    true,
		}

		telemetry.Info("DeepSeek StreamCompletion completed", "response_length", len(response.Content))
	}()

	return streamChan, nil
}

// GetModelInfo returns information about the DeepSeek model
func (p *DeepSeekProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "DeepSeek",
		ModelName:    p.model,
		Enabled:      true,
	}
}

// CheckHealth verifies if the DeepSeek provider is operational
func (p *DeepSeekProvider) CheckHealth(ctx context.Context) error {
	healthReq := &RequestModel{
		UserPrompt:  "Hello",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := p.GenerateCompletion(ctx, healthReq)
	if err != nil {
		telemetry.Warn("DeepSeek health check failed", "error", err)
		return err
	}

	telemetry.Info("DeepSeek health check passed")
	return nil
}
