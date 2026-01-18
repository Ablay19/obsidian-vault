package docs

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"obsidian-automation/cmd/mauritania-cli/internal/validation"
)

type DocAnalyzer struct {
	logger *log.Logger
}

func NewDocAnalyzer(logger *log.Logger) *DocAnalyzer {
	return &DocAnalyzer{logger: logger}
}

func (da *DocAnalyzer) ScanDocuments(root string) ([]validation.Document, error) {
	var documents []validation.Document
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".md") {
			docType := da.determineDocType(path)
			doc := validation.Document{
				ID:                path,
				Type:              docType,
				Location:          path,
				LastUpdated:       info.ModTime(),
				CompletenessScore: 0.5,
				CoverageFeature:   "general",
				MaintainedBy:      "team",
			}
			documents = append(documents, doc)
		}
		return nil
	})
	return documents, err
}

func (da *DocAnalyzer) determineDocType(path string) validation.DocumentType {
	if strings.Contains(path, "README") {
		return validation.TypeREADME
	}
	if strings.Contains(path, "api") {
		return validation.TypeAPI
	}
	if strings.Contains(path, "guide") {
		return validation.TypeGuide
	}
	return validation.TypeCodeComment
}

func (da *DocAnalyzer) AnalyzeCompleteness(doc *validation.Document) {
	if doc.Type == validation.TypeREADME {
		doc.CompletenessScore = 0.8
	} else {
		doc.CompletenessScore = 0.6
	}
}

func (da *DocAnalyzer) FindGaps(documents []validation.Document, features []validation.Feature) []string {
	var gaps []string
	featureMap := make(map[string]bool)
	for _, f := range features {
		featureMap[f.ID] = true
	}
	docMap := make(map[string]bool)
	for _, d := range documents {
		docMap[d.CoverageFeature] = true
	}
	for id := range featureMap {
		if !docMap[id] {
			gaps = append(gaps, "Missing documentation for feature: "+id)
		}
	}
	return gaps
}
