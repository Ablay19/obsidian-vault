package doppler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client represents a Doppler CLI client
type Client struct {
	project string
	config  string
	token   string
	timeout time.Duration
}

// NewClient creates a new Doppler client
func NewClient(project, config string) *Client {
	return &Client{
		project: project,
		config:  config,
		token:   os.Getenv("DOPPLER_TOKEN"),
		timeout: 30 * time.Second,
	}
}

// WithToken sets the service token
func (c *Client) WithToken(token string) *Client {
	c.token = token
	return c
}

// WithTimeout sets the command timeout
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.timeout = timeout
	return c
}

// GetSecret retrieves a single secret
func (c *Client) GetSecret(ctx context.Context, key string) (string, error) {
	args := []string{"secrets", "get", key, "--project", c.project, "--config", c.config, "--json"}

	if c.token != "" {
		args = append([]string{"--token", c.token}, args...)
	}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("doppler command failed: %w, output: %s", err, string(output))
	}

	var result struct {
		Value string `json:"value"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse doppler response: %w", err)
	}

	return result.Value, nil
}

// GetSecrets retrieves all secrets for the project/config
func (c *Client) GetSecrets(ctx context.Context) (map[string]string, error) {
	args := []string{"secrets", "download", "--project", c.project, "--config", c.config, "--json"}

	if c.token != "" {
		args = append([]string{"--token", c.token}, args...)
	}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("doppler command failed: %w, output: %s", err, string(output))
	}

	secrets := make(map[string]string)
	if err := json.Unmarshal(output, &secrets); err != nil {
		return nil, fmt.Errorf("failed to parse doppler response: %w", err)
	}

	return secrets, nil
}

// Run executes a command with Doppler environment variables
func (c *Client) Run(ctx context.Context, command []string) error {
	args := []string{"run", "--project", c.project, "--config", c.config, "--command", strings.Join(command, " ")}

	if c.token != "" {
		args = append([]string{"--token", c.token}, args...)
	}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// IsAvailable checks if Doppler CLI is available and authenticated
func (c *Client) IsAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "doppler", "me")
	if c.token != "" {
		cmd.Env = append(os.Environ(), "DOPPLER_TOKEN="+c.token)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("doppler not available or not authenticated: %w", err)
	}

	return nil
}
