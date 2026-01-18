package cleanup

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

type CleanupAnalyzer struct {
	logger *log.Logger
}

func NewCleanupAnalyzer(logger *log.Logger) *CleanupAnalyzer {
	return &CleanupAnalyzer{logger: logger}
}

func (ca *CleanupAnalyzer) AnalyzeDirectories(root string) ([]validation.Directory, error) {
	var directories []validation.Directory
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir := validation.Directory{
				ID:        path,
				Path:      path,
				Purpose:   ca.determinePurpose(path),
				FileCount: ca.countFiles(path),
				Module:    ca.determineModule(path),
			}
			directories = append(directories, dir)
		}
		return nil
	})
	return directories, err
}

func (ca *CleanupAnalyzer) determinePurpose(path string) string {
	if strings.Contains(path, "cmd") {
		return "CLI application"
	}
	if strings.Contains(path, "internal") {
		return "Internal packages"
	}
	return "General"
}

func (ca *CleanupAnalyzer) determineModule(path string) string {
	if strings.Contains(path, "doppler") {
		return "doppler"
	}
	return "main"
}

func (ca *CleanupAnalyzer) countFiles(path string) int {
	count := 0
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
