package bot

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileHash(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "dedup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Valid file", func(t *testing.T) {
		content := []byte("hello world")
		filePath := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		wantHash := sha256.Sum256(content)
		wantString := hex.EncodeToString(wantHash[:])

		got, err := getFileHash(filePath)
		if err != nil {
			t.Fatalf("getFileHash failed: %v", err)
		}

		if got != wantString {
			t.Errorf("getFileHash() = %s, want %s", got, wantString)
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := getFileHash(filepath.Join(tempDir, "nonexistent.txt"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}
