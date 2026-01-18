package doppler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// ServiceTokenManager handles Doppler service tokens
type ServiceTokenManager struct {
	project string
	token   string
}

// NewServiceTokenManager creates a new service token manager
func NewServiceTokenManager(project string) *ServiceTokenManager {
	return &ServiceTokenManager{
		project: project,
		token:   getDopplerToken(),
	}
}

// CreateToken creates a new service token for a config
func (stm *ServiceTokenManager) CreateToken(ctx context.Context, name, config string) (string, error) {
	args := []string{"service-tokens", "create", name,
		"--project", stm.project, "--config", config, "--json"}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = append(os.Environ(), "DOPPLER_TOKEN="+stm.token)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create service token: %w, output: %s", err, string(output))
	}

	var result struct {
		Key string `json:"key"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return result.Key, nil
}

// ListTokens lists all service tokens for the project
func (stm *ServiceTokenManager) ListTokens(ctx context.Context) ([]ServiceTokenInfo, error) {
	args := []string{"service-tokens", "--project", stm.project, "--json"}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = append(os.Environ(), "DOPPLER_TOKEN="+stm.token)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list service tokens: %w, output: %s", err, string(output))
	}

	var tokens []ServiceTokenInfo
	if err := json.Unmarshal(output, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse tokens response: %w", err)
	}

	return tokens, nil
}

// DeleteToken deletes a service token
func (stm *ServiceTokenManager) DeleteToken(ctx context.Context, slug string) error {
	args := []string{"service-tokens", "delete", slug, "--project", stm.project, "--yes"}

	cmd := exec.CommandContext(ctx, "doppler", args...)
	cmd.Env = append(os.Environ(), "DOPPLER_TOKEN="+stm.token)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete service token: %w, output: %s", err, string(output))
	}

	return nil
}

// ValidateToken checks if a service token is valid
func (stm *ServiceTokenManager) ValidateToken(ctx context.Context, token string) error {
	// Try to use the token to access secrets
	testClient := NewClient(stm.project, "dev").WithToken(token)
	return testClient.IsAvailable(ctx)
}

// ServiceTokenInfo represents service token information
type ServiceTokenInfo struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Project    string `json:"project"`
	Config     string `json:"config"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
	Access     string `json:"access"`
}

// getDopplerToken retrieves the Doppler token from environment
func getDopplerToken() string {
	// Check common environment variables
	envVars := []string{"DOPPLER_TOKEN", "DOPPLER_SERVICE_TOKEN"}

	for _, envVar := range envVars {
		if token := os.Getenv(envVar); token != "" {
			return token
		}
	}

	return ""
}

// GenerateTokenName generates a standardized token name
func GenerateTokenName(prefix, environment string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, environment, generateSlug())
}

// generateSlug creates a simple slug
func generateSlug() string {
	// In a real implementation, this would generate a unique slug
	// For now, return a placeholder
	return "auto"
}

// SetupCIEnvironment sets up environment for CI/CD
func SetupCIEnvironment(project, config, token string) error {
	envVars := map[string]string{
		"DOPPLER_TOKEN":       token,
		"DOPPLER_PROJECT":     project,
		"DOPPLER_CONFIG":      config,
		"DOPPLER_ENVIRONMENT": config,
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}
