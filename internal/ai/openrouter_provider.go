package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OpenRouterProvider implements the AIProvider interface for OpenRouter.
type OpenRouterProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider.
func NewOpenRouterProvider(apiKey string, modelName string, httpClient *http.Client) *OpenRouterProvider {
	if apiKey == "" {
		zap.S().Info("OpenRouter API key is empty. OpenRouter AI will be unavailable.")
		return nil
	}
	
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	return &OpenRouterProvider{
		apiKey:     apiKey,
		modelName:  modelName,
		httpClient: httpClient,
	}
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatRequest struct {
	Model          string              `json:"model"`
	Messages       []openRouterMessage `json:"messages"`
	Stream         bool                `json:"stream,omitempty"`
	Temperature    float64             `json:"temperature,omitempty"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type openRouterChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type openRouterStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// GenerateCompletion sends a request to the AI service and returns a complete response.
func (p *OpenRouterProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	messages := []openRouterMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, openRouterMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(openRouterChatRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/obsidian-automation-bot")
	httpReq.Header.Set("X-Title", "Obsidian Automation Bot")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, p.mapError(resp.StatusCode, string(body))
	}

	var chatResp openRouterChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("openrouter api error: %s", chatResp.Error.Message)
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	}

	return &ResponseModel{
		Content:      content,
		ProviderInfo: p.GetModelInfo(),
	}, nil
}

// StreamCompletion streams the response from the AI service.
func (p *OpenRouterProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	messages := []openRouterMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, openRouterMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(openRouterChatRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/obsidian-automation-bot")
	httpReq.Header.Set("X-Title", "Obsidian Automation Bot")

	// Execute request to get response headers and body reader
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	// Check status immediately
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, p.mapError(resp.StatusCode, string(body))
	}

	respChan := make(chan StreamResponse)

	go func() {
		defer resp.Body.Close()
		defer close(respChan)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					respChan <- StreamResponse{Done: true}
					break
				}
				respChan <- StreamResponse{Error: err}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				respChan <- StreamResponse{Done: true}
				break
			}

			var streamResp openRouterStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				// Don't kill stream on one bad line, just log
				zap.S().Debug("Error unmarshaling OpenRouter stream line", "error", err)
				continue
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				respChan <- StreamResponse{Content: content}
			}
		}
	}()

	return respChan, nil
}

// GetModelInfo returns information about the model.
func (p *OpenRouterProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "OpenRouter",
		ModelName:    p.modelName,
	}
}

// CheckHealth verifies if the provider is currently operational.
func (p *OpenRouterProvider) CheckHealth(ctx context.Context) error {
	_, err := p.GenerateCompletion(ctx, &RequestModel{UserPrompt: "ping"})
	return err
}

func (p *OpenRouterProvider) mapError(status int, body string) error {
	if status == 429 {
		return NewError(ErrCodeRateLimit, "openrouter rate limit exceeded", fmt.Errorf("status %d: %s", status, body))
	}
	if status == 404 || status == 400 {
		return NewError(ErrCodeInvalidRequest, "openrouter invalid request or model not found", fmt.Errorf("status %d: %s", status, body))
	}
	if status == 401 {
		return NewError(ErrCodeUnauthorized, "openrouter unauthorized", fmt.Errorf("status %d: %s", status, body))
	}
	if status >= 500 {
		return NewError(ErrCodeProviderOffline, "openrouter service unavailable", fmt.Errorf("status %d: %s", status, body))
	}
	return NewError(ErrCodeInternal, "openrouter internal error", fmt.Errorf("status %d: %s", status, body))
}