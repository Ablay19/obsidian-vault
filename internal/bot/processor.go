package bot

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/pipeline"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"go.uber.org/zap"
)

// Processor interface for the processing pipeline.
type Processor interface {
	Process(ctx context.Context, job pipeline.Job) (pipeline.Result, error)
}

type botProcessor struct {
	aiService ai.AIServiceInterface
}

// NewBotProcessor creates a new botProcessor instance.
func NewBotProcessor(aiService ai.AIServiceInterface) Processor {
	return &botProcessor{
		aiService: aiService,
	}
}

// Process implements the Processor interface for botProcessor.
func (p *botProcessor) Process(ctx context.Context, job pipeline.Job) (pipeline.Result, error) {
	streamCallback := func(chunk string) {
		zap.S().Debug("Stream chunk", "chunk", chunk)
	}
	updateStatus := func(statusMsg string) {
		zap.S().Info("Processing status", "status", statusMsg)
	}

	caption, _ := job.Metadata["caption"].(string)

	processedContent := processFileWithAI(
		ctx,
		job.FileLocalPath,
		job.ContentType.String(),
		p.aiService,
		streamCallback,
		job.UserContext.Language,
		updateStatus,
		caption,
	)

	if processedContent.Category == "unprocessed" || processedContent.Category == "error" {
		return pipeline.Result{
			JobID: job.ID,
			Success: false,
			Error: fmt.Errorf("file processing failed: %s", processedContent.Category),
			ProcessedAt: time.Now(),
			Output: processedContent, // Include processedContent even on error for debugging
		}, fmt.Errorf("file processing failed for job %s", job.ID)
	}

	return pipeline.Result{
		JobID: job.ID,
		Success: true,
		ProcessedAt: time.Now(),
		Output: processedContent,
	}, nil
}

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

var execCommand = exec.Command

func extractTextFromImage(imagePath string) (string, error) {
	zap.S().Info("Starting OCR text extraction from image", "path", imagePath)
	cmd := execCommand("tesseract", imagePath, "stdout", "-l", "eng+fra+ara")
	output, err := cmd.Output()
	if err != nil {
		zap.S().Warn("Tesseract failed with multi-language, retrying with default", "path", imagePath, "error", err)
		cmd = execCommand("tesseract", imagePath, "stdout")
		output, err = cmd.Output()
		if err != nil {
			zap.S().Error("Tesseract failed completely", "path", imagePath, "error", err)
			return "", fmt.Errorf("tesseract failed: %v", err)
		}
	}
	extracted := strings.TrimSpace(string(output))
	zap.S().Info("OCR extraction complete", "path", imagePath, "text_len", len(extracted))
	return extracted, nil
}

func extractTextFromPDF(pdfPath string) (string, error) {
	zap.S().Info("Starting text extraction from PDF", "path", pdfPath)
	cmd := execCommand("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		extracted := strings.TrimSpace(string(output))
		zap.S().Info("PDF text extraction complete (pdftotext)", "path", pdfPath, "text_len", len(extracted))
		return extracted, nil
	}

	zap.S().Warn("pdftotext failed or empty, falling back to go-pdf", "path", pdfPath, "error", err)
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		zap.S().Error("Failed to open PDF for extraction", "path", pdfPath, "error", err)
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
	extracted := strings.TrimSpace(text.String())
	zap.S().Info("PDF text extraction complete (go-pdf fallback)", "path", pdfPath, "text_len", len(extracted))
	return extracted, nil
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

	zap.S().Info("Processing file", "type", fileType, "path", filePath)

	if fileType == "image" {
		text, err = extractTextFromImage(filePath)
	} else if fileType == "pdf" {
		text, err = extractTextFromPDF(filePath)
	}

	if err != nil {
		zap.S().Error("Error processing file", "error", err)
		return ProcessedContent{Category: "unprocessed", Tags: []string{"error"}}
	}

	if len(text) < 10 {
		return ProcessedContent{Text: text, Category: "unclear", Tags: []string{"low-text"}, Confidence: 0.1}
	}

	return classifyContent(text)
}

func processFileWithAI(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), additionalContext string) ProcessedContent {
	// Do basic OCR/extraction first
	var text string
	var err error
	var fileData []byte

	zap.S().Info("Processing file with AI", "type", fileType, "path", filePath)
	updateStatus("🔍 Extracting text...")

	if fileType == "image" {
		text, err = extractTextFromImage(filePath)
		if err != nil {
			zap.S().Error("Error extracting text from image", "error", err)
			// Continue with image data only if text extraction fails
		}
		fileData, err = os.ReadFile(filePath)
		if err != nil {
			zap.S().Error("Error reading image file", "error", err)
			updateStatus("⚠️ Could not read image file.")
			return ProcessedContent{Category: "unprocessed", Tags: []string{"error", "read-error"}}
		}
	} else if fileType == "pdf" {
		text, err = extractTextFromPDF(filePath)
	}

	result := ProcessedContent{
		Text:       text,
		Category: "general",
		Tags:     []string{},
		Language: language, // Use the provided language
	}

	if err != nil || len(text) < 10 {
		zap.S().Warn("Text extraction issue", "error", err)
		result.Category = "unprocessed"
		result.Tags = []string{"error"}
		updateStatus("⚠️ Text extraction failed.")
		return result
	}

	if aiService != nil {
		zap.S().Info("Using AI for enhancement...")
		// Use result.AIProvider = aiService.GetActiveProviderName()
		result.AIProvider = aiService.GetActiveProviderName()
		updateStatus("🤖 Generating summary...")

		// 1. Get the summary (streaming)
		var summaryPrompt string
		if len(fileData) > 0 {
			summaryPrompt = fmt.Sprintf("Analyze the attached image and summarize its content in %s. If there is text in the image, use it as context. If there are any questions, answer them as part of the summary. Extracted text (if any):\n\n%s\n\nAdditional User Context:\n%s",
				language,
				text,
				additionalContext,
			)
		} else {
			summaryPrompt = fmt.Sprintf("Summarize the following text in %s. If the text contains any questions, answer them as part of the summary. Text:\n\n%s\n\nAdditional User Context:\n%s",
				language,
				text,
				additionalContext,
			)
		}

		// Prepare Request Model
		chatReq := &ai.RequestModel{
			UserPrompt: summaryPrompt,
			ImageData:  fileData,
			Temperature: 0.5,
		}

		var fullSummaryBuilder strings.Builder
		streamErr := aiService.Chat(ctx, chatReq, func(chunk string) {
			fullSummaryBuilder.WriteString(chunk)
			if streamCallback != nil {
				streamCallback(chunk)
			}
		})

		if streamErr != nil {
			zap.S().Error("Error from AI summary service", "error", streamErr)
			updateStatus("⚠️ AI summary failed. Falling back to basic classification.")
			return classifyContent(text)
		}
		result.Summary = fullSummaryBuilder.String()

		updateStatus("📊 Generating topics and questions...")
		
		// 2. Get the structured data
		analysisResult, err := aiService.AnalyzeTextWithParams(ctx, text, language, len(text), 1, 0.01)
		if err != nil {
			zap.S().Error("Error from AI analysis service", "error", err)
			updateStatus("⚠️ AI analysis failed. Using basic classification.")
			basicResult := classifyContent(text)
			result.Category = basicResult.Category
			result.Tags = basicResult.Tags
		} else {
			result.Category = analysisResult.Category
			result.Topics = analysisResult.Topics
			result.Questions = analysisResult.Questions
			result.Tags = append([]string{result.Category}, result.Topics...)
			result.Confidence = 0.95
		}

	} else {
		zap.S().Info("AI service unavailable, using basic classification")
		updateStatus("⚠️ AI service unavailable. Using basic classification.")
		result = classifyContent(text)
	}

	return result
}