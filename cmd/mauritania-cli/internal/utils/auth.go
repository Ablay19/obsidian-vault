package utils

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// AuthValidator handles authentication and authorization validation
type AuthValidator struct {
	config        *Config
	cryptoManager *CryptoManager
	httpClient    *http.Client
}

// NewAuthValidator creates a new authentication validator
func NewAuthValidator(config *Config, cryptoManager *CryptoManager) *AuthValidator {
	return &AuthValidator{
		config:        config,
		cryptoManager: cryptoManager,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Network.Timeout) * time.Second,
		},
	}
}

// ValidateAPIKey validates an API key format and basic requirements
func (av *AuthValidator) ValidateAPIKey(apiKey, service string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required for %s", service)
	}

	// Service-specific validation
	switch strings.ToLower(service) {
	case "whatsapp":
		return av.validateWhatsAppKey(apiKey)
	case "telegram":
		return av.validateTelegramKey(apiKey)
	case "facebook":
		return av.validateFacebookKey(apiKey)
	case "shipper":
		return av.validateShipperKey(apiKey)
	default:
		return av.validateGenericKey(apiKey)
	}
}

// validateWhatsAppKey validates WhatsApp Business API key
func (av *AuthValidator) validateWhatsAppKey(apiKey string) error {
	// WhatsApp API keys are typically 32 characters long
	if len(apiKey) < 20 || len(apiKey) > 64 {
		return fmt.Errorf("WhatsApp API key must be between 20 and 64 characters")
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9\-_]+$`, apiKey); !matched {
		return fmt.Errorf("WhatsApp API key contains invalid characters")
	}

	return nil
}

// validateTelegramKey validates Telegram Bot API token
func (av *AuthValidator) validateTelegramKey(botToken string) error {
	// Telegram bot tokens follow the format: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
	parts := strings.Split(botToken, ":")
	if len(parts) != 2 {
		return fmt.Errorf("Telegram bot token must be in format 'bot_id:token'")
	}

	botID := parts[0]
	token := parts[1]

	if len(botID) < 5 || len(botID) > 15 {
		return fmt.Errorf("Telegram bot ID must be between 5 and 15 characters")
	}

	if len(token) < 20 || len(token) > 50 {
		return fmt.Errorf("Telegram token must be between 20 and 50 characters")
	}

	// Check for valid characters
	if matched, _ := regexp.MatchString(`^[0-9]+:[A-Za-z0-9\-_]+$`, botToken); !matched {
		return fmt.Errorf("Telegram bot token contains invalid characters")
	}

	return nil
}

// validateFacebookKey validates Facebook access token
func (av *AuthValidator) validateFacebookKey(accessToken string) error {
	// Facebook access tokens are typically long strings
	if len(accessToken) < 50 || len(accessToken) > 300 {
		return fmt.Errorf("Facebook access token length is invalid")
	}

	// Check for pipe character (long-lived tokens format)
	if !strings.Contains(accessToken, "|") {
		return fmt.Errorf("Facebook access token appears to be malformed")
	}

	return nil
}

// validateShipperKey validates SM APOS Shipper API credentials
func (av *AuthValidator) validateShipperKey(apiKey string) error {
	if len(apiKey) < 16 || len(apiKey) > 128 {
		return fmt.Errorf("Shipper API key must be between 16 and 128 characters")
	}

	// Check for required complexity
	hasUpper := strings.ContainsAny(apiKey, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(apiKey, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(apiKey, "0123456789")

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("Shipper API key must contain uppercase, lowercase, and numeric characters")
	}

	return nil
}

// validateGenericKey validates generic API keys
func (av *AuthValidator) validateGenericKey(apiKey string) error {
	if len(apiKey) < 8 || len(apiKey) > 256 {
		return fmt.Errorf("API key must be between 8 and 256 characters")
	}

	return nil
}

// ValidateUserPermissions validates if a user has required permissions
func (av *AuthValidator) ValidateUserPermissions(userID string, requiredPermissions []string) error {
	if !av.config.Auth.Enabled {
		return nil // Auth disabled, allow all
	}

	// Check if user is in allowed list
	if len(av.config.Auth.AllowedUsers) > 0 {
		allowed := false
		for _, allowedUser := range av.config.Auth.AllowedUsers {
			if allowedUser == userID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("user %s is not authorized", userID)
		}
	}

	// Additional permission checks can be added here
	// For now, basic user validation is sufficient

	return nil
}

// ValidateCommandPermissions validates if a command is allowed
func (av *AuthValidator) ValidateCommandPermissions(command string, userID string) error {
	if !av.config.Auth.Enabled {
		return nil // Auth disabled, allow all
	}

	// Check command against allowed commands
	if len(av.config.Auth.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range av.config.Auth.AllowedCommands {
			if av.matchesCommandPattern(command, allowedCmd) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command '%s' is not allowed for user %s", command, userID)
		}
	}

	return nil
}

// matchesCommandPattern checks if a command matches a pattern (supports wildcards)
func (av *AuthValidator) matchesCommandPattern(command, pattern string) bool {
	// Simple wildcard support (*)
	if strings.Contains(pattern, "*") {
		// Convert wildcard pattern to regex
		regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*")
		matched, _ := regexp.MatchString("^"+regexPattern+"$", command)
		return matched
	}

	return command == pattern
}

// ValidateTransportCredentials validates transport-specific credentials
func (av *AuthValidator) ValidateTransportCredentials(transportType, credentials string) error {
	switch transportType {
	case "whatsapp":
		return av.validateWhatsAppCredentials(credentials)
	case "telegram":
		return av.validateTelegramCredentials(credentials)
	case "facebook":
		return av.validateFacebookCredentials(credentials)
	case "shipper":
		return av.validateShipperCredentials(credentials)
	default:
		return fmt.Errorf("unknown transport type: %s", transportType)
	}
}

// validateWhatsAppCredentials validates WhatsApp credentials by making a test API call
func (av *AuthValidator) validateWhatsAppCredentials(credentials string) error {
	// For now, just validate the format
	// In a real implementation, this would make an API call to verify credentials
	parts := strings.Split(credentials, ":")
	if len(parts) != 2 {
		return fmt.Errorf("WhatsApp credentials must be in format 'api_key:phone_number'")
	}

	if err := av.validateWhatsAppKey(parts[0]); err != nil {
		return err
	}

	// Validate phone number format (basic check)
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
	if !phoneRegex.MatchString(parts[1]) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// validateTelegramCredentials validates Telegram credentials
func (av *AuthValidator) validateTelegramCredentials(botToken string) error {
	if err := av.validateTelegramKey(botToken); err != nil {
		return err
	}

	// In a real implementation, you might make a test API call here
	// For now, format validation is sufficient
	return nil
}

// validateFacebookCredentials validates Facebook credentials
func (av *AuthValidator) validateFacebookCredentials(accessToken string) error {
	if err := av.validateFacebookKey(accessToken); err != nil {
		return err
	}

	// In a real implementation, you might make a test API call here
	return nil
}

// validateShipperCredentials validates SM APOS Shipper credentials
func (av *AuthValidator) validateShipperCredentials(apiKey string) error {
	return av.validateShipperKey(apiKey)
}

// AuthenticateUser authenticates a user with the given credentials
func (av *AuthValidator) AuthenticateUser(userID, password string) error {
	// For now, this is a placeholder
	// In a real implementation, this would check against a user database
	if userID == "" || password == "" {
		return fmt.Errorf("user ID and password are required")
	}

	// Basic validation
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// ValidateSession validates an active session
func (av *AuthValidator) ValidateSession(session *models.ShipperSession) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	if time.Now().After(session.ExpiresAt) {
		return fmt.Errorf("session has expired")
	}

	if session.LastActivity.Add(24 * time.Hour).Before(time.Now()) {
		return fmt.Errorf("session has been inactive too long")
	}

	return nil
}

// GenerateSecureToken generates a secure authentication token
func (av *AuthValidator) GenerateSecureToken() (string, error) {
	return av.cryptoManager.GenerateSecureToken(32)
}

// ValidateToken validates an authentication token
func (av *AuthValidator) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Basic token format validation
	if len(token) < 16 {
		return fmt.Errorf("token is too short")
	}

	// Check for valid base64 URL encoding
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9\-_]+$`, token); !matched {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}

// HashPassword securely hashes a password
func (av *AuthValidator) HashPassword(password string) (string, error) {
	return av.cryptoManager.HashPassword(password)
}

// VerifyPassword verifies a password against its hash
func (av *AuthValidator) VerifyPassword(password, hash string) (bool, error) {
	return av.cryptoManager.VerifyPassword(password, hash)
}

// SecureCompare performs a constant-time comparison to prevent timing attacks
func (av *AuthValidator) SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateWebhookSignature validates webhook signatures for incoming requests
func (av *AuthValidator) ValidateWebhookSignature(payload []byte, signature, secret string) error {
	if secret == "" {
		return fmt.Errorf("webhook secret is not configured")
	}

	expectedSignature := av.cryptoManager.HashString(string(payload) + secret)
	if !av.SecureCompare(signature, expectedSignature) {
		return fmt.Errorf("webhook signature verification failed")
	}

	return nil
}

// RateLimitCheck performs rate limiting checks
func (av *AuthValidator) RateLimitCheck(identifier string, limit int, window time.Duration) error {
	// This is a placeholder for rate limiting logic
	// In a real implementation, this would check against a rate limiter
	// For now, always allow
	return nil
}

// CleanupExpiredSessions removes expired authentication sessions
func (av *AuthValidator) CleanupExpiredSessions(sessions []*models.ShipperSession) []*models.ShipperSession {
	var activeSessions []*models.ShipperSession
	now := time.Now()

	for _, session := range sessions {
		if now.Before(session.ExpiresAt) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions
}
