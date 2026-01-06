package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EnvValidator validates environment variables
type EnvValidator struct {
	requiredVars []string
	optionalVars map[string]string
	validators   map[string]func(string) error
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Missing  []string `json:"missing"`
	Invalid  []string `json:"invalid"`
}

// NewEnvValidator creates a new environment validator
func NewEnvValidator() *EnvValidator {
	return &EnvValidator{
		requiredVars: []string{
			"TURSO_DATABASE_URL",
			"TURSO_AUTH_TOKEN",
			"TELEGRAM_BOT_TOKEN",
			"SESSION_SECRET",
		},
		optionalVars: map[string]string{
			"GEMINI_API_KEYS":       "Gemini API keys (comma-separated)",
			"GROQ_API_KEY":          "Groq API key",
			"HUGGINGFACE_API_KEY":   "HuggingFace API key",
			"OPENROUTER_API_KEY":    "OpenRouter API key",
			"WHATSAPP_ACCESS_TOKEN": "WhatsApp access token",
			"WHATSAPP_VERIFY_TOKEN": "WhatsApp verify token",
			"WHATSAPP_APP_SECRET":   "WhatsApp app secret",
			"GOOGLE_CLIENT_ID":      "Google OAuth client ID",
			"GOOGLE_CLIENT_SECRET":  "Google OAuth client secret",
			"GOOGLE_REDIRECT_URL":   "Google OAuth redirect URL",
			"VAULT_ADDR":            "Vault server address",
			"VAULT_TOKEN":           "Vault authentication token",
		},
		validators: map[string]func(string) error{
			"TURSO_DATABASE_URL":    validateTursoURL,
			"TELEGRAM_BOT_TOKEN":    validateTelegramToken,
			"SESSION_SECRET":        validateSessionSecret,
			"GOOGLE_CLIENT_ID":      validateGoogleClientID,
			"GOOGLE_CLIENT_SECRET":  validateGoogleClientSecret,
			"VAULT_ADDR":            validateVaultAddr,
			"GEMINI_API_KEYS":       validateGeminiKeys,
			"GROQ_API_KEY":          validateAPIKey,
			"HUGGINGFACE_API_KEY":   validateAPIKey,
			"OPENROUTER_API_KEY":    validateAPIKey,
			"WHATSAPP_ACCESS_TOKEN": validateAPIToken,
			"WHATSAPP_VERIFY_TOKEN": validateAPIToken,
			"WHATSAPP_APP_SECRET":   validateAPISecret,
			"VAULT_TOKEN":           validateVaultToken,
		},
	}
}

// ValidateEnvironment validates the current environment
func (ev *EnvValidator) ValidateEnvironment() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Missing:  []string{},
		Invalid:  []string{},
	}

	// Check required variables
	for _, varName := range ev.requiredVars {
		value := os.Getenv(varName)
		if value == "" {
			result.Missing = append(result.Missing, varName)
			result.Valid = false
		}
	}

	// Validate optional variables that are set
	for varName := range ev.optionalVars {
		value := os.Getenv(varName)
		if value != "" {
			if validator, exists := ev.validators[varName]; exists {
				if err := validator(value); err != nil {
					result.Invalid = append(result.Invalid, fmt.Sprintf("%s: %v", varName, err))
					result.Valid = false
				}
			}
		}
	}

	// Add warnings for recommended optional variables
	ev.addWarnings(result)

	return result
}

// ValidateEnvironmentForTesting validates environment for testing
func (ev *EnvValidator) ValidateEnvironmentForTesting() *ValidationResult {
	result := ev.ValidateEnvironment()

	// For testing, some required vars can be mocked
	testRequiredVars := []string{
		"TELEGRAM_BOT_TOKEN",
		"SESSION_SECRET",
	}

	// Remove test-required vars from missing list
	var filteredMissing []string
	for _, missing := range result.Missing {
		isTestVar := false
		for _, testVar := range testRequiredVars {
			if missing == testVar {
				isTestVar = true
				break
			}
		}
		if !isTestVar {
			filteredMissing = append(filteredMissing, missing)
		}
	}
	result.Missing = filteredMissing

	return result
}

// addWarnings adds contextual warnings
func (ev *EnvValidator) addWarnings(result *ValidationResult) {
	// Check for development vs production warnings
	envMode := os.Getenv("ENVIRONMENT_MODE")
	if envMode == "prod" {
		if os.Getenv("VAULT_TOKEN") == "root" {
			result.Warnings = append(result.Warnings, "Using 'root' Vault token in production is insecure")
		}

		if os.Getenv("SESSION_SECRET") == "change-me-to-something-very-secure" {
			result.Warnings = append(result.Warnings, "Using default session secret in production is insecure")
		}
	}

	// Check for missing AI providers
	hasAnyAI := os.Getenv("GEMINI_API_KEYS") != "" ||
		os.Getenv("GROQ_API_KEY") != "" ||
		os.Getenv("OPENROUTER_API_KEY") != "" ||
		os.Getenv("HUGGINGFACE_API_KEY") != ""

	if !hasAnyAI {
		result.Warnings = append(result.Warnings, "No AI provider keys configured - AI features will be disabled")
	}
}

// Validation functions
func validateTursoURL(url string) error {
	if url == "" {
		return nil // Empty is handled separately
	}

	if !strings.HasPrefix(url, "libsql://") {
		return fmt.Errorf("Turso URL must start with 'libsql://'")
	}

	return nil
}

func validateTelegramToken(token string) error {
	if token == "" {
		return nil // Empty is handled separately
	}

	// Telegram tokens are typically long alphanumeric strings
	if len(token) < 10 {
		return fmt.Errorf("Telegram token too short")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(token) {
		return fmt.Errorf("Telegram token contains invalid characters")
	}

	return nil
}

func validateSessionSecret(secret string) error {
	if secret == "" {
		return nil // Empty is handled separately
	}

	if len(secret) < 32 {
		return fmt.Errorf("Session secret must be at least 32 characters long")
	}

	if secret == "change-me-to-something-very-secure" {
		return fmt.Errorf("Session secret must be changed from default")
	}

	return nil
}

func validateGoogleClientID(clientID string) error {
	if clientID == "" {
		return nil // Empty is handled separately
	}

	// Google Client IDs are typically long alphanumeric strings
	if len(clientID) < 10 {
		return fmt.Errorf("Google Client ID too short")
	}

	if !regexp.MustCompile(`^[0-9a-zA-Z.-]+$`).MatchString(clientID) {
		return fmt.Errorf("Google Client ID format appears invalid")
	}

	return nil
}

func validateGoogleClientSecret(secret string) error {
	if secret == "" {
		return nil // Empty is handled separately
	}

	if len(secret) < 10 {
		return fmt.Errorf("Google Client Secret too short")
	}

	return nil
}

func validateVaultAddr(addr string) error {
	if addr == "" {
		return nil // Empty is handled separately
	}

	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return fmt.Errorf("Vault address must start with http:// or https://")
	}

	return nil
}

func validateGeminiKeys(keys string) error {
	if keys == "" {
		return nil // Empty is handled separately
	}

	keyList := strings.Split(keys, ",")
	for _, key := range keyList {
		key = strings.TrimSpace(key)
		if err := validateAPIKey(key); err != nil {
			return fmt.Errorf("Invalid Gemini API key: %v", err)
		}
	}

	return nil
}

func validateAPIKey(key string) error {
	if key == "" {
		return nil // Empty is handled separately
	}

	if len(key) < 10 {
		return fmt.Errorf("API key too short")
	}

	return nil
}

func validateAPIToken(token string) error {
	return validateAPIKey(token)
}

func validateAPISecret(secret string) error {
	if secret == "" {
		return nil // Empty is handled separately
	}

	if len(secret) < 8 {
		return fmt.Errorf("API secret too short")
	}

	return nil
}

func validateVaultToken(token string) error {
	if token == "" {
		return nil // Empty is handled separately
	}

	if len(token) < 8 {
		return fmt.Errorf("Vault token too short")
	}

	return nil
}

// Test functions
func TestEnvValidator_ValidateEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		validator *EnvValidator
		wantValid bool
	}{
		{
			name: "valid environment",
			setup: func() {
				os.Setenv("TURSO_DATABASE_URL", "libsql://test.turso.io")
				os.Setenv("TELEGRAM_BOT_TOKEN", "123456789:ABCDEF123456")
				os.Setenv("SESSION_SECRET", "this-is-a-very-secure-session-secret-32chars")
			},
			validator: NewEnvValidator(),
			wantValid: true,
		},
		{
			name: "missing required variables",
			setup: func() {
				os.Unsetenv("TURSO_DATABASE_URL")
				os.Unsetenv("SESSION_SECRET")
			},
			validator: NewEnvValidator(),
			wantValid: false,
		},
		{
			name: "invalid Turso URL",
			setup: func() {
				os.Setenv("TURSO_DATABASE_URL", "invalid-url")
			},
			validator: NewEnvValidator(),
			wantValid: false,
		},
		{
			name: "short session secret",
			setup: func() {
				os.Setenv("SESSION_SECRET", "short")
			},
			validator: NewEnvValidator(),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tt.setup()
			defer func() {
				// Cleanup test environment
				os.Unsetenv("TURSO_DATABASE_URL")
				os.Unsetenv("TELEGRAM_BOT_TOKEN")
				os.Unsetenv("SESSION_SECRET")
			}()

			result := tt.validator.ValidateEnvironment()
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestEnvValidator_ValidateEnvironmentForTesting(t *testing.T) {
	validator := NewEnvValidator()

	// Set minimal test environment
	os.Setenv("SESSION_SECRET", "test-secret-that-is-long-enough-32")
	defer os.Unsetenv("SESSION_SECRET")

	result := validator.ValidateEnvironmentForTesting()

	// Should be valid for testing with minimal config
	require.True(t, result.Valid)

	// Should not include missing vars that aren't required for testing
	for _, missing := range result.Missing {
		if missing == "TELEGRAM_BOT_TOKEN" {
			t.Errorf("TELEGRAM_BOT_TOKEN should not be required for testing")
		}
	}
}

// TestProductionWarnings tests production environment warnings
func TestProductionWarnings(t *testing.T) {
	validator := NewEnvValidator()

	// Set production environment with insecure defaults
	os.Setenv("ENVIRONMENT_MODE", "prod")
	os.Setenv("VAULT_TOKEN", "root")
	os.Setenv("SESSION_SECRET", "change-me-to-something-very-secure")
	os.Setenv("TURSO_DATABASE_URL", "libsql://test.turso.io")

	defer func() {
		os.Unsetenv("ENVIRONMENT_MODE")
		os.Unsetenv("VAULT_TOKEN")
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("TURSO_DATABASE_URL")
	}()

	result := validator.ValidateEnvironment()
	require.False(t, result.Valid)
	require.NotEmpty(t, result.Warnings)

	// Should have security warnings
	hasVaultWarning := false
	hasSessionWarning := false

	for _, warning := range result.Warnings {
		if strings.Contains(warning, "root Vault token") {
			hasVaultWarning = true
		}
		if strings.Contains(warning, "default session secret") {
			hasSessionWarning = true
		}
	}

	assert.True(t, hasVaultWarning, "Should warn about root Vault token")
	assert.True(t, hasSessionWarning, "Should warn about default session secret")
}

// Benchmark validation performance
func BenchmarkEnvValidation(b *testing.B) {
	validator := NewEnvValidator()

	// Setup environment once
	os.Setenv("TURSO_DATABASE_URL", "libsql://test.turso.io")
	os.Setenv("TELEGRAM_BOT_TOKEN", "123456789:ABCDEF123456")
	os.Setenv("SESSION_SECRET", "this-is-a-very-secure-session-secret-32chars")
	defer func() {
		os.Unsetenv("TURSO_DATABASE_URL")
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("SESSION_SECRET")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateEnvironment()
	}
}
