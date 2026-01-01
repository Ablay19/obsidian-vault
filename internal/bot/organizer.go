package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func organizeNote(notePath string, category string) error {
	if category == "general" || category == "" {
		return nil // No need to organize general notes
	}

	// Sanitize category name to be filesystem-friendly
	safeCategory := strings.ReplaceAll(strings.Title(category), " ", "")

	// Create the destination directory if it doesn't exist
	destDir := filepath.Join(vaultDir, safeCategory)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", destDir, err)
	}

	// Move the note
	destPath := filepath.Join(destDir, filepath.Base(notePath))
	if err := os.Rename(notePath, destPath); err != nil {
		return fmt.Errorf("could not move note from %s to %s: %w", notePath, destPath, err)
	}

	log.Printf("Organized note: %s -> %s", notePath, destPath)
	return nil
}
