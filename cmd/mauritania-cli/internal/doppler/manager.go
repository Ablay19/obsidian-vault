package doppler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Manager handles environment variable loading and management
type Manager struct {
	client    *Client
	fallbacks map[string]string
	cache     map[string]string
	cacheMux  sync.RWMutex
}

// NewManager creates a new environment variable manager
func NewManager(project, config string) *Manager {
	return &Manager{
		client:    NewClient(project, config),
		fallbacks: make(map[string]string),
		cache:     make(map[string]string),
	}
}

// WithFallbacks sets fallback values for when Doppler is unavailable
func (m *Manager) WithFallbacks(fallbacks map[string]string) *Manager {
	for k, v := range fallbacks {
		m.fallbacks[k] = v
	}
	return m
}

// LoadSecret loads a single secret with caching and fallback
func (m *Manager) LoadSecret(ctx context.Context, key string) (string, error) {
	// Check cache first
	m.cacheMux.RLock()
	if value, exists := m.cache[key]; exists {
		m.cacheMux.RUnlock()
		return value, nil
	}
	m.cacheMux.RUnlock()

	// Try Doppler
	if value, err := m.client.GetSecret(ctx, key); err == nil {
		m.cacheMux.Lock()
		m.cache[key] = value
		m.cacheMux.Unlock()
		return value, nil
	}

	// Try environment variable
	if value := os.Getenv(key); value != "" {
		m.cacheMux.Lock()
		m.cache[key] = value
		m.cacheMux.Unlock()
		return value, nil
	}

	// Try fallback
	if value, exists := m.fallbacks[key]; exists {
		m.cacheMux.Lock()
		m.cache[key] = value
		m.cacheMux.Unlock()
		return value, nil
	}

	return "", fmt.Errorf("secret %s not found in Doppler, environment, or fallbacks", key)
}

// LoadAllSecrets loads all secrets for the project/config
func (m *Manager) LoadAllSecrets(ctx context.Context) (map[string]string, error) {
	// Try Doppler first
	if secrets, err := m.client.GetSecrets(ctx); err == nil {
		m.cacheMux.Lock()
		for k, v := range secrets {
			m.cache[k] = v
		}
		m.cacheMux.Unlock()
		return secrets, nil
	}

	// Fallback to environment + fallbacks
	secrets := make(map[string]string)

	// Add environment variables
	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx > 0 {
			key := env[:idx]
			value := env[idx+1:]
			secrets[key] = value
		}
	}

	// Add fallbacks
	for k, v := range m.fallbacks {
		if _, exists := secrets[k]; !exists {
			secrets[k] = v
		}
	}

	if len(secrets) == 0 {
		return nil, fmt.Errorf("no secrets available from Doppler, environment, or fallbacks")
	}

	m.cacheMux.Lock()
	for k, v := range secrets {
		m.cache[k] = v
	}
	m.cacheMux.Unlock()

	return secrets, nil
}

// SetEnvironment applies secrets to the current process environment
func (m *Manager) SetEnvironment(ctx context.Context) error {
	secrets, err := m.LoadAllSecrets(ctx)
	if err != nil {
		return err
	}

	for key, value := range secrets {
		os.Setenv(key, value)
	}

	return nil
}

// ClearCache clears the secret cache
func (m *Manager) ClearCache() {
	m.cacheMux.Lock()
	m.cache = make(map[string]string)
	m.cacheMux.Unlock()
}

// IsAvailable checks if Doppler is available
func (m *Manager) IsAvailable(ctx context.Context) bool {
	return m.client.IsAvailable(ctx) == nil
}
