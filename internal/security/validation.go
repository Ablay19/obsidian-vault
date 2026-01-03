package security

import (
	"regexp"
)

var (
	// Basic regex for common API key formats (can be refined per provider)
	// Gemini: usually starts with AIza
	// Groq: starts with gsk_
	// OpenRouter: sk-or-v1-
	apiKeyRegex = regexp.MustCompile(`^[A-Za-z0-9_\-\.]{8,256}$`)
)

// ValidateAPIKeyFormat checks if the key matches basic structural rules.
func ValidateAPIKeyFormat(key string) bool {
	return apiKeyRegex.MatchString(key)
}
