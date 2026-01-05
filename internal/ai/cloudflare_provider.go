package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CloudflareProvider implements AIProvider interface for Cloudflare Workers
type CloudflareProvider struct {
	endpoint string
	client   *http.Client
	model    string
}

type CloudflareRequest struct {
	Prompt string `json:"prompt"`
}

type CloudflareResponse struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

// NewCloudflareProvider creates a new Cloudflare Worker provider
func NewCloudflareProvider(workerURL string) *CloudflareProvider {
	return &CloudflareProvider{
		endpoint: workerURL + "/ai/proxy/cloudflare",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "@cf/meta/llama-3-8b-instruct",
	}
}

// GenerateCompletion implements AIProvider interface
func (p *CloudflareProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	// Combine system and user prompts
	fullPrompt := req.SystemPrompt + "\n" + req.UserPrompt

	// Prepare request
	cfReq := CloudflareRequest{
		Prompt: fullPrompt,
	}

	reqBody, err := json.Marshal(cfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "text/plain")
	httpReq.Header.Set("User-Agent", "ObsidianBot/1.0")

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Cloudflare Worker error: %d %s", resp.StatusCode, string(body))
	}

	// Parse response
	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !cfResp.Success && cfResp.Response == "" {
		cfResp.Success = true // Worker returns success implicitly
	}

	return &ResponseModel{
		Content: cfResp.Response,
		Usage: TokenUsage{
			InputTokens:  estimateTokens(fullPrompt),
			OutputTokens: estimateTokens(cfResp.Response),
			TotalTokens:  estimateTokens(fullPrompt) + estimateTokens(cfResp.Response),
		},
		Cached: false, // Worker doesn't provide cache info
		ProviderInfo: ModelInfo{
			ProviderName: "cloudflare",
			ModelName:    p.model,
			KeyID:        "",
			Enabled:      true,
			Blocked:      false,
		},
	}, nil
}

// StreamCompletion implements AIProvider interface (not supported by worker yet)
func (p *CloudflareProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	// For now, implement as regular response
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

// GetModelInfo returns information about the model
func (p *CloudflareProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "cloudflare",
		ModelName:    p.model,
		KeyID:        "",
		Enabled:      true,
		Blocked:      false,
	}
}

// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
func estimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// CheckHealth verifies if the provider is currently operational
func (p *CloudflareProvider) CheckHealth(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/../health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}
