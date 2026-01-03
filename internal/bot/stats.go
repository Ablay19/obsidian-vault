package bot

import (
	"encoding/json"
	"os"
	"sync"
)

type Stats struct {
	TotalFiles int            `json:"total_files"`
	ImageCount int            `json:"image_count"`
	PDFCount   int            `json:"pdf_count"`
	Categories map[string]int `json:"categories"`
	mu         sync.Mutex
}

var stats = Stats{Categories: make(map[string]int)}

func (s *Stats) recordFile(fileType, category string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalFiles++
	if fileType == "image" {
		s.ImageCount++
	} else if fileType == "pdf" {
		s.PDFCount++
	}
	s.Categories[category]++
	s.save()
}

func (s *Stats) save() {
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile("stats.json", data, 0644)
}

func (s *Stats) Load() {
	data, err := os.ReadFile("stats.json")
	if err != nil {
		return
	}
	json.Unmarshal(data, s)
}
