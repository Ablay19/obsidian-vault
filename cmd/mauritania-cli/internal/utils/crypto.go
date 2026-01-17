package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	KeySize       int    // Size of encryption key in bytes
	SaltSize      int    // Size of salt in bytes
	Algorithm     string // "aes" or "chacha"
	KeyDerivation string // "scrypt" or "argon2"
}

// DefaultEncryptionConfig returns secure default encryption settings
func DefaultEncryptionConfig() EncryptionConfig {
	return EncryptionConfig{
		KeySize:       32, // 256-bit key
		SaltSize:      16, // 128-bit salt
		Algorithm:     "aes",
		KeyDerivation: "scrypt",
	}
}

// CryptoManager handles encryption and decryption operations
type CryptoManager struct {
	config    EncryptionConfig
	masterKey []byte
}

// NewCryptoManager creates a new crypto manager with the given master key
func NewCryptoManager(masterKey string, config EncryptionConfig) (*CryptoManager, error) {
	if len(masterKey) < 16 {
		return nil, fmt.Errorf("master key must be at least 16 characters long")
	}

	// Derive master key using key derivation function
	derivedKey, err := deriveKey([]byte(masterKey), config)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}

	return &CryptoManager{
		config:    config,
		masterKey: derivedKey,
	}, nil
}

// Encrypt encrypts plaintext data
func (cm *CryptoManager) Encrypt(plaintext string) (string, error) {
	if cm.masterKey == nil {
		return "", fmt.Errorf("crypto manager not initialized")
	}

	plaintextBytes := []byte(plaintext)

	// Generate random nonce/IV
	nonce := make([]byte, cm.getNonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(cm.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt
	ciphertext := aesgcm.Seal(nil, nonce, plaintextBytes, nil)

	// Combine nonce and ciphertext
	combined := append(nonce, ciphertext...)

	// Base64 encode for storage
	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts encrypted data
func (cm *CryptoManager) Decrypt(encrypted string) (string, error) {
	if cm.masterKey == nil {
		return "", fmt.Errorf("crypto manager not initialized")
	}

	// Base64 decode
	combined, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	nonceSize := cm.getNonceSize()
	if len(combined) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	// Extract nonce and ciphertext
	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	// Create cipher
	block, err := aes.NewCipher(cm.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword creates a secure hash of a password
func (cm *CryptoManager) HashPassword(password string) (string, error) {
	// Generate salt
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Combine salt and hash
	combined := append(salt, hash...)

	// Base64 encode
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifies a password against its hash
func (cm *CryptoManager) VerifyPassword(password, hash string) (bool, error) {
	// Decode hash
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	if len(decoded) < 16 {
		return false, fmt.Errorf("invalid hash format")
	}

	// Extract salt and hash
	salt := decoded[:16]
	storedHash := decoded[16:]

	// Hash the provided password
	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Compare hashes securely
	return compareHashes(computedHash, storedHash), nil
}

// GenerateSecureToken generates a cryptographically secure random token
func (cm *CryptoManager) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashString creates a SHA-256 hash of a string (for non-sensitive data)
func (cm *CryptoManager) HashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// EncryptCredentials encrypts credential data for storage
func (cm *CryptoManager) EncryptCredentials(credentials map[string]string) (map[string]string, error) {
	encrypted := make(map[string]string)

	for key, value := range credentials {
		encryptedKey, err := cm.Encrypt(key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credential key %s: %w", key, err)
		}

		encryptedValue, err := cm.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt credential value for %s: %w", key, err)
		}

		encrypted[encryptedKey] = encryptedValue
	}

	return encrypted, nil
}

// DecryptCredentials decrypts credential data
func (cm *CryptoManager) DecryptCredentials(encrypted map[string]string) (map[string]string, error) {
	decrypted := make(map[string]string)

	for encryptedKey, encryptedValue := range encrypted {
		key, err := cm.Decrypt(encryptedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential key: %w", err)
		}

		value, err := cm.Decrypt(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential value for %s: %w", key, err)
		}

		decrypted[key] = value
	}

	return decrypted, nil
}

// deriveKey derives an encryption key from a password using configured KDF
func deriveKey(password []byte, config EncryptionConfig) ([]byte, error) {
	salt := make([]byte, config.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	switch config.KeyDerivation {
	case "scrypt":
		return scrypt.Key(password, salt, 32768, 8, 1, config.KeySize)
	case "argon2":
		return argon2.IDKey(password, salt, 1, 64*1024, 4, uint32(config.KeySize)), nil
	default:
		return nil, fmt.Errorf("unsupported key derivation function: %s", config.KeyDerivation)
	}
}

// getNonceSize returns the nonce size for the configured algorithm
func (cm *CryptoManager) getNonceSize() int {
	switch cm.config.Algorithm {
	case "aes":
		return 12 // GCM nonce size
	case "chacha":
		return 24 // ChaCha20-Poly1305 nonce size
	default:
		return 12 // Default to GCM
	}
}

// compareHashes securely compares two hashes
func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// ValidateMasterKey validates that a master key meets security requirements
func ValidateMasterKey(key string) error {
	if len(key) < 16 {
		return fmt.Errorf("master key must be at least 16 characters long")
	}

	if len(key) > 128 {
		return fmt.Errorf("master key must not exceed 128 characters")
	}

	// Check for basic complexity (at least one uppercase, lowercase, digit)
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range key {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("master key must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	return nil
}
