package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ReplicateProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
	baseURL    string
}

type replicateRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

type replicateResponse struct {
	Output []string `json:"output"`
	Error  string   `json:"error,omitempty"`
}

func NewReplicateProvider(apiKey, model string) *ReplicateProvider {
	if apiKey == "" {
		return nil
	}
	return &ReplicateProvider{
		apiKey:    apiKey,
		modelName: model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Replicate can be slower
		},
		baseURL: "https://api.replicate.com/v1",
	}
}

func (p *ReplicateProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	// For simplicity, use a free Llama model
	version := "meta/llama-2-7b-chat:13c3cdee13ee059ab779f0291d29054dab00a47dad8261375654de5540165fb46"

	input := map[string]interface{}{
		"prompt":         req.UserPrompt,
		"system_prompt":  req.SystemPrompt,
		"max_new_tokens": 500,
		"temperature":    0.7,
	}

	reqBody, err := json.Marshal(replicateRequest{
		Version: version,
		Input:   input,
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/predictions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Token "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, NewError(ErrCodeNetworkError, "replicate network error", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, p.mapError(resp.StatusCode, string(body))
	}

	var predResp struct {
		URLs struct {
			Get string `json:"get"`
		} `json:"urls"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, err
	}

	// Poll for completion (simplified - in production, handle async properly)
	time.Sleep(2 * time.Second) // Wait for completion

	getResp, err := http.Get(predResp.URLs.Get)
	if err != nil {
		return nil, err
	}
	defer getResp.Body.Close()

	var result replicateResponse
	if err := json.NewDecoder(getResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	content := ""
	if len(result.Output) > 0 {
		content = strings.Join(result.Output, "")
	}

	return &ResponseModel{
		Content: content,
		ProviderInfo: ModelInfo{
			ProviderName: "replicate",
			ModelName:    p.modelName,
		},
	}, nil
}

func (p *ReplicateProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	// For free tier, fall back to non-streaming
	resp, err := p.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- StreamResponse{
			Content: resp.Content,
			Done:    true,
		}
	}()
	return ch, nil
}

func (p *ReplicateProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "replicate",
		ModelName:    p.modelName,
	}
}

func (p *ReplicateProvider) CheckHealth(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	req.Header.Set("Authorization", "Token "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("replicate health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *ReplicateProvider) mapError(status int, body string) error {
	switch status {
	case 401:
		return NewError(ErrCodeUnauthorized, "replicate authentication failed", fmt.Errorf("status %d: %s", status, body))
	case 402:
		return NewError(ErrCodeQuotaExceeded, "replicate quota exceeded", fmt.Errorf("status %d: %s", status, body))
	case 429:
		return NewError(ErrCodeRateLimit, "replicate rate limit exceeded", fmt.Errorf("status %d: %s", status, body))
	case 400, 422:
		return NewError(ErrCodeInvalidRequest, "replicate invalid request", fmt.Errorf("status %d: %s", status, body))
	case 500, 502, 503:
		return NewError(ErrCodeProviderOffline, "replicate service unavailable", fmt.Errorf("status %d: %s", status, body))
	default:
		return NewError(ErrCodeInternal, "replicate internal error", fmt.Errorf("status %d: %s", status, body))
	}
}
