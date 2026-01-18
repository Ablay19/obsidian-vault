package doppler

import (
	"fmt"
	"regexp"
	"strings"
)

// SecurityManager handles credential sanitization and security
type SecurityManager struct {
	sensitivePatterns []*regexp.Regexp
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		sensitivePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(password|passwd|pwd|secret|key|token|auth)\s*[:=]\s*["']?[^"'\s]+["']?`),
			regexp.MustCompile(`(?i)(api_key|apikey|access_token)\s*[:=]\s*["']?[^"'\s]+["']?`),
			regexp.MustCompile(`(?i)bearer\s+[^"'\s]+`),
			regexp.MustCompile(`(?i)authorization\s*[:=]\s*["']?[^"'\s]+["']?`),
		},
	}
}

// SanitizeString removes sensitive information from a string
func (sm *SecurityManager) SanitizeString(input string) string {
	result := input
	for _, pattern := range sm.sensitivePatterns {
		result = pattern.ReplaceAllString(result, "$1: [REDACTED]")
	}
	return result
}

// SanitizeMap removes sensitive information from a map
func (sm *SecurityManager) SanitizeMap(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	for k, v := range data {
		if sm.isSensitiveKey(k) {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = sm.sanitizeValue(v)
		}
	}
	return sanitized
}

// IsSecureValue checks if a value appears to be properly secured
func (sm *SecurityManager) IsSecureValue(key, value string) bool {
	if sm.isSensitiveKey(key) {
		// Sensitive values should not be empty or obvious placeholders
		if value == "" || strings.Contains(strings.ToLower(value), "placeholder") ||
			strings.Contains(strings.ToLower(value), "example") || strings.Contains(value, "test_") {
			return false
		}
		// Should not contain common insecure patterns
		if strings.Contains(value, "123456") || strings.Contains(value, "password") {
			return false
		}
	}
	return true
}

// ValidateCredentials performs basic validation on credentials
func (sm *SecurityManager) ValidateCredentials(creds map[string]string) []string {
	var issues []string

	for key, value := range creds {
		if !sm.IsSecureValue(key, value) {
			issues = append(issues, fmt.Sprintf("insecure value for %s", key))
		}
	}

	return issues
}

// isSensitiveKey checks if a key contains sensitive information
func (sm *SecurityManager) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd", "secret", "key", "token", "auth",
		"api_key", "apikey", "access_token", "bearer", "authorization",
		"client_secret", "private_key", "session_secret",
	}

	keyLower := strings.ToLower(key)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}
	return false
}

// sanitizeValue recursively sanitizes values
func (sm *SecurityManager) sanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return sm.SanitizeString(val)
	case map[string]interface{}:
		return sm.SanitizeMap(val)
	case []interface{}:
		var sanitized []interface{}
		for _, item := range val {
			sanitized = append(sanitized, sm.sanitizeValue(item))
		}
		return sanitized
	default:
		return val
	}
}

// MaskSecrets creates a display-safe version of secrets
func (sm *SecurityManager) MaskSecrets(secrets map[string]string) map[string]string {
	masked := make(map[string]string)
	for k, v := range secrets {
		if sm.isSensitiveKey(k) {
			if len(v) > 4 {
				masked[k] = v[:2] + strings.Repeat("*", len(v)-4) + v[len(v)-2:]
			} else {
				masked[k] = strings.Repeat("*", len(v))
			}
		} else {
			masked[k] = v
		}
	}
	return masked
}
