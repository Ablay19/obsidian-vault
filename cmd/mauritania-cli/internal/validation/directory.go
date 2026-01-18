package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Directory struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Purpose     string    `json:"purpose"`
	FileCount   int       `json:"file_count"`
	LastCleaned time.Time `json:"last_cleaned"`
	Module      string    `json:"module"`
}

type DirectoryTracker struct {
	dataDir string
}

func NewDirectoryTracker(dataDir string) *DirectoryTracker {
	return &DirectoryTracker{dataDir: dataDir}
}

func (dt *DirectoryTracker) LoadDirectories() ([]Directory, error) {
	filePath := filepath.Join(dt.dataDir, "directories.json")
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Directory{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var directories []Directory
	if err := json.NewDecoder(file).Decode(&directories); err != nil {
		return nil, err
	}
	return directories, nil
}

func (dt *DirectoryTracker) SaveDirectories(directories []Directory) error {
	filePath := filepath.Join(dt.dataDir, "directories.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(directories)
}
