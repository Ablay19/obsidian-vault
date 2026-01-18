package cleanup

import (
	"log"
	"os"
	"path/filepath"
)

type CleanupExecutor struct {
	logger *log.Logger
}

func NewCleanupExecutor(logger *log.Logger) *CleanupExecutor {
	return &CleanupExecutor{logger: logger}
}

func (ce *CleanupExecutor) IdentifyTargets(root string) ([]string, error) {
	var targets []string
	patterns := []string{"*.tmp", "*.bak", "*.log", "*.out", "coverage.out", "*.test"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, "**", pattern))
		if err != nil {
			return nil, err
		}
		targets = append(targets, matches...)
	}
	return targets, nil
}

func (ce *CleanupExecutor) ExecuteCleanup(targets []string, dryRun bool) error {
	for _, target := range targets {
		if dryRun {
			ce.logger.Printf("Would remove: %s", target)
		} else {
			if err := os.Remove(target); err != nil {
				ce.logger.Printf("Failed to remove %s: %v", target, err)
			} else {
				ce.logger.Printf("Removed: %s", target)
			}
		}
	}
	return nil
}
