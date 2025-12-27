package main

import (
"fmt"
"log"
"os"
"path/filepath"
"time"
)

// Auto-organize notes into category folders
func organizeNote(notePath string, category string) error {
// Don't organize unprocessed/general/unclear
if category == "unprocessed" || category == "general" || category == "unclear" || category == "error" {
return nil
}

// Create category folder if not exists
categoryDir := filepath.Join(vaultDir, category)
os.MkdirAll(categoryDir, 0755)

// Move note to category folder
baseName := filepath.Base(notePath)
newPath := filepath.Join(categoryDir, baseName)

// Wait a bit to ensure file is written
time.Sleep(100 * time.Millisecond)

err := os.Rename(notePath, newPath)
if err != nil {
return fmt.Errorf("failed to move note: %v", err)
}

log.Printf("✅ Organized: %s -> %s/", baseName, category)
return nil
}
