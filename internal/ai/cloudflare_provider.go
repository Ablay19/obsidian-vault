package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obsidian-automation/internal/telemetry"
	"time"
)

// CloudflareProvider implements AIProvider interface for Cloudflare Workers
type CloudflareProvider struct {
	endpoint string
	client   *http.Client
	model    string
	provider string
}

type CloudflareRequest struct {
	Prompt string `json:"prompt"`
}

type CloudflareResponse struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
	Error    string `json:"error"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// NewCloudflareProvider creates a new Cloudflare Worker provider
func NewCloudflareProvider(workerURL string) *CloudflareProvider {
	return &CloudflareProvider{
		endpoint: workerURL + "/ai",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		model:    "worker-proxy",
		provider: "cloudflare", // default
	}
}

// NewCloudflareProviderWithProvider creates a provider for a specific provider via worker
func NewCloudflareProviderWithProvider(workerURL, provider string) *CloudflareProvider {
	return &CloudflareProvider{
		endpoint: workerURL + "/ai",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		model:    provider + "-via-worker",
		provider: provider,
	}
}

// GenerateCompletion implements AIProvider interface
func (p *CloudflareProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	// Combine system and user prompts
	fullPrompt := req.SystemPrompt + "\n" + req.UserPrompt

	// Prepare request as JSON
	requestBody := map[string]interface{}{
		"prompt":   fullPrompt,
		"provider": p.provider,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	telemetry.Info("Cloudflare response status: " + fmt.Sprintf("%d", resp.StatusCode) + ", body length: " + fmt.Sprintf("%d", len(body)))

	// Check status
	if resp.StatusCode != http.StatusOK {
		telemetry.Error("Cloudflare Worker error: " + fmt.Sprintf("%d", resp.StatusCode))
		return nil, fmt.Errorf("Cloudflare Worker error: %d %s", resp.StatusCode, string(body))
	}

	// Log first 100 chars of response for debugging
	telemetry.Info("Cloudflare response preview: " + string(body[:min(100, len(body))]))

	telemetry.Info("Cloudflare response status: " + fmt.Sprintf("%d", resp.StatusCode) + ", body length: " + fmt.Sprintf("%d", len(body)))

	// Check status
	if resp.StatusCode != http.StatusOK {
		telemetry.Error("Cloudflare Worker error: " + fmt.Sprintf("%d", resp.StatusCode) + " Body: " + string(body[:min(200, len(body))]))
		return nil, fmt.Errorf("Cloudflare Worker error: %d %s", resp.StatusCode, string(body))
	}

	// Log first 100 chars of response for debugging
	responseStr := string(body)
	telemetry.Info("Cloudflare response preview: " + responseStr[:min(100, len(responseStr))])

	// Parse response
	var cfResp CloudflareResponse
	if err := json.Unmarshal(body, &cfResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !cfResp.Success {
		if cfResp.Error != "" {
			return nil, fmt.Errorf("worker error: %s", cfResp.Error)
		}
		return nil, fmt.Errorf("worker returned unsuccessful response")
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
			ProviderName: cfResp.Provider,
			ModelName:    cfResp.Model,
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
