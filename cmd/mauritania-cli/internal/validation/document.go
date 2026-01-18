package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type DocumentType string

const (
	TypeREADME      DocumentType = "README"
	TypeAPI         DocumentType = "API"
	TypeGuide       DocumentType = "GUIDE"
	TypeCodeComment DocumentType = "CODE_COMMENT"
)

type Document struct {
	ID                string       `json:"id"`
	Type              DocumentType `json:"type"`
	Location          string       `json:"location"`
	LastUpdated       time.Time    `json:"last_updated"`
	CompletenessScore float64      `json:"completeness_score"`
	CoverageFeature   string       `json:"coverage_feature"`
	MaintainedBy      string       `json:"maintained_by"`
}

type DocumentTracker struct {
	dataDir string
}

func NewDocumentTracker(dataDir string) *DocumentTracker {
	return &DocumentTracker{dataDir: dataDir}
}

func (dt *DocumentTracker) LoadDocuments() ([]Document, error) {
	filePath := filepath.Join(dt.dataDir, "documents.json")
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Document{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var documents []Document
	if err := json.NewDecoder(file).Decode(&documents); err != nil {
		return nil, err
	}
	return documents, nil
}

func (dt *DocumentTracker) SaveDocuments(documents []Document) error {
	filePath := filepath.Join(dt.dataDir, "documents.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(documents)
}
