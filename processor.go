package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Summary    string
	Topics     []string
	Questions  []string
	AIProvider string
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
		Text:       text,
		Category:   "general",
		Language:   detectLanguage(text),
		AIProvider: "None",
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

func processFileWithAI(filePath, fileType string, aiService *AIService) ProcessedContent {
	// Do basic OCR/extraction first
	var text string
	var err error
	var fileData []byte

	log.Printf("Processing %s: %s", fileType, filePath)

	if fileType == "image" {
		text, err = extractTextFromImage(filePath)
		if err == nil {
			fileData, err = ioutil.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading image file: %v", err)
			}
		}
	} else if fileType == "pdf" {
		text, err = extractTextFromPDF(filePath)
	}

	result := ProcessedContent{
		Text:     text,
		Category: "general",
		Tags:     []string{},
	}

	if err != nil || len(text) < 10 {
		log.Printf("Text extraction issue: %v", err)
		result.Category = "unprocessed"
		result.Tags = []string{"error"}
		return result
	}

	if aiService != nil {
		log.Println("Using Gemini for AI enhancement...")
		result.AIProvider = "Gemini"

		prompt := fmt.Sprintf(`Analyze the following text and return a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "summary": a 2-3 sentence summary of the text.
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.

Here is the text to analyze:
%s`, text)

		model := ModelProComplex
		if len(fileData) > 0 {
			model = ModelImageGen
		}

		// Use ModelProComplex for better analysis
		aiResponse, err := aiService.GenerateContent(context.Background(), prompt, fileData, model)
		if err != nil {
			log.Printf("Error from AI service: %v", err)
			// Fallback to basic classification
			return classifyContent(text)
		}

		// Parse the JSON response from the AI
		var aiResult struct {
			Category  string   `json:"category"`
			Summary   string   `json:"summary"`
			Topics    []string `json:"topics"`
			Questions []string `json:"questions"`
		}

		// Clean up the response string before unmarshalling
		aiResponse = strings.TrimPrefix(aiResponse, "```json")
		aiResponse = strings.TrimSuffix(aiResponse, "```")
		aiResponse = strings.TrimSpace(aiResponse)

		if err := json.Unmarshal([]byte(aiResponse), &aiResult); err != nil {
			log.Printf("Error parsing AI response: %v", err)
			// Fallback to basic classification if parsing fails
			return classifyContent(text)
		}

		result.Category = aiResult.Category
		result.Summary = aiResult.Summary
		result.Topics = aiResult.Topics
		result.Questions = aiResult.Questions
		result.Tags = append([]string{result.Category}, result.Topics...)
		result.Confidence = 0.95 // Gemini is very reliable

	} else {
		// Fallback to basic classification
		log.Println("AI service unavailable, using basic classification")
		result = classifyContent(text)
	}

	result.Language = detectLanguage(text)

	return result
}
