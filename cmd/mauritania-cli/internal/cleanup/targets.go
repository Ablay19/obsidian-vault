package cleanup

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CleanupTargets struct {
	logger *log.Logger
}

func NewCleanupTargets(logger *log.Logger) *CleanupTargets {
	return &CleanupTargets{logger: logger}
}

func (ct *CleanupTargets) IdentifyTargets(root string, gitignorePatterns []string) ([]string, error) {
	var targets []string
	patterns := []string{"*.tmp", "*.bak", "*.log", "*.out", "coverage.out", "*.test"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, "**", pattern))
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			if !ct.isProtected(match, gitignorePatterns) {
				targets = append(targets, match)
			}
		}
	}
	return targets, nil
}

func (ct *CleanupTargets) isProtected(path string, gitignorePatterns []string) bool {
	for _, pattern := range gitignorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (ct *CleanupTargets) LoadGitignore(root string) ([]string, error) {
	gitignorePath := filepath.Join(root, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return nil, err
	}
	var patterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns, nil
}
