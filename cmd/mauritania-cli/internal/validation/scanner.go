package validation

import (
	"os"
	"path/filepath"
	"strings"
)

type FeatureScanner struct {
	root string
}

func NewFeatureScanner(root string) *FeatureScanner {
	return &FeatureScanner{root: root}
}

func (fs *FeatureScanner) ScanFeatures() ([]Feature, error) {
	var features []Feature
	err := filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && fs.isFeatureDir(path) {
			feature := Feature{
				ID:          strings.ReplaceAll(path, "/", "-"),
				Name:        filepath.Base(path),
				Description: "Auto-discovered feature",
				Module:      path,
				Status:      StatusPending,
			}
			features = append(features, feature)
		}
		return nil
	})
	return features, err
}

func (fs *FeatureScanner) isFeatureDir(path string) bool {
	// Simple logic: has _test.go files or is internal package
	if strings.Contains(path, "internal") || strings.Contains(path, "cmd") {
		return true
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_test.go") {
			return true
		}
	}
	return false
}
