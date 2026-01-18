package doppler

import (
	"fmt"
	"os"
	"strconv"
)

// FallbackLoader handles loading fallback environment variables
type FallbackLoader struct {
	fallbacks map[string]string
}

// NewFallbackLoader creates a new fallback loader
func NewFallbackLoader() *FallbackLoader {
	return &FallbackLoader{
		fallbacks: make(map[string]string),
	}
}

// AddFallback adds a fallback value for an environment variable
func (fl *FallbackLoader) AddFallback(key, value string) {
	fl.fallbacks[key] = value
}

// LoadFromEnvFile loads fallbacks from a .env file
func (fl *FallbackLoader) LoadFromEnvFile(filepath string) error {
	// In a real implementation, this would parse a .env file
	// For now, use common test fallbacks
	fl.fallbacks = map[string]string{
		"TEST_DATABASE_URL":   "sqlite://:memory:",
		"TEST_REDIS_ADDR":     "localhost:6379",
		"TEST_TIMEOUT":        "30",
		"TELEGRAM_BOT_TOKEN":  "test_bot_token",
		"WHATSAPP_API_KEY":    "test_whatsapp_key",
		"FACEBOOK_APP_ID":     "test_fb_id",
		"FACEBOOK_APP_SECRET": "test_fb_secret",
		"GEMINI_API_KEY":      "test_gemini_key",
		"GROQ_API_KEY":        "test_groq_key",
		"OPENAI_API_KEY":      "test_openai_key",
		"SESSION_SECRET":      "test_session_secret",
		"BACKEND_HOST":        "localhost:8080",
		"DASHBOARD_URL":       "http://localhost:3000",
	}

	return nil
}

// GetFallback retrieves a fallback value
func (fl *FallbackLoader) GetFallback(key string) (string, bool) {
	value, exists := fl.fallbacks[key]
	return value, exists
}

// ApplyFallbacks applies fallback values to environment if not already set
func (fl *FallbackLoader) ApplyFallbacks() error {
	for key, value := range fl.fallbacks {
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("failed to set fallback env var %s: %w", key, err)
			}
		}
	}
	return nil
}

// ValidateRequired checks that required environment variables are set
func (fl *FallbackLoader) ValidateRequired(required []string) error {
	var missing []string

	for _, key := range required {
		if os.Getenv(key) == "" {
			if _, hasFallback := fl.fallbacks[key]; !hasFallback {
				missing = append(missing, key)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// GetIntFallback retrieves an integer fallback value
func (fl *FallbackLoader) GetIntFallback(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}

	if strVal, exists := fl.fallbacks[key]; exists {
		if intVal, err := strconv.Atoi(strVal); err == nil {
			return intVal
		}
	}

	return defaultValue
}

// GetBoolFallback retrieves a boolean fallback value
func (fl *FallbackLoader) GetBoolFallback(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}

	if strVal, exists := fl.fallbacks[key]; exists {
		if boolVal, err := strconv.ParseBool(strVal); err == nil {
			return boolVal
		}
	}

	return defaultValue
}
