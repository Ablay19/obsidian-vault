package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Status string

const (
	StatusValidated Status = "validated"
	StatusFailed    Status = "failed"
	StatusPending   Status = "pending"
)

type Feature struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Module        string     `json:"module"`
	Status        Status     `json:"status"`
	TestCoverage  float64    `json:"test_coverage"`
	LastValidated time.Time  `json:"last_validated"`
	TestCases     []TestCase `json:"test_cases"`
}

type TestCase struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	FeatureID string        `json:"feature_id"`
	Type      string        `json:"type"`
	Status    string        `json:"status"`
	LastRun   time.Time     `json:"last_run"`
	Duration  time.Duration `json:"duration"`
}

type FeatureTracker struct {
	dataDir string
}

func NewFeatureTracker(dataDir string) *FeatureTracker {
	return &FeatureTracker{dataDir: dataDir}
}

func (ft *FeatureTracker) LoadFeatures() ([]Feature, error) {
	filePath := filepath.Join(ft.dataDir, "features.json")
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Feature{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var features []Feature
	if err := json.NewDecoder(file).Decode(&features); err != nil {
		return nil, err
	}
	return features, nil
}

func (ft *FeatureTracker) SaveFeatures(features []Feature) error {
	filePath := filepath.Join(ft.dataDir, "features.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(features)
}
