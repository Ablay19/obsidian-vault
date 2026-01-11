package ai

import (
	"context"
	"io"
	"net/http"
	"obsidian-automation/internal/telemetry"
	"strings"
	"time"

	"github.com/magicx-ai/groq-go/groq"
)

// GroqProvider implements the AIProvider interface for Groq.
type GroqProvider struct {
	client    groq.Client
	modelName string
	key       string
}

// NewGroqProvider creates a new Groq provider for a single API key.
func NewGroqProvider(apiKey string, modelName string, httpClient *http.Client) *GroqProvider {
	if apiKey == "" {
		telemetry.Info("Groq API key is empty. Groq AI will be unavailable for this provider instance.")
		return nil
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}

	client := groq.NewClient(apiKey, httpClient)

	return &GroqProvider{
		client:    client,
		modelName: modelName,
		key:       apiKey,
	}
}

// GenerateCompletion sends a request to the AI service and returns a complete response.
func (p *GroqProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	model := p.modelName
	if req.Model != "" {
		model = req.Model
	}

	messages := []groq.Message{}
	if req.SystemPrompt != "" {
		messages = append(messages, groq.Message{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, groq.Message{
		Role:    "user",
		Content: req.UserPrompt,
	})

	gReq := groq.ChatCompletionRequest{
		Model:    groq.ModelID(model),
		Messages: messages,
	}

	// ResponseFormat might not be supported in this SDK version, relying on prompt.
	if req.Temperature != 0 {
		gReq.Temperature = req.Temperature
	}
	if req.MaxTokens != 0 {
		gReq.MaxTokens = req.MaxTokens
	}

	resp, err := p.client.CreateChatCompletion(gReq)
	if err != nil {
		return nil, p.mapError(err)
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &ResponseModel{
		Content:      content,
		ProviderInfo: p.GetModelInfo(),
	}, nil
}

// StreamCompletion streams the response from the AI service.
func (p *GroqProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	model := p.modelName
	if req.Model != "" {
		model = req.Model
	}

	messages := []groq.Message{}
	if req.SystemPrompt != "" {
		messages = append(messages, groq.Message{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, groq.Message{
		Role:    "user",
		Content: req.UserPrompt,
	})

	gReq := groq.ChatCompletionRequest{
		Model:    groq.ModelID(model),
		Messages: messages,
		Stream:   true,
	}

	if req.Temperature != 0 {
		gReq.Temperature = req.Temperature
	}

	respCh, _, err := p.client.CreateChatCompletionStream(ctx, gReq)
	if err != nil {
		return nil, p.mapError(err)
	}

	outputChan := make(chan StreamResponse)

	go func() {
		defer close(outputChan)
		for res := range respCh {
			if res.Error != nil {
				if res.Error == io.EOF {
					outputChan <- StreamResponse{Done: true}
					return
				}
				outputChan <- StreamResponse{Error: p.mapError(res.Error)}
				return
			}
			if len(res.Response.Choices) > 0 {
				outputChan <- StreamResponse{Content: res.Response.Choices[0].Delta.Content}
			}
		}
	}()

	return outputChan, nil
}

func (p *GroqProvider) mapError(err error) error {
	// Simple mapping, actual SDK might have specific error types
	errStr := err.Error()
	if strings.Contains(errStr, "429") {
		return NewError(ErrCodeRateLimit, "groq rate limit exceeded", err)
	}
	if strings.Contains(errStr, "404") || strings.Contains(errStr, "400") {
		return NewError(ErrCodeInvalidRequest, "groq invalid request or model not found", err)
	}
	if strings.Contains(errStr, "401") {
		return NewError(ErrCodeUnauthorized, "groq unauthorized (check API key)", err)
	}
	if strings.Contains(errStr, "503") || strings.Contains(errStr, "500") {
		return NewError(ErrCodeProviderOffline, "groq service unavailable", err)
	}
	return NewError(ErrCodeInternal, "groq internal error", err)
}

// GetModelInfo returns information about the model.
func (p *GroqProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Groq",
		ModelName:    p.modelName,
	}
}

// CheckHealth verifies if the provider is currently operational.
func (p *GroqProvider) CheckHealth(ctx context.Context) error {
	req := groq.ChatCompletionRequest{
		Model: groq.ModelID(p.modelName),
		Messages: []groq.Message{
			{Role: "user", Content: "ping"},
		},
	}
	_, err := p.client.CreateChatCompletion(req)
	return err
}
