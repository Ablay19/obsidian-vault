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

	"github.com/spf13/viper"
)

// HuggingFaceProvider implements the AIProvider interface for Hugging Face Router (OpenAI-compatible).
type HuggingFaceProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

// NewHuggingFaceProvider creates a new HuggingFaceProvider.
func NewHuggingFaceProvider(apiKey, model string) *HuggingFaceProvider {
	if apiKey == "" {
		return nil
	}
	return &HuggingFaceProvider{
		apiKey:    apiKey,
		modelName: model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type hfMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hfChatRequest struct {
	Model    string      `json:"model"`
	Messages []hfMessage `json:"messages"`
	Stream   bool        `json:"stream,omitempty"`
}

type hfChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error interface{} `json:"error,omitempty"`
}

type hfStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// GenerateCompletion sends a request to the Hugging Face service and returns a complete response.
func (p *HuggingFaceProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	messages := []hfMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, hfMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, hfMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(hfChatRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", p.getBaseURL())
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, p.mapError(resp.StatusCode, string(body))
	}

	var chatResp hfChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
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
func (p *HuggingFaceProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	messages := []hfMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, hfMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, hfMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(hfChatRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", p.getBaseURL())
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	respChan := make(chan StreamResponse)
	go func() {
		defer resp.Body.Close()
		defer close(respChan)

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			respChan <- StreamResponse{Error: p.mapError(resp.StatusCode, string(body))}
			return
		}

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

			var streamResp hfStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
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

func (p *HuggingFaceProvider) getBaseURL() string {
	baseURL := viper.GetString("HUGGINGFACE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://router.huggingface.co/v1"
	}
	return strings.TrimSuffix(baseURL, "/")
}

// GetModelInfo returns information about the model.
func (p *HuggingFaceProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Hugging Face",
		ModelName:    p.modelName,
	}
}

// CheckHealth verifies if the provider is currently operational.
func (p *HuggingFaceProvider) CheckHealth(ctx context.Context) error {
	_, err := p.GenerateCompletion(ctx, &RequestModel{UserPrompt: "ping"})
	return err
}

func (p *HuggingFaceProvider) mapError(status int, body string) error {
	if status == 429 {
		return NewError(ErrCodeRateLimit, "hugging face rate limit exceeded", fmt.Errorf("status %d: %s", status, body))
	}
	if status == 404 || status == 400 {
		return NewError(ErrCodeInvalidRequest, "hugging face invalid request or model not found", fmt.Errorf("status %d: %s", status, body))
	}
	if status >= 500 {
		return NewError(ErrCodeProviderOffline, "hugging face service unavailable", fmt.Errorf("status %d: %s", status, body))
	}
	return NewError(ErrCodeInternal, "hugging face internal error", fmt.Errorf("status %d: %s", status, body))
}
