package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"obsidian-automation/internal/security"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
		zap.S().Warn("Could not load state from DB, initializing with defaults", "error", err)
		// If DB load fails, initialize from .env (bootstrap)
		rcm.initializeFromEnv()
	} else {
		zap.S().Info("State loaded from DB successfully. Merging keys from environment.")
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
		zap.S().Info("AI_ENABLED status overridden by environment variable.", "enabled", rcm.config.AIEnabled)
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

	// Merge Gemini Keys (Handle both plural and singular env vars)
	geminiAPIKeys := os.Getenv("GEMINI_API_KEYS")
	if geminiAPIKeys == "" {
		geminiAPIKeys = os.Getenv("GEMINI_API_KEY")
	}

	if geminiAPIKeys != "" {
		for i, keyVal := range splitAPIKeys(geminiAPIKeys) {
			if validateAPIKey(keyVal) {
				mergeKey("Gemini", i, keyVal)
			}
		}
	}

	// Merge Groq Key
	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if validateAPIKey(groqAPIKey) {
		mergeKey("Groq", 0, groqAPIKey)
	}

	// Merge Hugging Face Key
	huggingFaceAPIKey := os.Getenv("HUGGINGFACE_API_KEY")
	if huggingFaceAPIKey == "" {
		huggingFaceAPIKey = os.Getenv("HF_TOKEN")
	}
	if validateAPIKey(huggingFaceAPIKey) {
		mergeKey("Hugging Face", 0, huggingFaceAPIKey)
	}

	// Merge OpenRouter Key
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	if validateAPIKey(openRouterAPIKey) {
		mergeKey("OpenRouter", 0, openRouterAPIKey)
	}

	// Merge Cloudflare Worker URL
	cloudflareWorkerURL := os.Getenv("CLOUDFLARE_WORKER_URL")
	if cloudflareWorkerURL != "" {
		// For Cloudflare, we validate URL format instead of API key format
		if len(cloudflareWorkerURL) > 10 && (strings.Contains(cloudflareWorkerURL, "workers.dev") || strings.Contains(cloudflareWorkerURL, "workers.workers.dev")) {
			mergeKey("Cloudflare", 0, cloudflareWorkerURL)
		}
	}
}

// initializeProviderStates sets up the Providers map based on current environment/config.
func (rcm *RuntimeConfigManager) initializeProviderStates() {
	// Helper to set and normalize provider
	setProvider := func(name, model string, enabled bool) {
		// Crucial: remove any legacy lowercase keys that might be in the map from previous versions
		delete(rcm.config.Providers, strings.ToLower(name))
		delete(rcm.config.Providers, name) // reset

		rcm.config.Providers[name] = ProviderState{
			Name:      name,
			Enabled:   enabled,
			ModelName: model,
		}
	}

	setProvider("Gemini", viper.GetString("providers.gemini.model"), os.Getenv("GEMINI_API_KEYS") != "" || os.Getenv("GEMINI_API_KEY") != "")
	setProvider("Groq", viper.GetString("providers.groq.model"), os.Getenv("GROQ_API_KEY") != "")
	setProvider("Hugging Face", viper.GetString("providers.huggingface.model"), os.Getenv("HUGGINGFACE_API_KEY") != "" || os.Getenv("HF_TOKEN") != "")
	setProvider("OpenRouter", viper.GetString("providers.openrouter.model"), os.Getenv("OPENROUTER_API_KEY") != "")
	setProvider("Cloudflare", "@cf/meta/llama-3-8b-instruct", os.Getenv("CLOUDFLARE_WORKER_URL") != "")

	// Normalize ActiveProvider if it was lowercase
	if rcm.config.ActiveProvider != "" {
		for _, name := range []string{"Gemini", "Groq", "Hugging Face", "OpenRouter", "Cloudflare"} {
			if strings.EqualFold(rcm.config.ActiveProvider, name) {
				rcm.config.ActiveProvider = name
				break
			}
		}
	}
}

// initializeFromEnv populates the state from environment variables (bootstrap only).
func (rcm *RuntimeConfigManager) initializeFromEnv() {
	// Global AI enabled (default to true if not explicitly false)
	rcm.config.AIEnabled = true
	if os.Getenv("AI_ENABLED") != "" {
		rcm.config.AIEnabled = os.Getenv("AI_ENABLED") == "true"
	}

	// Set default provider to Cloudflare if available, otherwise use environment preference
	if os.Getenv("CLOUDFLARE_WORKER_URL") != "" {
		rcm.config.ActiveProvider = "Cloudflare"
	} else if os.Getenv("ACTIVE_PROVIDER") != "" {
		rcm.config.ActiveProvider = os.Getenv("ACTIVE_PROVIDER")
	} else if os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GEMINI_API_KEYS") != "" {
		rcm.config.ActiveProvider = "Gemini"
	}

	// Environment
	rcm.config.Environment.Mode = os.Getenv("ENVIRONMENT_MODE")
	rcm.config.Environment.BackendHost = os.Getenv("BACKEND_HOST")
	rcm.config.Environment.IsolationEnabled = os.Getenv("ENVIRONMENT_ISOLATION_ENABLED") == "true"

	// Initialize all supported providers with their configured models
	rcm.initializeProviderStates()

	// Providers and API Keys
	// Gemini
	geminiAPIKeys := os.Getenv("GEMINI_API_KEYS")
	if geminiAPIKeys != "" {
		for i, keyVal := range splitAPIKeys(geminiAPIKeys) {
			if validateAPIKey(keyVal) {
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
	}

	// Groq
	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if validateAPIKey(groqAPIKey) {
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
	huggingFaceAPIKey := os.Getenv("HUGGINGFACE_API_KEY")
	if huggingFaceAPIKey == "" {
		huggingFaceAPIKey = os.Getenv("HF_TOKEN")
	}
	if validateAPIKey(huggingFaceAPIKey) {
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
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	if validateAPIKey(openRouterAPIKey) {
		id := generateKeyID("OpenRouter", 0)
		rcm.config.APIKeys[id] = APIKeyState{
			ID:        id,
			Provider:  "OpenRouter",
			Value:     openRouterAPIKey,
			Enabled:   true,
			IsDefault: true,
		}
	}

	// Cloudflare
	cloudflareWorkerURL := os.Getenv("CLOUDFLARE_WORKER_URL")
	if cloudflareWorkerURL != "" {
		// Validate URL format for Cloudflare
		if len(cloudflareWorkerURL) > 10 && (strings.Contains(cloudflareWorkerURL, "workers.dev") || strings.Contains(cloudflareWorkerURL, "workers.workers.dev")) {
			id := generateKeyID("Cloudflare", 0)
			rcm.config.APIKeys[id] = APIKeyState{
				ID:        id,
				Provider:  "Cloudflare",
				Value:     cloudflareWorkerURL,
				Enabled:   true,
				IsDefault: true,
			}
		}
	}

	// TODO: Add other providers (ONNX if re-enabled)
}

// GetConfig provides a read-only copy of the current RuntimeConfig.
func (rcm *RuntimeConfigManager) GetConfig(redact bool) RuntimeConfig {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	// Perform a deep copy to ensure isolation
	copy := RuntimeConfig{
		AIEnabled:      rcm.config.AIEnabled,
		ActiveProvider: rcm.config.ActiveProvider,
		Providers:      make(map[string]ProviderState),
		APIKeys:        make(map[string]APIKeyState),
		Environment:    rcm.config.Environment,
	}

	for k, v := range rcm.config.Providers {
		copy.Providers[k] = v
	}

	for k, v := range rcm.config.APIKeys {
		if redact {
			masked := v
			if masked.Value != "" {
				masked.Value = "********" // Redact sensitive value
			}
			masked.EncryptedValue = "" // Don't expose encrypted value either
			copy.APIKeys[k] = masked
		} else {
			copy.APIKeys[k] = v // Return original unmasked value
		}
	}

	return copy
}

// SetAIEnabled updates the global AI enabled status.
func (rcm *RuntimeConfigManager) SetAIEnabled(enabled bool) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	rcm.config.AIEnabled = enabled
	return rcm.persistStateToDBUnprotected()
}

// SetActiveProvider updates the preferred active AI provider.
func (rcm *RuntimeConfigManager) SetActiveProvider(providerName string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	if _, ok := rcm.config.Providers[providerName]; ok || providerName == "None" {
		rcm.config.ActiveProvider = providerName
		return rcm.persistStateToDBUnprotected()
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
		return rcm.persistStateToDBUnprotected()
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

	if !validateAPIKey(keyValue) {
		return "", fmt.Errorf("invalid API key format")
	}

	keyID := uuid.New().String()
	rcm.config.APIKeys[keyID] = APIKeyState{
		ID:        keyID,
		Provider:  providerName,
		Value:     strings.TrimSpace(keyValue),
		Enabled:   enabled,
		IsDefault: false, // Newly added keys are not from .env
	}

	zap.S().Info("New API key added", "provider", providerName, "key_id", keyID)

	if err := rcm.persistStateToDBUnprotected(); err != nil {
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
		zap.S().Info("API key removed", "key_id", keyID)
		return rcm.persistStateToDBUnprotected()
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
		return rcm.persistStateToDBUnprotected()
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
			zap.S().Info("Rotated API key for provider.", "provider", providerName, "new_key_id", nextKs.ID)
			return rcm.persistStateToDBUnprotected()
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
		return rcm.persistStateToDBUnprotected()
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
	return rcm.persistStateToDBUnprotected()
}

// PersistStateToDB saves the current RuntimeConfig to the database.
func (rcm *RuntimeConfigManager) PersistStateToDB() error {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()
	return rcm.persistStateToDBUnprotected()
}

// persistStateToDBUnprotected performs the actual database save without acquiring locks.
// It assumes the caller is already holding a lock (Lock or RLock).
func (rcm *RuntimeConfigManager) persistStateToDBUnprotected() error {
	// Prepare for persistence: encrypt all keys
	for id, ks := range rcm.config.APIKeys {
		if ks.Value != "" {
			encrypted, err := security.Encrypt(ks.Value)
			if err != nil {
				zap.S().Error("Failed to encrypt API key", "key_id", id, "error", err)
				continue
			}
			ks.EncryptedValue = encrypted
			rcm.config.APIKeys[id] = ks
		}
	}

	data, err := json.Marshal(rcm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal runtime config: %w", err)
	}

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

	// Decrypt loaded keys
	for id, ks := range rcm.config.APIKeys {
		if ks.EncryptedValue != "" {
			decrypted, err := security.Decrypt(ks.EncryptedValue)
			if err != nil {
				zap.S().Error("Failed to decrypt API key from storage", "key_id", id, "error", err)
				continue
			}
			ks.Value = decrypted
			rcm.config.APIKeys[id] = ks
		}
	}

	// Sanitize loaded keys: remove any that are invalid and normalize provider names
	dirty := false
	for id, ks := range rcm.config.APIKeys {
		if !validateAPIKey(ks.Value) {
			zap.S().Warn("Removing invalid API key found in DB", "key_id", id, "provider", ks.Provider)
			delete(rcm.config.APIKeys, id)
			dirty = true
			continue
		}

		// Normalize provider name
		for _, name := range []string{"Gemini", "Groq", "Hugging Face", "OpenRouter"} {
			if strings.EqualFold(ks.Provider, name) && ks.Provider != name {
				zap.S().Info("Normalizing provider name for API key", "old", ks.Provider, "new", name)
				ks.Provider = name
				rcm.config.APIKeys[id] = ks
				dirty = true
				break
			}
		}
	}
	if dirty {
		// Persist the clean state back to DB immediately (in background to avoid lock issues)
		go func() {
			time.Sleep(100 * time.Millisecond) // yield
			rcm.mu.Lock()
			defer rcm.mu.Unlock()
			if err := rcm.persistStateToDBUnprotected(); err != nil {
				zap.S().Error("Failed to persist cleaned state", "error", err)
			}
		}()
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

// validateAPIKey checks if a key is valid (non-empty, minimum length).
func validateAPIKey(key string) bool {
	return security.ValidateAPIKeyFormat(key)
}

// GetDB returns the underlying database connection.
func (rcm *RuntimeConfigManager) GetDB() *sql.DB {
	return rcm.db
}

// ResetState clears all API keys and resets providers to disabled. Used primarily for testing.
func (rcm *RuntimeConfigManager) ResetState() {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()
	rcm.config.APIKeys = make(map[string]APIKeyState)

	// Reset providers to known defaults but disabled
	rcm.config.Providers = map[string]ProviderState{
		"Gemini":       {Name: "Gemini", Enabled: false, ModelName: "gemini-pro"},
		"Groq":         {Name: "Groq", Enabled: false, ModelName: "llama3-70b"},
		"Hugging Face": {Name: "Hugging Face", Enabled: false, ModelName: "gpt2"},
		"OpenRouter":   {Name: "OpenRouter", Enabled: false, ModelName: "gpt-3.5-turbo"},
	}
	rcm.config.ActiveProvider = ""
	rcm.persistStateToDBUnprotected()
}
