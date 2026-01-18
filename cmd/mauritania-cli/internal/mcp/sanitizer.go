package mcp

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Sanitizer provides data sanitization utilities
type Sanitizer struct {
	sensitivePatterns []*regexp.Regexp
}

// NewSanitizer creates a new sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		sensitivePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(password|passwd|pwd|secret|key|token|auth)\s*[:=]\s*["']?[^"'\s]+["']?`),
			regexp.MustCompile(`(?i)(api_key|apikey|access_token)\s*[:=]\s*["']?[^"'\s]+["']?`),
			regexp.MustCompile(`(?i)bearer\s+[^"'\s]+`),
			regexp.MustCompile(`(?i)authorization\s*[:=]\s*["']?[^"'\s]+["']?`),
		},
	}
}

// SanitizeString removes sensitive information from a string
func (s *Sanitizer) SanitizeString(input string) string {
	result := input
	for _, pattern := range s.sensitivePatterns {
		result = pattern.ReplaceAllString(result, "$1: [REDACTED]")
	}
	return result
}

// SanitizeMap removes sensitive information from a map
func (s *Sanitizer) SanitizeMap(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	for k, v := range data {
		if s.isSensitiveKey(k) {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = s.sanitizeValue(v)
		}
	}
	return sanitized
}

// SanitizeJSON sanitizes a JSON byte array
func (s *Sanitizer) SanitizeJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data, err
	}
	sanitized := s.sanitizeValue(obj)
	return json.Marshal(sanitized)
}

// isSensitiveKey checks if a key contains sensitive information
func (s *Sanitizer) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd", "secret", "key", "token", "auth",
		"api_key", "apikey", "access_token", "bearer", "authorization",
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
func (s *Sanitizer) sanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return s.SanitizeString(val)
	case map[string]interface{}:
		return s.SanitizeMap(val)
	case []interface{}:
		var sanitized []interface{}
		for _, item := range val {
			sanitized = append(sanitized, s.sanitizeValue(item))
		}
		return sanitized
	default:
		return val
	}
}
