package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GitHubProvider implements AIProvider for GitHub Models
type GitHubProvider struct {
	token   string
	baseURL string
	client  *http.Client
	org     string // optional organization
}

// GitHubRequest represents the request to GitHub Models API
type GitHubRequest struct {
	Messages    []GitHubMessage `json:"messages"`
	Model       string          `json:"model"`
	Stream      bool            `json:"stream,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
}

// GitHubMessage represents a message in the conversation
type GitHubMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GitHubResponse represents the response from GitHub Models API
type GitHubResponse struct {
	Choices []GitHubChoice `json:"choices"`
	Usage   GitHubUsage    `json:"usage"`
}

// GitHubChoice represents a choice in the response
type GitHubChoice struct {
	Message      GitHubMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// GitHubUsage represents token usage
type GitHubUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewGitHubProvider creates a new GitHub Models provider
func NewGitHubProvider(token string, org string) *GitHubProvider {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &GitHubProvider{
		token:   token,
		baseURL: "https://models.github.ai",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		org: org,
	}
}

// GenerateCompletion generates a completion using GitHub Models
func (p *GitHubProvider) GenerateCompletion(ctx context.Context, req *RequestModel) (*ResponseModel, error) {
	// Convert request to GitHub format
	ghReq := GitHubRequest{
		Messages: []GitHubMessage{
			{Role: "user", Content: req.Prompt},
		},
		Model: req.Model,
	}

	if req.MaxTokens > 0 {
		ghReq.MaxTokens = req.MaxTokens
	}

	// Extract options
	if req.Options != nil {
		if temp, ok := req.Options["temperature"].(float64); ok {
			ghReq.Temperature = temp
		}
		if topP, ok := req.Options["top_p"].(float64); ok {
			ghReq.TopP = topP
		}
	}

	// Make API call
	var resp *http.Response
	var err error

	if p.org != "" {
		resp, err = p.makeRequest(ctx, "POST", fmt.Sprintf("/orgs/%s/inference", p.org), ghReq)
	} else {
		resp, err = p.makeRequest(ctx, "POST", "/inference", ghReq)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var ghResp GitHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(ghResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ResponseModel{
		Content: ghResp.Choices[0].Message.Content,
		Model:   req.Model,
		Tokens:  ghResp.Usage.TotalTokens,
	}, nil
}

// StreamCompletion streams completion (not implemented yet)
func (p *GitHubProvider) StreamCompletion(ctx context.Context, req *RequestModel) (<-chan StreamResponse, error) {
	// TODO: Implement streaming
	return nil, fmt.Errorf("streaming not implemented")
}

// CheckHealth checks if the provider is healthy
func (p *GitHubProvider) CheckHealth(ctx context.Context) error {
	// Simple health check by listing models
	resp, err := p.makeRequest(ctx, "GET", "/catalog/models", nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check API error: %d", resp.StatusCode)
	}

	return nil
}

// GetModelInfo returns model information
func (p *GitHubProvider) GetModelInfo() ModelInfo {
	// Return default model info, could be enhanced to fetch from catalog
	return ModelInfo{
		Name:           "github-models-default",
		Provider:       "github",
		MaxTokens:      4096,
		SupportsStream: false, // TODO: implement streaming
	}
}

// makeRequest makes an HTTP request to GitHub Models API
func (p *GitHubProvider) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := p.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	return p.client.Do(req)
}
