#!/bin/bash
set -e

echo "🚀 Setting up Obsidian Automation..."

# Create processor.go
cat > processor.go << 'EOF'
package main

import (
"fmt"
"log"
"os/exec"
"regexp"
"strings"

"github.com/ledongthuc/pdf"
)

type ProcessedContent struct {
Text       string
Category   string
Tags       []string
Confidence float64
Language   string
}

func extractTextFromImage(imagePath string) (string, error) {
cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "eng+fra+ara")
output, err := cmd.Output()
if err != nil {
cmd = exec.Command("tesseract", imagePath, "stdout")
output, err = cmd.Output()
if err != nil {
return "", fmt.Errorf("tesseract failed: %v", err)
}
}
return strings.TrimSpace(string(output)), nil
}

func extractTextFromPDF(pdfPath string) (string, error) {
cmd := exec.Command("pdftotext", pdfPath, "-")
output, err := cmd.Output()
if err == nil && len(output) > 0 {
return strings.TrimSpace(string(output)), nil
}

f, r, err := pdf.Open(pdfPath)
if err != nil {
return "", err
}
defer f.Close()

var text strings.Builder
for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
p := r.Page(pageNum)
if p.V.IsNull() {
continue
}
pageText, _ := p.GetPlainText(nil)
text.WriteString(pageText)
text.WriteString("\n\n")
}
return strings.TrimSpace(text.String()), nil
}

func classifyContent(text string) ProcessedContent {
text = strings.ToLower(text)
result := ProcessedContent{
Text:     text,
Category: "general",
Language: detectLanguage(text),
}

patterns := map[string][]string{
"physics":   {`force`, `energy`, `mass`, `velocity`, `acceleration`},
"math":      {`equation`, `function`, `derivative`, `integral`, `matrix`},
"chemistry": {`molecule`, `atom`, `reaction`, `chemical`},
"admin":     {`invoice`, `contract`, `form`, `certificate`},
}

scores := make(map[string]int)
for category, pats := range patterns {
scores[category] = countMatches(text, pats)
}

maxScore := 0
for cat, score := range scores {
if score > maxScore {
maxScore = score
result.Category = cat
}
}

total := 0
for _, score := range scores {
total += score
}
if total > 0 {
result.Confidence = float64(maxScore) / float64(total)
}
if result.Confidence < 0.3 || maxScore < 2 {
result.Category = "general"
}

result.Tags = []string{result.Category}
return result
}

func countMatches(text string, patterns []string) int {
count := 0
for _, pattern := range patterns {
re := regexp.MustCompile(pattern)
count += len(re.FindAllString(text, -1))
}
return count
}

func detectLanguage(text string) string {
frWords := []string{"le", "la", "de", "et", "un"}
count := 0
for _, w := range frWords {
if strings.Contains(" "+text+" ", " "+w+" ") {
count++
}
}
if count > 3 {
return "french"
}
return "english"
}

func processFile(filePath, fileType string) ProcessedContent {
var text string
var err error

log.Printf("Processing %s: %s", fileType, filePath)

if fileType == "image" {
text, err = extractTextFromImage(filePath)
} else if fileType == "pdf" {
text, err = extractTextFromPDF(filePath)
}

if err != nil {
log.Printf("Error: %v", err)
return ProcessedContent{Category: "unprocessed", Tags: []string{"error"}}
}

if len(text) < 10 {
return ProcessedContent{Text: text, Category: "unclear", Tags: []string{"low-text"}, Confidence: 0.1}
}

return classifyContent(text)
}
EOF

# Create health.go
cat > health.go << 'EOF'
package main

import (
"fmt"
"log"
"net/http"
"time"
)

var lastActivity time.Time

func startHealthServer() {
lastActivity = time.Now()
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
uptime := time.Since(lastActivity)
fmt.Fprintf(w, `{"status": "healthy", "uptime": "%s"}`, uptime)
})
log.Println("Health check at :8080/health")
go http.ListenAndServe(":8080", nil)
}

func updateActivity() {
lastActivity = time.Now()
}
EOF

# Create stats.go
cat > stats.go << 'EOF'
package main

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

func (s *Stats) load() {
data, err := os.ReadFile("stats.json")
if err != nil {
return
}
json.Unmarshal(data, s)
}
EOF

# Create dedup.go
cat > dedup.go << 'EOF'
package main

import (
"crypto/sha256"
"encoding/hex"
"io"
"os"
)

var processedHashes = make(map[string]string)

func getFileHash(filePath string) (string, error) {
f, err := os.Open(filePath)
if err != nil {
return "", err
}
defer f.Close()
h := sha256.New()
io.Copy(h, f)
return hex.EncodeToString(h.Sum(nil)), nil
}

func isDuplicate(filePath string) bool {
hash, err := getFileHash(filePath)
if err != nil {
return false
}
if _, exists := processedHashes[hash]; exists {
return true
}
processedHashes[hash] = filePath
return false
}
EOF

echo "✅ All files created!"
echo "Now run: go mod tidy && go build"
