package state

import (
	"time"
)

// RuntimeConfig holds the authoritative operational state of the bot.
type RuntimeConfig struct {
	AIEnabled   bool                     `json:"ai_enabled"`
	Providers   map[string]ProviderState `json:"providers"`
	APIKeys     map[string]APIKeyState   `json:"api_keys"`
	Environment EnvironmentState         `json:"environment"`
	// Add other global flags/settings here
}

// ProviderState defines the state for an individual AI provider.
type ProviderState struct {
	Name          string `json:"name"`
	ModelName     string `json:"model_name"`     // The model name used by this provider (e.g., gemini-pro)
	Enabled       bool   `json:"enabled"`        // Global enable/disable for this provider
	Paused        bool   `json:"paused"`         // Temporarily paused
	Blocked       bool   `json:"blocked"`        // Hard-stopped/blocked with reason
	BlockedReason string `json:"blocked_reason"` // Reason for blocking
	// Add other provider-specific settings here
}

// APIKeyState defines the state for an individual API key.
type APIKeyState struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"` // Which provider this key belongs to
	Value          string    `json:"-"`        // The actual API key (sensitive, omit from JSON)
	Enabled        bool      `json:"enabled"`  // Enable/disable individual key
	Blocked        bool      `json:"blocked"`  // If the key itself is blocked (e.g., invalid, quota)
	BlockedReason  string    `json:"blocked_reason"`
	LastError      string    `json:"last_error"`
	QuotaRemaining int       `json:"quota_remaining"` // Estimated quota (if available)
	LastUsedAt     time.Time `json:"last_used_at"`
	IsDefault      bool      `json:"is_default"` // Was this key initially from .env?
	RotatedCount   int       `json:"rotated_count"`
	// Add other key-specific settings here
}

// EnvironmentState defines the bot's operational environment.
type EnvironmentState struct {
	Mode             string `json:"mode"`              // "dev" | "prod"
	BackendHost      string `json:"backend_host"`      // Base URL for backend API calls
	IsolationEnabled bool   `json:"isolation_enabled"` // If true, strictly prevents cross-env requests
	// Add other environment-specific settings here
}
