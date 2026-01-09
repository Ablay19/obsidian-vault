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
)

type TogetherProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
	baseURL    string
}

type togetherMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type togetherRequest struct {
	Model    string            `json:"model"`
	Messages []togetherMessage `json:"messages"`
	Stream   bool              `json:"stream,omitempty"`
}

type togetherResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type togetherStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func NewTogetherProvider(apiKey, model string) *TogetherProvider {
	if apiKey == "" {
		return nil
	}
	return &TogetherProvider{
		apiKey:    apiKey,
		modelName: model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.together.xyz/v1",
	}
}

func (p *TogetherProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	messages := []togetherMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, togetherMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, togetherMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(togetherRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, NewError(ErrCodeNetworkError, "together network error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, p.mapError(resp.StatusCode, string(body))
	}

	var chatResp togetherResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	}

	return &ResponseModel{
		Content: content,
		ProviderInfo: ModelInfo{
			ProviderName: "together",
			ModelName:    p.modelName,
		},
	}, nil
}

func (p *TogetherProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	messages := []togetherMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, togetherMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, togetherMessage{Role: "user", Content: req.UserPrompt})

	reqBody, err := json.Marshal(togetherRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, NewError(ErrCodeNetworkError, "together network error", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
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

			var streamResp togetherStreamResponse
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

func (p *TogetherProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "together",
		ModelName:    p.modelName,
	}
}

func (p *TogetherProvider) CheckHealth(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("together health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *TogetherProvider) mapError(status int, body string) error {
	switch status {
	case 401:
		return NewError(ErrCodeUnauthorized, "together authentication failed", fmt.Errorf("status %d: %s", status, body))
	case 402:
		return NewError(ErrCodeQuotaExceeded, "together quota exceeded", fmt.Errorf("status %d: %s", status, body))
	case 429:
		return NewError(ErrCodeRateLimit, "together rate limit exceeded", fmt.Errorf("status %d: %s", status, body))
	case 400, 422:
		return NewError(ErrCodeInvalidRequest, "together invalid request", fmt.Errorf("status %d: %s", status, body))
	case 408:
		return NewError(ErrCodeTimeout, "together request timeout", fmt.Errorf("status %d: %s", status, body))
	case 500, 502, 503:
		return NewError(ErrCodeProviderOffline, "together service unavailable", fmt.Errorf("status %d: %s", status, body))
	default:
		return NewError(ErrCodeInternal, "together internal error", fmt.Errorf("status %d: %s", status, body))
	}
}
