package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/pipeline"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
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
			JobID:       job.ID,
			Success:     false,
			Error:       fmt.Errorf("file processing failed: %s", processedContent.Category),
			ProcessedAt: time.Now(),
			Output:      processedContent, // Include processedContent even on error for debugging
		}, fmt.Errorf("file processing failed for job %s", job.ID)
	}

	return pipeline.Result{
		JobID:       job.ID,
		Success:     true,
		ProcessedAt: time.Now(),
		Output:      processedContent,
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

// AIServiceLLM wraps the aiService to implement LangChain's LLM interface
type AIServiceLLM struct {
	aiService ai.AIServiceInterface
	modelName string
}

func (l *AIServiceLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	// Use the aiService to generate content
	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.5,
	}
	var response strings.Builder
	err := l.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})
	if err != nil {
		return "", err
	}
	return response.String(), nil
}

func (l *AIServiceLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Simplified implementation for LangChain compatibility
	prompt := ""
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				prompt += textPart.Text
			}
		}
	}
	content, err := l.Call(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: content,
			},
		},
	}, nil
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

// createLangChainSummaryChain creates a LangChain chain for summarization
func createLangChainSummaryChain(aiService ai.AIServiceInterface) chains.Chain {
	llm := &AIServiceLLM{aiService: aiService, modelName: aiService.GetActiveProviderName()}
	prompt := prompts.NewPromptTemplate("Summarize the following text in {{.language}}. If the text contains any questions, answer them as part of the summary. Text:\n\n{{.text}}", []string{"language", "text"})
	return chains.NewLLMChain(llm, prompt)
}

// createLangChainAnalysisChain creates a LangChain chain for JSON analysis
func createLangChainAnalysisChain(aiService ai.AIServiceInterface) chains.Chain {
	llm := &AIServiceLLM{aiService: aiService, modelName: aiService.GetActiveProviderName()}
	prompt := prompts.NewPromptTemplate("Analyze the following text and provide a JSON response with 'category', 'topics' (array), and 'questions' (array). Text:\n\n{{.text}}", []string{"text"})
	return chains.NewLLMChain(llm, prompt)
}

func processFileWithAI(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), additionalContext string) ProcessedContent {
	// Do basic OCR/extraction first
	var text string
	var err error

	zap.S().Info("Processing file with AI", "type", fileType, "path", filePath)
	updateStatus("🔍 Extracting text...")

	if fileType == "image" {
		text, err = extractTextFromImage(filePath)
		if err != nil {
			zap.S().Error("Error extracting text from image", "error", err)
			// Continue with text only
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
		zap.S().Warn("Text extraction issue", "error", err)
		result.Category = "unprocessed"
		result.Tags = []string{"error"}
		updateStatus("⚠️ Text extraction failed.")
		return result
	}

	if aiService != nil {
		zap.S().Info("Using LangChain for enhancement...")
		result.AIProvider = aiService.GetActiveProviderName()
		updateStatus("🤖 Generating summary with LangChain...")

		// Create LangChain summary chain
		summaryChain := createLangChainSummaryChain(aiService)

		// Prepare inputs for the chain
		inputs := map[string]any{
			"language": language,
			"text":     text,
		}
		if additionalContext != "" {
			inputs["text"] = text + "\n\nAdditional User Context:\n" + additionalContext
		}

		// Run the chain
		summaryResult, err := chains.Call(ctx, summaryChain, inputs)
		if err != nil {
			zap.S().Error("Error from LangChain summary", "error", err)
			updateStatus("⚠️ LangChain summary failed. Falling back to basic classification.")
			return classifyContent(text)
		}

		// Extract summary from result
		if textResult, ok := summaryResult["text"].(string); ok {
			result.Summary = textResult
			// For streaming, we could split the result, but for now, send as one chunk
			if streamCallback != nil {
				streamCallback(result.Summary)
			}
		} else {
			zap.S().Warn("Unexpected summary result format", "result", summaryResult)
			result.Summary = "Summary generated but format unexpected"
		}

		updateStatus("📊 Generating topics and questions with LangChain...")

		// Create LangChain analysis chain
		analysisChain := createLangChainAnalysisChain(aiService)

		// Run the analysis chain
		analysisInputs := map[string]any{
			"text": text,
		}
		analysisResult, err := chains.Call(ctx, analysisChain, analysisInputs)
		if err != nil {
			zap.S().Error("Error from LangChain analysis", "error", err)
			updateStatus("⚠️ LangChain analysis failed. Using basic classification.")
			basicResult := classifyContent(text)
			result.Category = basicResult.Category
			result.Tags = basicResult.Tags
		} else {
			// Parse the JSON result
			if jsonStr, ok := analysisResult["text"].(string); ok {
				var aiResult struct {
					Category  string   `json:"category"`
					Topics    []string `json:"topics"`
					Questions []string `json:"questions"`
				}
				if parseErr := json.Unmarshal([]byte(jsonStr), &aiResult); parseErr != nil {
					zap.S().Error("Error parsing LangChain analysis JSON", "error", parseErr)
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
				zap.S().Warn("Unexpected analysis result format", "result", analysisResult)
				basicResult := classifyContent(text)
				result.Category = basicResult.Category
				result.Tags = basicResult.Tags
			}
		}

	} else {
		zap.S().Info("AI service unavailable, using basic classification")
		updateStatus("⚠️ AI service unavailable. Using basic classification.")
		result = classifyContent(text)
	}

	return result
}
