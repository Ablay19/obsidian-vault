package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
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

	patterns := config.AppConfig.Classification.Patterns

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
	frWords := config.AppConfig.LanguageDetection.FrenchWords
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

func processFileWithAI(filePath, fileType string, aiService *ai.AIService, streamCallback func(string), language string, updateStatus func(string)) ProcessedContent {
	// Do basic OCR/extraction first
	var text string
	var err error
	var fileData []byte

	log.Printf("Processing %s: %s", fileType, filePath)
	updateStatus("🔍 Extracting text...")

	if fileType == "image" {
		text, err = extractTextFromImage(filePath)
		if err != nil {
			log.Printf("Error extracting text from image: %v", err)
			// Continue with image data only if text extraction fails
		}
		fileData, err = ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading image file: %v", err)
			updateStatus("⚠️ Could not read image file.")
			return ProcessedContent{Category: "unprocessed", Tags: []string{"error", "read-error"}}
		}
	} else if fileType == "pdf" {
		text, err = extractTextFromPDF(filePath)
	}

	result := ProcessedContent{
		Text:     text,
		Category: "general",
		Tags:     []string{},
		Language: language, // Use the provided language
	}

	if err != nil || len(text) < 10 {
		log.Printf("Text extraction issue: %v", err)
		result.Category = "unprocessed"
		result.Tags = []string{"error"}
		updateStatus("⚠️ Text extraction failed.")
		return result
	}

	if aiService != nil {
		log.Println("Using AI for enhancement...")
		provider, _, err := aiService.GetActiveProvider(context.Background())
		if err != nil {
			log.Printf("Error getting active AI provider: %v", err)
			result.AIProvider = "None" // Fallback
		} else {
			result.AIProvider = provider.GetModelInfo().ProviderName
		}
		updateStatus("🤖 Generating summary...")

		// Determine model to use based on whether image data is present
		modelProvider, _, err := aiService.GetActiveProvider(context.Background())
		if err != nil {
			log.Printf("Error getting active AI provider for model selection: %v", err)
			return result // Or handle error appropriately
		}
		modelToUse := modelProvider.GetModelInfo().ModelName
		// If image data is present and the active provider is Gemini, we might want to use a vision-capable model
		// This logic needs to be refined based on actual model capabilities and configuration
		// For now, we'll just use the default configured model.

		// 1. Get the summary (streaming)
		var summaryPrompt string
		if len(fileData) > 0 {
			summaryPrompt = fmt.Sprintf("Analyze the attached image and summarize its content in %s. If there is text in the image, use it as context. If there are any questions, answer them as part of the summary. Extracted text (if any):\n\n%s",
				language,
				text,
			)
		} else {
			summaryPrompt = fmt.Sprintf("Summarize the following text in %s. If the text contains any questions, answer them as part of the summary. Text:\n\n%s",
				language,
				text,
			)
		}

		fullSummary, err := aiService.GenerateContent(context.Background(), summaryPrompt, fileData, modelToUse, streamCallback)
		if err != nil {
			log.Printf("Error from AI summary service: %v", err)
			// Fallback to basic classification if summary fails
			updateStatus("⚠️ AI summary failed. Falling back to basic classification.")
			return classifyContent(text)
		}
		result.Summary = fullSummary

		updateStatus("📊 Generating topics and questions...")
		// 2. Get the JSON data (non-streaming)
		jsonStr, err := aiService.GenerateJSONData(context.Background(), text, language)
		if err != nil {
			log.Printf("Error from AI JSON service: %v", err)
			// Proceed without JSON data, just use basic classification
			updateStatus("⚠️ AI analysis failed. Using basic classification.")
			basicResult := classifyContent(text)
			result.Category = basicResult.Category
			result.Tags = basicResult.Tags
			return result
		}

		var aiResult struct {
			Category  string   `json:"category"`
			Topics    []string `json:"topics"`
			Questions []string `json:"questions"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &aiResult); err != nil {
			log.Printf("Error parsing AI response JSON: %v", err)
			updateStatus("⚠️ AI response parsing failed. Using basic classification.")
			// Proceed without JSON data
			basicResult := classifyContent(text)
			result.Category = basicResult.Category
			result.Tags = basicResult.Tags
		} else {
			result.Category = aiResult.Category
			result.Topics = aiResult.Topics
			result.Questions = aiResult.Questions
			result.Tags = append([]string{result.Category}, result.Topics...)
			result.Confidence = 0.95
		}

	} else {
		log.Println("AI service unavailable, using basic classification")
		updateStatus("⚠️ AI service unavailable. Using basic classification.")
		result = classifyContent(text)
	}

	return result
}
