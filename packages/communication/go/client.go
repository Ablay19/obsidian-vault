package communication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

type HttpClient struct {
	client     *http.Client
	baseURL    string
	logger     *slog.Logger
	timeout    time.Duration
	maxRetries int
}

type RequestConfig struct {
	Method   string
	Endpoint string
	Headers  map[string]string
	Body     interface{}
	Timeout  time.Duration
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

func NewHttpClient(baseURL string, logger *slog.Logger) *HttpClient {
	return &HttpClient{
		client:     &http.Client{},
		baseURL:    baseURL,
		logger:     logger,
		timeout:    30 * time.Second,
		maxRetries: 0, // Fail-fast: no retries
	}
}

func (c *HttpClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.client.Timeout = timeout
}

func (c *HttpClient) Do(config RequestConfig) (*types.APIResponse, error) {
	startTime := time.Now()

	url := c.baseURL + config.Endpoint
	types.LogDebug(c.logger, "Making HTTP request", "method", config.Method, "url", url, "timeout", config.Timeout)

	var body io.Reader
	if config.Body != nil {
		jsonBody, err := json.Marshal(config.Body)
		if err != nil {
			types.LogError(c.logger, err, "Failed to marshal request body")
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(config.Method, url, body)
	if err != nil {
		types.LogError(c.logger, err, "Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		types.LogError(c.logger, err, "HTTP request failed", "url", url)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	types.LogInfo(c.logger, "HTTP request completed",
		"method", config.Method,
		"url", url,
		"status", resp.StatusCode,
		"duration_ms", duration.Milliseconds(),
	)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		types.LogError(c.logger, err, "Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Fail-fast: return error immediately on non-2xx status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errResp := &types.ErrorResponse{}
		if err := json.Unmarshal(respBody, errResp); err != nil {
			types.LogError(c.logger, err, "Failed to unmarshal error response", "status", resp.StatusCode)
			return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
		}
		types.LogError(c.logger, fmt.Errorf(errResp.Error), "HTTP request failed", "status", resp.StatusCode, "url", url)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, errResp.Message)
	}

	apiResp := &types.APIResponse{}
	if err := json.Unmarshal(respBody, apiResp); err != nil {
		types.LogError(c.logger, err, "Failed to unmarshal response", "body", string(respBody))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	types.LogDebug(c.logger, "Response parsed successfully", "status", apiResp.Status)
	return apiResp, nil
}

func (c *HttpClient) Get(endpoint string, headers map[string]string) (*types.APIResponse, error) {
	return c.Do(RequestConfig{
		Method:   "GET",
		Endpoint: endpoint,
		Headers:  headers,
		Timeout:  c.timeout,
	})
}

func (c *HttpClient) Post(endpoint string, body interface{}, headers map[string]string) (*types.APIResponse, error) {
	return c.Do(RequestConfig{
		Method:   "POST",
		Endpoint: endpoint,
		Headers:  headers,
		Body:     body,
		Timeout:  c.timeout,
	})
}

func (c *HttpClient) Put(endpoint string, body interface{}, headers map[string]string) (*types.APIResponse, error) {
	return c.Do(RequestConfig{
		Method:   "PUT",
		Endpoint: endpoint,
		Headers:  headers,
		Body:     body,
		Timeout:  c.timeout,
	})
}

func (c *HttpClient) Delete(endpoint string, headers map[string]string) (*types.APIResponse, error) {
	return c.Do(RequestConfig{
		Method:   "DELETE",
		Endpoint: endpoint,
		Headers:  headers,
		Timeout:  c.timeout,
	})
}

func (c *HttpClient) HealthCheck(endpoint string) error {
	types.LogInfo(c.logger, "Performing health check", "endpoint", endpoint)

	startTime := time.Now()
	resp, err := c.client.Get(c.baseURL + endpoint)
	if err != nil {
		types.LogError(c.logger, err, "Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		types.LogError(c.logger, fmt.Errorf("unexpected status code"), "status", resp.StatusCode)
		return fmt.Errorf("health check failed: unexpected status code %d", resp.StatusCode)
	}

	types.LogInfo(c.logger, "Health check passed", "duration_ms", duration.Milliseconds())
	return nil
}

func (c *HttpClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}
