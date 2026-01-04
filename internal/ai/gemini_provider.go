package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiProvider implements the ai.AIProvider interface for Google Gemini.
type GeminiProvider struct {
	client    *genai.Client
	modelName string
	key       string
	mu        sync.Mutex
}

// NewGeminiProvider creates a new Gemini provider for a single API key.
func NewGeminiProvider(ctx context.Context, apiKey string, modelName string, opts ...option.ClientOption) *GeminiProvider {
	if apiKey == "" {
		zap.S().Info("Gemini API key is empty. Gemini AI will be unavailable for this provider instance.")
		return nil
	}

	finalOpts := append([]option.ClientOption{option.WithAPIKey(apiKey)}, opts...)
	client, err := genai.NewClient(ctx, finalOpts...)
	if err != nil {
		zap.S().Error("Error creating Gemini client", "error", err)
		return nil
	}

	return &GeminiProvider{
		client:    client,
		modelName: modelName,
		key:       apiKey,
	}
}

// GenerateCompletion sends a request to the AI service and returns a complete response.
func (p *GeminiProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	model := p.client.GenerativeModel(p.modelName)
	if req.Model != "" {
		model = p.client.GenerativeModel(req.Model)
	}

	if req.Temperature != 0 {
		model.SetTemperature(float32(req.Temperature))
	}
	if req.MaxTokens != 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}
	if req.JSONMode {
		model.ResponseMIMEType = "application/json"
	}

	var parts []genai.Part
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
		}
	}

	parts = append(parts, genai.Text(req.UserPrompt))
	if len(req.ImageData) > 0 {
		parts = append(parts, genai.ImageData("jpeg", req.ImageData))
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, p.mapError(err)
	}

	content := ""
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			content = string(txt)
		}
	}

	return &ResponseModel{
		Content:      content,
		ProviderInfo: p.GetModelInfo(),
	}, nil
}

// StreamCompletion streams the response from the AI service.
func (p *GeminiProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	model := p.client.GenerativeModel(p.modelName)
	if req.Model != "" {
		model = p.client.GenerativeModel(req.Model)
	}

	// Configuration
	if req.Temperature != 0 {
		model.SetTemperature(float32(req.Temperature))
	}

	var parts []genai.Part
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
		}
	}

	parts = append(parts, genai.Text(req.UserPrompt))
	if len(req.ImageData) > 0 {
		parts = append(parts, genai.ImageData("jpeg", req.ImageData))
	}

	iter := model.GenerateContentStream(ctx, parts...)
	respChan := make(chan StreamResponse)

	go func() {
		defer close(respChan)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				respChan <- StreamResponse{Done: true}
				return
			}
			if err != nil {
				respChan <- StreamResponse{Error: p.mapError(err)}
				return
			}

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					respChan <- StreamResponse{Content: string(txt)}
				}
			}
		}
	}()

	return respChan, nil
}

func (p *GeminiProvider) mapError(err error) error {
	zap.S().Error("Gemini API error", "error", err, "model", p.modelName)
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		if gerr.Code == 429 {
			return NewError(ErrCodeRateLimit, "gemini rate limit exceeded", err)
		}
		if gerr.Code == 404 || gerr.Code == 400 {
			return NewError(ErrCodeInvalidRequest, "gemini invalid request or model not found", err)
		}
		if gerr.Code >= 500 {
			return NewError(ErrCodeProviderOffline, "gemini service unavailable", err)
		}
	}
	return NewError(ErrCodeInternal, "gemini internal error", err)
}

// GetModelInfo returns information about the model.
func (p *GeminiProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Gemini",
		ModelName:    p.modelName,
	}
}

// CheckHealth verifies if the provider is currently operational.
func (p *GeminiProvider) CheckHealth(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("gemini client is nil")
	}
	model := p.client.GenerativeModel(p.modelName)
	_, err := model.GenerateContent(ctx, genai.Text("ping"))
	return err
}
