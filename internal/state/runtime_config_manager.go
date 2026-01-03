package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// RuntimeConfigManager is the central authority for the bot's runtime configuration.
type RuntimeConfigManager struct {
	mu     sync.RWMutex
	config RuntimeConfig
	db     *sql.DB
	// Add other dependencies here if needed (e.g., a logger)
}

// NewRuntimeConfigManager creates and initializes a new RuntimeConfigManager.
func NewRuntimeConfigManager(db *sql.DB) (*RuntimeConfigManager, error) {
	rcm := &RuntimeConfigManager{
		db: db,
		config: RuntimeConfig{
			AIEnabled: true, // Default to AI enabled
			Providers: make(map[string]ProviderState),
			APIKeys:   make(map[string]APIKeyState),
			Environment: EnvironmentState{
				Mode:             "prod", // Default to prod
				IsolationEnabled: true,
			},
		},
	}

	// Load initial state from DB
	if err := rcm.LoadStateFromDB(); err != nil {
		slog.Warn("Could not load state from DB, initializing with defaults", "error", err)
		// If DB load fails, initialize from .env (bootstrap)
		rcm.initializeFromEnv()
	} else {
		slog.Info("State loaded from DB successfully. Merging keys from environment.")
		rcm.mergeKeysFromEnv()
	}

	// Persist the current state after initialization (either from DB or env)
	if err := rcm.PersistStateToDB(); err != nil {
		return nil, fmt.Errorf("failed to persist initial state to DB: %w", err)
	}

	return rcm, nil
}

// mergeKeysFromEnv ensures that keys provided via environment variables are injected
// into the runtime configuration, overwriting or supplementing what was loaded from the DB.
// This is necessary because keys are not persisted to the DB for security.
func (rcm *RuntimeConfigManager) mergeKeysFromEnv() {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	// Allow environment variable to override global AI enabled status
	if viper.IsSet("AI_ENABLED") {
		rcm.config.AIEnabled = viper.GetBool("AI_ENABLED")
		slog.Info("AI_ENABLED status overridden by environment variable.", "enabled", rcm.config.AIEnabled)
	}

	// Re-initialize provider states from env/config to ensure enabled status is correct based on current env
	rcm.initializeProviderStates()

	// Helper to merge a single key
	mergeKey := func(provider string, index int, value string) {
		id := generateKeyID(provider, index)
		if ks, ok := rcm.config.APIKeys[id]; ok {
			ks.Value = value
			rcm.config.APIKeys[id] = ks
		} else {
			// If not found by ID, look for any default key for this provider and index
			// (handles cases where the ID might have changed in code but we want to recover)
			found := false
			for existingID, ks := range rcm.config.APIKeys {
				if ks.Provider == provider && ks.IsDefault && strings.HasPrefix(existingID, fmt.Sprintf("%s-%d", provider, index)) {
					ks.Value = value
					rcm.config.APIKeys[existingID] = ks
					found = true
					break
				}
			}

			if !found {
				rcm.config.APIKeys[id] = APIKeyState{
					ID:        id,
					Provider:  provider,
					Value:     value,
					Enabled:   true,
					IsDefault: true,
				}
			}
		}
	}

	// Merge Gemini Keys
	geminiAPIKeys := viper.GetString("GEMINI_API_KEYS")
	if geminiAPIKeys != "" {
		for i, keyVal := range splitAPIKeys(geminiAPIKeys) {
			mergeKey("Gemini", i, keyVal)
		}
	}

	// Merge Groq Key
	groqAPIKey := viper.GetString("GROQ_API_KEY")
	if groqAPIKey != "" {
		mergeKey("Groq", 0, groqAPIKey)
	}

	// Merge Hugging Face Key
	huggingFaceAPIKey := viper.GetString("HUGGINGFACE_API_KEY")
	if huggingFaceAPIKey == "" {
		huggingFaceAPIKey = viper.GetString("HF_TOKEN")
	}
	if huggingFaceAPIKey != "" {
		mergeKey("Hugging Face", 0, huggingFaceAPIKey)
	}

	// Merge OpenRouter Key
	openRouterAPIKey := viper.GetString("OPENROUTER_API_KEY")
	if openRouterAPIKey != "" {
		mergeKey("OpenRouter", 0, openRouterAPIKey)
	}
}

// initializeProviderStates sets up the Providers map based on current environment/config.
func (rcm *RuntimeConfigManager) initializeProviderStates() {
	rcm.config.Providers["Gemini"] = ProviderState{
		Name:      "Gemini",
		Enabled:   viper.GetString("GEMINI_API_KEYS") != "",
		ModelName: viper.GetString("providers.gemini.model"),
	}
	rcm.config.Providers["Groq"] = ProviderState{
		Name:      "Groq",
		Enabled:   viper.GetString("GROQ_API_KEY") != "",
		ModelName: viper.GetString("providers.groq.model"),
	}
	rcm.config.Providers["Hugging Face"] = ProviderState{
		Name:      "Hugging Face",
		Enabled:   viper.GetString("HUGGINGFACE_API_KEY") != "" || viper.GetString("HF_TOKEN") != "",
		ModelName: viper.GetString("providers.huggingface.model"),
	}
	rcm.config.Providers["OpenRouter"] = ProviderState{
		Name:      "OpenRouter",
		Enabled:   viper.GetString("OPENROUTER_API_KEY") != "",
		ModelName: viper.GetString("providers.openrouter.model"),
	}
}

// initializeFromEnv populates the state from environment variables (bootstrap only).
func (rcm *RuntimeConfigManager) initializeFromEnv() {
	// Global AI enabled (default to true if not explicitly false)
	rcm.config.AIEnabled = true
	if viper.IsSet("AI_ENABLED") {
		rcm.config.AIEnabled = viper.GetBool("AI_ENABLED")
	}

	// Environment
	rcm.config.Environment.Mode = viper.GetString("ENVIRONMENT_MODE")
	rcm.config.Environment.BackendHost = viper.GetString("BACKEND_HOST")
	rcm.config.Environment.IsolationEnabled = viper.GetBool("ENVIRONMENT_ISOLATION_ENABLED")

	// Initialize all supported providers with their configured models
	rcm.initializeProviderStates()

	// Providers and API Keys
	// Gemini
	geminiAPIKeys := viper.GetString("GEMINI_API_KEYS")
	if geminiAPIKeys != "" {
		for i, keyVal := range splitAPIKeys(geminiAPIKeys) {
			id := generateKeyID("Gemini", i)
			rcm.config.APIKeys[id] = APIKeyState{
				ID:        id,
				Provider:  "Gemini",
				Value:     keyVal,
				Enabled:   true,
				IsDefault: true,
			}
		}
	}

	// Groq
	groqAPIKey := viper.GetString("GROQ_API_KEY")
	if groqAPIKey != "" {
		id := generateKeyID("Groq", 0)
		rcm.config.APIKeys[id] = APIKeyState{
			ID:        id,
			Provider:  "Groq",
			Value:     groqAPIKey,
			Enabled:   true,
			IsDefault: true,
		}
	}

	// Hugging Face
	huggingFaceAPIKey := viper.GetString("HUGGINGFACE_API_KEY")
	if huggingFaceAPIKey == "" {
		huggingFaceAPIKey = viper.GetString("HF_TOKEN")
	}
	if huggingFaceAPIKey != "" {
		id := generateKeyID("Hugging Face", 0)
		rcm.config.APIKeys[id] = APIKeyState{
			ID:        id,
			Provider:  "Hugging Face",
			Value:     huggingFaceAPIKey,
			Enabled:   true,
			IsDefault: true,
		}
	}

	// OpenRouter
	openRouterAPIKey := viper.GetString("OPENROUTER_API_KEY")
	if openRouterAPIKey != "" {
		id := generateKeyID("OpenRouter", 0)
		rcm.config.APIKeys[id] = APIKeyState{
			ID:        id,
			Provider:  "OpenRouter",
			Value:     openRouterAPIKey,
			Enabled:   true,
			IsDefault: true,
		}
	}

	// TODO: Add other providers (ONNX if re-enabled)
}

// GetConfig provides a read-only copy of the current RuntimeConfig.
func (rcm *RuntimeConfigManager) GetConfig() RuntimeConfig {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()
	// Return a deep copy to prevent external modification
	return rcm.config // Shallow copy is okay for structs, but maps need deep copy
}

// SetAIEnabled updates the global AI enabled status.
func (rcm *RuntimeConfigManager) SetAIEnabled(enabled bool) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	rcm.config.AIEnabled = enabled
	return rcm.PersistStateToDB()
}

// SetActiveProvider updates the preferred active AI provider.
func (rcm *RuntimeConfigManager) SetActiveProvider(providerName string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if _, ok := rcm.config.Providers[providerName]; ok || providerName == "None" {
		rcm.config.ActiveProvider = providerName
		return rcm.PersistStateToDB()
	}
	return fmt.Errorf("provider %s not found", providerName)
}

// SetProviderState updates the state of a specific provider.
func (rcm *RuntimeConfigManager) SetProviderState(providerName string, enabled, paused, blocked bool, blockedReason string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if ps, ok := rcm.config.Providers[providerName]; ok {
		ps.Enabled = enabled
		ps.Paused = paused
		ps.Blocked = blocked
		ps.BlockedReason = blockedReason
		rcm.config.Providers[providerName] = ps
		return rcm.PersistStateToDB()
	}
	return fmt.Errorf("provider %s not found", providerName)
}

// AddAPIKey adds a new API key for a provider.
func (rcm *RuntimeConfigManager) AddAPIKey(providerName, keyValue string, enabled bool) (string, error) {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	if _, ok := rcm.config.Providers[providerName]; !ok {
		return "", fmt.Errorf("provider %s not found", providerName)
	}

	keyID := uuid.New().String()
	rcm.config.APIKeys[keyID] = APIKeyState{
		ID:        keyID,
		Provider:  providerName,
		Value:     keyValue,
		Enabled:   enabled,
		IsDefault: false, // Newly added keys are not from .env
	}
	if err := rcm.PersistStateToDB(); err != nil {
		return "", err
	}
	return keyID, nil
}

// RemoveAPIKey removes an API key.
func (rcm *RuntimeConfigManager) RemoveAPIKey(keyID string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if _, ok := rcm.config.APIKeys[keyID]; ok {
		delete(rcm.config.APIKeys, keyID)
		return rcm.PersistStateToDB()
	}
	return fmt.Errorf("API key %s not found", keyID)
}

// SetAPIKeyStatus updates the enabled/blocked status of an API key.
func (rcm *RuntimeConfigManager) SetAPIKeyStatus(keyID string, enabled, blocked bool, blockedReason string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if ks, ok := rcm.config.APIKeys[keyID]; ok {
		ks.Enabled = enabled
		ks.Blocked = blocked
		ks.BlockedReason = blockedReason
		rcm.config.APIKeys[keyID] = ks
		return rcm.PersistStateToDB()
	}
	return fmt.Errorf("API key %s not found", keyID)
}

// RotateAPIKey marks the current key as blocked and selects another enabled key for the provider.
func (rcm *RuntimeConfigManager) RotateAPIKey(providerName string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	// Find active key for provider
	var currentKeyID string
	for _, ks := range rcm.config.APIKeys {
		if ks.Provider == providerName && ks.Enabled && !ks.Blocked {
			currentKeyID = ks.ID
			break
		}
	}

	if currentKeyID == "" {
		return fmt.Errorf("no active API key found for provider %s to rotate", providerName)
	}

	// Block current key
	ks := rcm.config.APIKeys[currentKeyID]
	ks.Enabled = false
	ks.Blocked = true
	ks.BlockedReason = "rotated_out_manually" // Or "rotated_out_automatically"
	ks.RotatedCount++
	rcm.config.APIKeys[currentKeyID] = ks

	// Find next available key
	for _, nextKs := range rcm.config.APIKeys {
		if nextKs.Provider == providerName && nextKs.ID != currentKeyID && nextKs.Enabled && !nextKs.Blocked {
			// Found a new key, mark it active (no need to change anything if it's already enabled and not blocked)
			// For now, simply finding the next available is enough. The AI service will pick it up.
			slog.Info("Rotated API key for provider.", "provider", providerName, "new_key_id", nextKs.ID)
			return rcm.PersistStateToDB()
		}
	}

	return fmt.Errorf("no suitable API key found for provider %s after rotating current key %s", providerName, currentKeyID)
}

// UpdateKeyUsage updates key's last used time and optionally quota.
func (rcm *RuntimeConfigManager) UpdateKeyUsage(keyID string, lastError string, quotaRemaining int) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if ks, ok := rcm.config.APIKeys[keyID]; ok {
		ks.LastUsedAt = time.Now()
		ks.LastError = lastError
		if quotaRemaining >= 0 { // Only update if a valid quota is provided
			ks.QuotaRemaining = quotaRemaining
		}
		if lastError != "" {
			ks.Blocked = true
			ks.BlockedReason = lastError // Use error as reason
		} else {
			ks.Blocked = false
			ks.BlockedReason = ""
		}
		rcm.config.APIKeys[keyID] = ks
		return rcm.PersistStateToDB()
	}
	return fmt.Errorf("API key %s not found", keyID)
}

// SetEnvironmentState updates the bot's operational environment.
func (rcm *RuntimeConfigManager) SetEnvironmentState(mode, backendHost string, isolationEnabled bool) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	rcm.config.Environment.Mode = mode
	rcm.config.Environment.BackendHost = backendHost
	rcm.config.Environment.IsolationEnabled = isolationEnabled
	return rcm.PersistStateToDB()
}

// PersistStateToDB saves the current RuntimeConfig to the database.
func (rcm *RuntimeConfigManager) PersistStateToDB() error {
	rcm.mu.RLock() // Use RLock here to avoid modifying config while marshalling
	data, err := json.Marshal(rcm.config)
	rcm.mu.RUnlock() // Release RLock before potential heavy DB operation

	if err != nil {
		return fmt.Errorf("failed to marshal runtime config: %w", err)
	}

	// Use a dedicated table for runtime config, or a key-value store in a table
	// For simplicity, let's assume a single row in a 'runtime_config' table
	_, err = rcm.db.Exec(`
		INSERT OR REPLACE INTO runtime_config (id, config_data, updated_at)
		VALUES (1, ?, ?)
	`, data, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save runtime config to DB: %w", err)
	}
	return nil
}

// LoadStateFromDB loads the RuntimeConfig from the database.
func (rcm *RuntimeConfigManager) LoadStateFromDB() error {
	var configData []byte
	var updatedAt time.Time

	row := rcm.db.QueryRow("SELECT config_data, updated_at FROM runtime_config WHERE id = 1")
	err := row.Scan(&configData, &updatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("runtime config not found in DB, starting with defaults")
	}
	if err != nil {
		return fmt.Errorf("failed to query runtime config from DB: %w", err)
	}

	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if err := json.Unmarshal(configData, &rcm.config); err != nil {
		return fmt.Errorf("failed to unmarshal runtime config from DB: %w", err)
	}
	return nil
}

// generateKeyID creates a unique ID for an API key.
// For default keys (from env), it produces a deterministic ID.
func generateKeyID(providerName string, index int) string {
	return fmt.Sprintf("%s-%d-default", providerName, index)
}

// splitAPIKeys splits a comma-separated string of API keys.
func splitAPIKeys(keys string) []string {
	var result []string
	if keys == "" {
		return result
	}
	for _, key := range strings.Split(keys, ",") {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey != "" {
			result = append(result, trimmedKey)
		}
	}
	return result
}

// GetDB returns the underlying database connection.
func (rcm *RuntimeConfigManager) GetDB() *sql.DB {
	return rcm.db
}
