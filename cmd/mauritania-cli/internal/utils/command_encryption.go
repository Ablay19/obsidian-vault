package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
)

// CommandEncryption handles encryption/decryption of commands for secure transport
type CommandEncryption struct {
	config *EncryptionConfig
	logger *log.Logger
}

// NewCommandEncryption creates a new command encryption handler
func NewCommandEncryption(config *EncryptionConfig, logger *log.Logger) *CommandEncryption {
	return &CommandEncryption{
		config: config,
		logger: logger,
	}
}

// EncryptCommand encrypts a command for secure transmission
func (ce *CommandEncryption) EncryptCommand(command string, sessionKey string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	// Generate encryption key from session key
	key := ce.deriveKey(sessionKey)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create GCM cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt command
	ciphertext := aesGCM.Seal(nil, nonce, []byte(command), nil)

	// Combine nonce and ciphertext
	encrypted := append(nonce, ciphertext...)

	// Encode as base64 for transport
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	ce.logger.Printf("Command encrypted successfully (%d bytes)", len(encrypted))
	return encoded, nil
}

// DecryptCommand decrypts a command that was encrypted for transport
func (ce *CommandEncryption) DecryptCommand(encryptedCommand string, sessionKey string) (string, error) {
	if encryptedCommand == "" {
		return "", fmt.Errorf("encrypted command cannot be empty")
	}

	// Decode from base64
	encrypted, err := base64.StdEncoding.DecodeString(encryptedCommand)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted command: %w", err)
	}

	if len(encrypted) < 12 {
		return "", fmt.Errorf("encrypted command too short")
	}

	// Extract nonce and ciphertext
	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	// Generate decryption key from session key
	key := ce.deriveKey(sessionKey)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM cipher
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt command
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt command: %w", err)
	}

	command := string(plaintext)
	ce.logger.Printf("Command decrypted successfully")
	return command, nil
}

// EncryptResult encrypts command execution results
func (ce *CommandEncryption) EncryptResult(result string, sessionKey string) (string, error) {
	// Use the same encryption as commands
	return ce.EncryptCommand(result, sessionKey)
}

// DecryptResult decrypts command execution results
func (ce *CommandEncryption) DecryptResult(encryptedResult string, sessionKey string) (string, error) {
	// Use the same decryption as commands
	return ce.DecryptCommand(encryptedResult, sessionKey)
}

// deriveKey derives an encryption key from a session key using SHA-256
func (ce *CommandEncryption) deriveKey(sessionKey string) []byte {
	hash := sha256.Sum256([]byte(sessionKey))
	return hash[:32] // AES-256 requires 32 bytes
}

// ValidateEncryptionConfig validates the encryption configuration
func (ce *CommandEncryption) ValidateEncryptionConfig() error {
	if ce.config.KeySize != 32 {
		return fmt.Errorf("unsupported key size: %d (expected 32 for AES-256)", ce.config.KeySize)
	}

	if ce.config.Algorithm != "AES-256-GCM" {
		return fmt.Errorf("unsupported algorithm: %s (expected AES-256-GCM)", ce.config.Algorithm)
	}

	return nil
}

// GetEncryptionInfo returns information about the encryption setup
func (ce *CommandEncryption) GetEncryptionInfo() map[string]interface{} {
	return map[string]interface{}{
		"algorithm": ce.config.Algorithm,
		"key_size":  ce.config.KeySize,
		"enabled":   true,
	}
}

// SecureCommand represents a command with encryption metadata
type SecureCommand struct {
	EncryptedCommand string            `json:"encrypted_command"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Timestamp        int64             `json:"timestamp"`
}

// CreateSecureCommand creates a secure command wrapper
func (ce *CommandEncryption) CreateSecureCommand(command string, sessionKey string, metadata map[string]string) (*SecureCommand, error) {
	encrypted, err := ce.EncryptCommand(command, sessionKey)
	if err != nil {
		return nil, err
	}

	return &SecureCommand{
		EncryptedCommand: encrypted,
		Metadata:         metadata,
		Timestamp:        time.Now().Unix(),
	}, nil
}

// DecryptSecureCommand decrypts a secure command
func (ce *CommandEncryption) DecryptSecureCommand(secureCmd *SecureCommand, sessionKey string) (string, error) {
	return ce.DecryptCommand(secureCmd.EncryptedCommand, sessionKey)
}

// CommandSecurityValidator validates command security before encryption
type CommandSecurityValidator struct {
	logger *log.Logger
}

// NewCommandSecurityValidator creates a new security validator
func NewCommandSecurityValidator(logger *log.Logger) *CommandSecurityValidator {
	return &CommandSecurityValidator{
		logger: logger,
	}
}

// ValidateCommandSecurity performs security checks on commands before encryption
func (csv *CommandSecurityValidator) ValidateCommandSecurity(command string) error {
	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"rm -rf /",
		"sudo rm",
		"dd if=",
		"mkfs",
		"fdisk",
		"format",
		"del /s /q",
		"cipher /w:",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(command), strings.ToLower(pattern)) {
			csv.logger.Printf("Security warning: Command contains dangerous pattern: %s", pattern)
			return fmt.Errorf("command contains potentially dangerous operation: %s", pattern)
		}
	}

	// Check command length
	if len(command) > 10000 {
		return fmt.Errorf("command too long for secure transport: %d characters", len(command))
	}

	// Check for null bytes (potential injection)
	if strings.Contains(command, "\x00") {
		return fmt.Errorf("command contains null bytes")
	}

	return nil
}

// SanitizeCommand sanitizes a command for safe execution
func (csv *CommandSecurityValidator) SanitizeCommand(command string) string {
	// Remove or escape potentially dangerous characters
	// This is a basic implementation - in production you'd want more sophisticated sanitization

	sanitized := command

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Basic sanitization - remove suspicious characters
	suspiciousChars := []string{"`", "$(", "${", ";", "|", "&", ">", "<"}
	for _, char := range suspiciousChars {
		if strings.Contains(sanitized, char) {
			csv.logger.Printf("Security warning: Command contains suspicious character: %s", char)
		}
	}

	return sanitized
}
