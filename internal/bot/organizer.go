package bot

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// organizeNote moves a note to the correct sub-directory based on its category.
func organizeNote(notePath string, category string) error {
	if category == "" || category == "general" {
		// No need to move, it stays in the Inbox
		return nil
	}

	// Sanitize category to create a valid directory name
	dirName := strings.Title(strings.ToLower(category))
	destDir := filepath.Join("vault", dirName)

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Move the file
	destPath := filepath.Join(destDir, filepath.Base(notePath))
	if err := os.Rename(notePath, destPath); err != nil {
		return fmt.Errorf("failed to move note from %s to %s: %w", notePath, destPath, err)
	}

	slog.Info("Organized note", "from", notePath, "to", destPath)
	return nil
}
