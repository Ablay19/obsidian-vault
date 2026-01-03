package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var encryptionKey []byte

func init() {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Fallback for dev, but warn
		key = "change-me-at-least-32-chars-long-!!!!"
	}
	if len(key) < 32 {
		// Pad or truncate to 32 bytes for AES-256
		newKey := make([]byte, 32)
		copy(newKey, key)
		encryptionKey = newKey
	} else {
		encryptionKey = []byte(key[:32])
	}
}

// Encrypt encrypts plain text using AES-256 GCM.
func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded cipher text using AES-256 GCM.
func Decrypt(encodedCiphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
