package communication

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HttpClient wraps HTTP client functionality
type HttpClient struct {
	baseURL string
	client  *http.Client
	logger  *slog.Logger
}

// RequestConfig represents HTTP request configuration
type RequestConfig struct {
	Method   string
	Endpoint string
	Body     []byte
	Headers  map[string]string
	Timeout  time.Duration
}

// Response represents HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

// NewHttpClient creates a new HTTP client
func NewHttpClient(baseURL string, logger *slog.Logger) *HttpClient {
	return &HttpClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		logger:  logger,
	}
}

// Do performs an HTTP request
func (c *HttpClient) Do(config RequestConfig) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	url := c.baseURL + config.Endpoint
	req, err := http.NewRequestWithContext(ctx, config.Method, url, bytes.NewReader(config.Body))
	if err != nil {
		return nil, err
	}

	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}
