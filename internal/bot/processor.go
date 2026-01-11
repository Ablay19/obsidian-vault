package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/utils"
	"obsidian-automation/internal/vision"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go.uber.org/zap"
	"obsidian-automation/internal/vectorstore"
)

// Global vector store for RAG functionality
var globalVectorStore vectorstore.VectorStore

// Processor interface for the processing pipeline.
type Processor interface {
	Process(ctx context.Context, job pipeline.Job) (pipeline.Result, error)
}

type botProcessor struct {
	aiService   ai.AIServiceInterface
	vectorStore vectorstore.VectorStore
	visionProc  *vision.VisionProcessor
}

// NewBotProcessor creates a new botProcessor instance.
func NewBotProcessor(aiService ai.AIServiceInterface) Processor {
	return &botProcessor{
		aiService:  aiService,
		visionProc: vision.NewVisionProcessor(aiService),
	}
}

// Process implements the Processor interface for botProcessor.
func (p *botProcessor) Process(ctx context.Context, job pipeline.Job) (pipeline.Result, error) {
	streamCallback := func(chunk string) {
		zap.S().Debug("Stream chunk", "chunk", chunk)
	}

	// Extract bot and chat info for progress updates
	var updateStatus func(string)
	if bot, ok := job.Metadata["bot"].(Bot); ok {
		if chatID, ok := job.Metadata["chat_id"].(int64); ok {
			if messageID, ok := job.Metadata["message_id"].(int); ok {
				updateStatus = func(statusMsg string) {
					zap.S().Info("Processing status", "status", statusMsg)
					// Send progress update to user
					editMsg := tgbotapi.NewEditMessageText(chatID, messageID, statusMsg)
					bot.Send(editMsg)
				}
			}
		}
	}

	// Fallback to logging only if bot context not available
	if updateStatus == nil {
		updateStatus = func(statusMsg string) {
			zap.S().Info("Processing status", "status", statusMsg)
		}
	}

	caption, _ := job.Metadata["caption"].(string)

	// Use multi-strategy processing to avoid bad results
	processedContent := p.processFileWithMultipleStrategies(
		ctx,
		job.FileLocalPath,
		job.ContentType.String(),
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

// ProcessingStrategy represents different ways to process a file
type ProcessingStrategy struct {
	Name        string
	Description string
	ProcessFunc func(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), caption string) ProcessedContent
}

// processFileWithMultipleStrategies tries multiple processing approaches to avoid bad results
func (p *botProcessor) processFileWithMultipleStrategies(
	ctx context.Context,
	filePath, fileType string,
	streamCallback func(string),
	language string,
	updateStatus func(string),
	caption string,
) ProcessedContent {

	strategies := []ProcessingStrategy{
		{
			Name:        "Vision + AI Fusion",
			Description: "Advanced multimodal processing with vision encoder",
			ProcessFunc: p.processWithVisionFusion,
		},
		{
			Name:        "Primary AI Processing",
			Description: "Full AI analysis with OCR, summarization, and categorization",
			ProcessFunc: p.processWithPrimaryAI,
		},
		{
			Name:        "Enhanced OCR + Basic AI",
			Description: "Advanced OCR with simplified AI analysis",
			ProcessFunc: p.processWithEnhancedOCR,
		},
		{
			Name:        "Fallback Basic Processing",
			Description: "Basic OCR and text analysis only",
			ProcessFunc: p.processWithBasicFallback,
		},
	}

	var bestResult ProcessedContent
	var bestScore float64 = -1

	updateStatus("🤖 Trying multiple processing strategies...")

	for i, strategy := range strategies {
		updateStatus(fmt.Sprintf("🎯 Strategy %d/%d: %s", i+1, len(strategies), strategy.Name))

		result := strategy.ProcessFunc(ctx, filePath, fileType, p.aiService, streamCallback, language, func(s string) {
			// Don't spam with sub-strategy updates
		}, caption)

		// Score the result quality
		score := p.scoreProcessingResult(result)

		zap.S().Info("Strategy completed", "strategy", strategy.Name, "score", score, "category", result.Category)

		// Keep the best result
		if score > bestScore {
			bestScore = score
			bestResult = result
			bestResult.AIProvider = fmt.Sprintf("%s (%s)", result.AIProvider, strategy.Name)

			// If we have a good result, we can stop early
			if score >= 0.8 {
				zap.S().Info("Found good result, stopping strategy search", "strategy", strategy.Name, "score", score)
				break
			}
		}

		// Safety check - don't try too many strategies for performance
		if i >= 2 && bestScore >= 0.5 {
			break
		}
	}

	updateStatus(fmt.Sprintf("✅ Best result: %s (score: %.2f)", bestResult.AIProvider, bestScore))
	return bestResult
}

// scoreProcessingResult evaluates the quality of a processing result
func (p *botProcessor) scoreProcessingResult(result ProcessedContent) float64 {
	score := 0.0

	// Text quality (40% weight)
	if len(result.Text) > 100 {
		score += 0.4
	} else if len(result.Text) > 50 {
		score += 0.2
	} else if len(result.Text) > 10 {
		score += 0.1
	}

	// Summary quality (30% weight)
	if len(result.Summary) > 100 {
		score += 0.3
	} else if len(result.Summary) > 50 {
		score += 0.15
	}

	// Categorization quality (20% weight)
	if result.Category != "general" && result.Category != "unclear" && result.Category != "unprocessed" {
		score += 0.2
	}

	// Topics and questions (10% weight)
	if len(result.Topics) > 0 || len(result.Questions) > 0 {
		score += 0.1
	}

	return score
}

// processWithVisionFusion - Advanced multimodal processing with vision encoder and progress tracking
func (p *botProcessor) processWithVisionFusion(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), caption string) ProcessedContent {
	// Check if vision processing is available and enabled
	if p.visionProc == nil || !p.visionProc.IsAvailable() {
		zap.S().Info("Vision processing not available, falling back to primary AI")
		return p.processWithPrimaryAI(ctx, filePath, fileType, aiService, streamCallback, language, updateStatus, caption)
	}

	// Only process images for now (can extend to PDFs later)
	if fileType != "image" {
		zap.S().Debug("Vision fusion only supports images, using primary AI", "fileType", fileType)
		return p.processWithPrimaryAI(ctx, filePath, fileType, aiService, streamCallback, language, updateStatus, caption)
	}

	// Initialize progress tracker
	tracker := utils.CreateImageProcessingTracker()

	// Stage 1: Upload/Validation (already complete)
	tracker.SetCurrent("upload")
	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	tracker.SetCurrent("validation")
	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Stage 2: OCR Extraction
	tracker.SetCurrent("ocr_extraction")
	updateStatus(tracker.RenderCurrent())

	var extractedText string
	var enhancedText string
	var err error

	// Progress through OCR steps
	tracker.GetCurrent().Update(25)
	updateStatus(tracker.RenderCurrent())

	extractedText, err = extractTextFromImageEnhanced(filePath)
	if err != nil || len(extractedText) < 10 {
		zap.S().Warn("Enhanced OCR failed for vision processing, using basic OCR")
		tracker.GetCurrent().Update(50)
		updateStatus("Enhanced OCR failed, trying basic OCR... " + tracker.RenderCurrent())

		extractedText, err = extractTextFromImageBasic(filePath)
		if err != nil {
			tracker.GetCurrent().Update(0) // Failed
			updateStatus("OCR extraction failed")
			return ProcessedContent{Category: "unprocessed", Tags: []string{"vision_failed", "ocr_failed"}, AIProvider: "Vision + AI Fusion"}
		}
	}

	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Stage 3: Vision Encoding
	tracker.SetCurrent("vision_encoding")
	updateStatus(tracker.RenderCurrent())

	// Process with vision encoder
	multimodalEmbedding, err := p.visionProc.ProcessImage(ctx, filePath, extractedText)
	if err != nil {
		zap.S().Warn("Vision processing failed, falling back to primary AI", "error", err)
		updateStatus("Vision encoding failed, falling back to standard processing")
		return p.processWithPrimaryAI(ctx, filePath, fileType, aiService, streamCallback, language, updateStatus, caption)
	}

	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Get the enhanced text from vision processing
	enhancedText = extractedText // Default fallback

	// Check confidence threshold
	minConfidence := config.AppConfig.Vision.MinConfidence
	if multimodalEmbedding.Confidence < minConfidence {
		zap.S().Info("Vision confidence too low, falling back to primary AI", "confidence", multimodalEmbedding.Confidence, "threshold", minConfidence)
		updateStatus(fmt.Sprintf("Vision confidence (%.2f) below threshold (%.2f), using standard AI", multimodalEmbedding.Confidence, minConfidence))
		return p.processWithPrimaryAI(ctx, filePath, fileType, aiService, streamCallback, language, updateStatus, caption)
	}

	// Stage 4: Multimodal Fusion
	tracker.SetCurrent("multimodal_fusion")
	updateStatus(tracker.RenderCurrent())

	// Fusion is already done in ProcessImage, just mark as complete
	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Stage 5: AI Analysis
	tracker.SetCurrent("ai_analysis")
	updateStatus(tracker.RenderCurrent())

	// Create enhanced AI analysis with multimodal context
	result := ProcessedContent{
		Text:     extractedText,
		Category: "general",
		Tags:     []string{"vision_enhanced"},
		Language: language,
	}

	// Use LangChain with multimodal context
	result.AIProvider = fmt.Sprintf("%s (Vision Enhanced)", p.visionProc.GetEncoderName())

	tracker.GetCurrent().Update(50)
	updateStatus("Generating summary... " + tracker.RenderCurrent())

	result.Summary = p.generateMultimodalSummary(ctx, multimodalEmbedding, enhancedText, aiService, streamCallback)

	tracker.GetCurrent().Update(75)
	updateStatus("Extracting topics and questions... " + tracker.RenderCurrent())

	result.Topics, result.Questions = p.generateMultimodalTopicsAndQuestions(ctx, multimodalEmbedding, enhancedText, aiService)

	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Stage 6: Enhanced categorization
	result.Category = p.categorizeWithVision(multimodalEmbedding, enhancedText)
	result.Tags = append(result.Tags, result.Category)
	result.Confidence = multimodalEmbedding.Confidence

	// Update result text with enhanced version
	result.Text = enhancedText

	// Stage 7: Summarization (already done)
	tracker.SetCurrent("summarization")
	tracker.GetCurrent().Complete()
	updateStatus(tracker.RenderCurrent())

	// Stage 8: Storage
	tracker.SetCurrent("storage")
	updateStatus(tracker.RenderCurrent())

	// Store in vector store with multimodal embeddings
	if globalVectorStore != nil {
		docID := fmt.Sprintf("vision_%s_%s", fileType, filePath)
		doc := vectorstore.Document{
			ID:      docID,
			Content: result.Summary,
			Metadata: map[string]interface{}{
				"file_path":         filePath,
				"file_type":         fileType,
				"category":          result.Category,
				"language":          result.Language,
				"ai_provider":       result.AIProvider,
				"vision_confidence": multimodalEmbedding.Confidence,
				"vision_encoder":    p.visionProc.GetEncoderName(),
				"processing_time":   time.Since(tracker.GetBar("upload").GetStartTime()).Seconds(),
			},
			Vector: multimodalEmbedding.FusedVector,
		}

		if err := globalVectorStore.AddDocuments(ctx, []vectorstore.Document{doc}); err != nil {
			zap.S().Error("Failed to store vision document in vector store", "error", err)
			updateStatus("Vector storage failed, but processing complete")
		} else {
			zap.S().Info("Vision-enhanced document stored in vector store", "id", docID)
		}
	}

	tracker.GetCurrent().Complete()
	totalTime := time.Since(tracker.GetBar("upload").GetStartTime()).Seconds()
	updateStatus(fmt.Sprintf("Processing complete in %.1fs - %s", totalTime, tracker.RenderCurrent()))

	return result
}

// generateMultimodalSummary creates a summary using multimodal context
func (p *botProcessor) generateMultimodalSummary(ctx context.Context, embedding vision.MultimodalEmbedding, text string, aiService ai.AIServiceInterface, streamCallback func(string)) string {
	prompt := fmt.Sprintf(`Based on the multimodal analysis of an image and its extracted text, provide a comprehensive summary.

Image Analysis: The image has been analyzed using advanced vision AI and contains semantic information about visual content, layout, and structure.

Extracted Text: %s

Key multimodal insights:
- Vision confidence: %.2f
- Content appears to be document/image type
- Visual and textual information has been fused for enhanced understanding

Please provide a detailed summary that combines both visual and textual understanding:`, text, embedding.Confidence)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.3,
	}

	var response strings.Builder
	err := aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
		if streamCallback != nil {
			streamCallback(chunk)
		}
	})

	if err != nil {
		zap.S().Error("Failed to generate multimodal summary", "error", err)
		return "Summary generation failed due to AI service error."
	}

	return strings.TrimSpace(response.String())
}

// generateMultimodalTopicsAndQuestions generates topics, questions, and answers using multimodal context
func (p *botProcessor) generateMultimodalTopicsAndQuestions(ctx context.Context, embedding vision.MultimodalEmbedding, text string, aiService ai.AIServiceInterface) ([]string, []string) {
	// First, generate topics and questions
	topicsPrompt := fmt.Sprintf(`Analyze this content using multimodal understanding (vision + text) and provide topics and questions.

Content: %s

Vision Analysis: Advanced AI vision processing has analyzed the visual elements, layout, and structure with %.2f confidence.

Provide a JSON response with:
- "topics": array of key topics/themes identified
- "questions": array of insightful questions about the content

Focus on questions that show deep understanding of both visual and textual elements.`, text, embedding.Confidence)

	req := &ai.RequestModel{
		UserPrompt:  topicsPrompt,
		Temperature: 0.4,
	}

	var response strings.Builder
	err := aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		zap.S().Error("Failed to generate multimodal topics/questions", "error", err)
		return []string{}, []string{}
	}

	// Parse JSON response
	var topicsResult struct {
		Topics    []string `json:"topics"`
		Questions []string `json:"questions"`
	}

	jsonStr := strings.TrimSpace(response.String())
	if err := json.Unmarshal([]byte(jsonStr), &topicsResult); err != nil {
		zap.S().Warn("Failed to parse multimodal topics/questions JSON", "error", err)
		return []string{}, []string{}
	}

	// Now generate answers for the questions
	if len(topicsResult.Questions) > 0 {
		answersPrompt := fmt.Sprintf(`Based on the content and your understanding, provide detailed answers to these questions:

Content: %s

Questions to answer:
%s

Provide a JSON response with:
- "answers": array of detailed answers corresponding to each question

Be thorough and use the content to support your answers.`, text, p.formatQuestionsForAnswering(topicsResult.Questions))

		answerReq := &ai.RequestModel{
			UserPrompt:  answersPrompt,
			Temperature: 0.3, // Lower temperature for more accurate answers
		}

		var answerResponse strings.Builder
		err := aiService.Chat(ctx, answerReq, func(chunk string) {
			answerResponse.WriteString(chunk)
		})

		if err == nil {
			var answersResult struct {
				Answers []string `json:"answers"`
			}

			answerJsonStr := strings.TrimSpace(answerResponse.String())
			if err := json.Unmarshal([]byte(answerJsonStr), &answersResult); err == nil {
				// Combine questions and answers for display
				combinedQuestions := make([]string, len(topicsResult.Questions))
				for i, question := range topicsResult.Questions {
					answer := ""
					if i < len(answersResult.Answers) {
						answer = answersResult.Answers[i]
					}
					combinedQuestions[i] = fmt.Sprintf("Q: %s\nA: %s", question, answer)
				}
				return topicsResult.Topics, combinedQuestions
			}
		}
	}

	return topicsResult.Topics, topicsResult.Questions
}

// formatQuestionsForAnswering formats questions as a numbered list for the AI to answer
func (p *botProcessor) formatQuestionsForAnswering(questions []string) string {
	var formatted strings.Builder
	for i, question := range questions {
		formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, question))
	}
	return formatted.String()
}

// categorizeWithVision categorizes content using vision-enhanced analysis
func (p *botProcessor) categorizeWithVision(embedding vision.MultimodalEmbedding, text string) string {
	// Use pattern matching with enhanced context from vision
	textLower := strings.ToLower(text)

	// Enhanced patterns considering both text and visual context
	patterns := map[string][]string{
		"technical": {"code", "api", "function", "algorithm", "programming", "software", "database", "server", "network", "api", "json", "python", "javascript", "go", "rust"},
		"business":  {"meeting", "project", "strategy", "revenue", "business", "market", "client", "sales", "financial", "report", "presentation", "quarterly", "annual"},
		"academic":  {"research", "study", "analysis", "paper", "university", "academic", "thesis", "experiment", "literature", "review", "methodology", "conclusion"},
		"personal":  {"note", "reminder", "personal", "diary", "thought", "todo", "schedule", "appointment", "recipe", "shopping", "list"},
		"document":  {"report", "document", "file", "record", "information", "data", "record", "form", "contract", "agreement", "certificate", "invoice", "receipt"},
		"image":     {"photo", "picture", "image", "screenshot", "diagram", "chart", "graph", "photo", "scan", "photograph"},
		"pdf":       {"pdf", "document", "form", "application", "contract", "agreement", "certificate", "manual", "guide", "book"},
	}

	scores := make(map[string]int)
	for category, pats := range patterns {
		scores[category] = countMatches(textLower, pats)
	}

	// Boost scores based on vision confidence for certain categories
	if embedding.Confidence > 0.8 {
		// High confidence vision might indicate charts/diagrams
		if scores["image"] > 0 || scores["technical"] > 0 {
			scores["technical"] += 2
		}
	}

	maxScore := 0
	bestCategory := "general"
	for cat, score := range scores {
		if score > maxScore {
			maxScore = score
			bestCategory = cat
		}
	}

	if maxScore < 2 {
		bestCategory = "general"
	}

	return bestCategory
}

// processWithPrimaryAI - Full AI processing with OCR and analysis
func (p *botProcessor) processWithPrimaryAI(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), caption string) ProcessedContent {
	return processFileWithAI(ctx, filePath, fileType, aiService, streamCallback, language, updateStatus, caption)
}

// processWithEnhancedOCR - Enhanced OCR with simplified AI
func (p *botProcessor) processWithEnhancedOCR(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), caption string) ProcessedContent {
	// Extract text with enhanced OCR
	var text string
	var err error

	if fileType == "image" {
		text, err = extractTextFromImageEnhanced(filePath)
	} else if fileType == "pdf" {
		text, err = extractTextFromPDFEnhanced(filePath)
	}

	if err != nil || len(text) < 10 {
		return ProcessedContent{Category: "unprocessed", Tags: []string{"ocr_failed"}, AIProvider: "Enhanced OCR"}
	}

	// Simplified AI analysis - just categorization and basic summary
	result := ProcessedContent{
		Text:     text,
		Category: "general",
		Tags:     []string{},
		Language: language,
	}

	// Try basic AI categorization
	if aiService != nil {
		prompt := fmt.Sprintf("Categorize this text into one of: technical, business, academic, personal, document, image, pdf. Return only the category name.\n\nText: %s", text[:min(500, len(text))])

		req := &ai.RequestModel{
			UserPrompt:  prompt,
			Temperature: 0.3,
		}

		var response strings.Builder
		err := aiService.Chat(ctx, req, func(chunk string) {
			response.WriteString(chunk)
		})

		if err == nil {
			category := strings.ToLower(strings.TrimSpace(response.String()))
			validCategories := []string{"technical", "business", "academic", "personal", "document", "image", "pdf"}
			for _, valid := range validCategories {
				if strings.Contains(category, valid) {
					result.Category = valid
					break
				}
			}
		}
	}

	result.Tags = []string{result.Category}
	result.AIProvider = "Enhanced OCR + Basic AI"
	return result
}

// processWithBasicFallback - Basic OCR only, no AI
func (p *botProcessor) processWithBasicFallback(ctx context.Context, filePath, fileType string, aiService ai.AIServiceInterface, streamCallback func(string), language string, updateStatus func(string), caption string) ProcessedContent {
	// Basic processing without AI
	var text string
	var err error

	if fileType == "image" {
		text, err = extractTextFromImageBasic(filePath)
	} else if fileType == "pdf" {
		text, err = extractTextFromPDFBasic(filePath)
	}

	result := ProcessedContent{
		Text:     text,
		Category: "general",
		Tags:     []string{"basic_processing"},
		Language: language,
	}

	if err != nil {
		result.Category = "error"
		result.Tags = []string{"processing_failed"}
	}

	result.AIProvider = "Basic OCR Only"
	return result
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
	zap.S().Info("Starting advanced OCR text extraction from image", "path", imagePath)

	// Enhanced preprocessing for scanned documents
	if err := preprocessImageForOCR(imagePath); err != nil {
		zap.S().Warn("Image preprocessing failed, proceeding with original", "path", imagePath, "error", err)
	}

	var extracted string
	var lastError error

	// Try multiple OCR strategies in order of preference
	strategies := []struct {
		name string
		args []string
	}{
		{"English PSM6", []string{"-l", "eng", "--psm", "6", "--oem", "3"}},      // Uniform block of text
		{"Multi-lang PSM6", []string{"-l", "eng+fra+ara+deu+spa", "--psm", "6"}}, // Multi-language support
		{"English PSM3", []string{"-l", "eng", "--psm", "3", "--oem", "3"}},      // Fully automatic
		{"English PSM12", []string{"-l", "eng", "--psm", "12", "--oem", "3"}},    // Sparse text
		{"Default", []string{}}, // Fallback
	}

	for _, strategy := range strategies {
		args := append([]string{imagePath, "stdout"}, strategy.args...)
		cmd := execCommand("tesseract", args...)

		output, err := cmd.Output()
		if err == nil {
			text := strings.TrimSpace(string(output))
			if !isGarbledText(text) && len(text) > 10 {
				extracted = text
				zap.S().Info("OCR successful", "strategy", strategy.name, "text_len", len(text))
				break
			}
			zap.S().Debug("OCR strategy returned garbled or empty text", "strategy", strategy.name)
		} else {
			lastError = err
			zap.S().Debug("OCR strategy failed", "strategy", strategy.name, "error", err)
		}
	}

	if extracted == "" {
		zap.S().Error("All OCR strategies failed", "path", imagePath, "last_error", lastError)
		return "", fmt.Errorf("OCR failed with all strategies, last error: %v", lastError)
	}

	// Post-process the extracted text
	extracted = postProcessOCRText(extracted)

	zap.S().Info("Advanced OCR extraction complete", "path", imagePath, "text_len", len(extracted))
	return extracted, nil
}

// preprocessImageForOCR enhances image for better OCR using ImageMagick
func preprocessImageForOCR(imagePath string) error {
	// Create a temporary preprocessed image
	preprocessedPath := imagePath + "_ocr"

	// Multi-step preprocessing for scanned documents
	// 1. Increase resolution and contrast
	// 2. Convert to grayscale
	// 3. Apply unsharp mask for sharpness
	// 4. Normalize and auto-level
	cmd := execCommand("convert", imagePath,
		"-resize", "200%", // Increase resolution
		"-colorspace", "Gray", // Convert to grayscale
		"-contrast-stretch", "0", // Auto contrast
		"-normalize",      // Normalize
		"-unsharp", "0x1", // Sharpen
		"-threshold", "50%", // Binarize for better OCR
		preprocessedPath)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// Replace original with preprocessed
	return execCommand("mv", preprocessedPath, imagePath).Run()
}

// postProcessOCRText cleans up common OCR errors and formatting issues
func postProcessOCRText(text string) string {
	// Remove excessive whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Fix common OCR character mistakes
	replacements := map[string]string{
		"l":  "l", // Keep as is, but could be "1" or "I"
		"0":  "O", // Could be "0" or "O"
		"1":  "I", // Could be "1" or "I"
		"|":  "I", // Pipe often mistaken for I
		"rn": "m", // rn often becomes m
		"vv": "w", // vv often becomes w
	}

	for old, new := range replacements {
		// Only replace if it's clearly a mistake (surrounded by letters)
		pattern := fmt.Sprintf(`(\w)%s(\w)`, regexp.QuoteMeta(old))
		text = regexp.MustCompile(pattern).ReplaceAllStringFunc(text, func(match string) string {
			parts := regexp.MustCompile(pattern).FindStringSubmatch(match)
			if len(parts) == 3 {
				return parts[1] + new + parts[2]
			}
			return match
		})
	}

	// Fix line breaks in the middle of sentences
	text = regexp.MustCompile(`([a-z])\s*\n\s*([a-z])`).ReplaceAllString(text, "$1 $2")

	// Remove single character lines that are likely OCR artifacts
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 1 || (len(line) == 1 && regexp.MustCompile(`[a-zA-Z0-9]`).MatchString(line)) {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// isGarbledText checks if OCR result is mostly random characters
func isGarbledText(text string) bool {
	if len(text) < 5 {
		return false
	}

	// Count non-alphanumeric characters (excluding common punctuation)
	nonAlnum := 0
	for _, r := range text {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '.' || r == ',' || r == '!' || r == '?' || r == '-' || r == '\n') {
			nonAlnum++
		}
	}

	// If more than 40% non-alphanumeric, consider it garbled
	garbledRatio := float64(nonAlnum) / float64(len(text))
	isGarbled := garbledRatio > 0.4

	if isGarbled {
		zap.S().Debug("Text detected as garbled", "ratio", garbledRatio, "text_sample", text[:min(50, len(text))])
	}

	return isGarbled
}

// assessTextQuality evaluates the quality of extracted text
func assessTextQuality(text string) int {
	if len(text) < 10 {
		return 0
	}

	score := 0

	// Length bonus
	if len(text) > 100 {
		score += 30
	} else if len(text) > 50 {
		score += 20
	} else if len(text) > 25 {
		score += 10
	}

	// Word count bonus (prefer text with actual words)
	words := strings.Fields(text)
	wordCount := len(words)
	if wordCount > 20 {
		score += 25
	} else if wordCount > 10 {
		score += 15
	} else if wordCount > 5 {
		score += 5
	}

	// Sentence structure bonus (sentences with proper punctuation)
	sentences := strings.Split(text, ".")
	if len(sentences) > 3 {
		score += 20
	} else if len(sentences) > 1 {
		score += 10
	}

	// English word recognition (simple heuristic)
	englishWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
	englishCount := 0
	textLower := strings.ToLower(text)
	for _, word := range englishWords {
		if strings.Contains(textLower, " "+word+" ") {
			englishCount++
		}
	}
	score += englishCount * 2

	// Penalty for excessive symbols
	symbols := 0
	for _, r := range text {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == ' ' || r == '.' || r == ',' || r == '!' || r == '?' || r == '-' || r == '\n' ||
			r == ':' || r == ';' || r == '(' || r == ')' || r == '[' || r == ']' || r == '"' || r == '\'') {
			symbols++
		}
	}
	symbolRatio := float64(symbols) / float64(len(text))
	if symbolRatio > 0.3 {
		score -= 20
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// extractTextFromImageEnhanced - Enhanced OCR with multiple strategies and tools
func extractTextFromImageEnhanced(imagePath string) (string, error) {
	zap.S().Info("Starting enhanced OCR extraction", "path", imagePath)

	var allTexts []string

	// Strategy 1: Multiple Tesseract PSM modes
	psmStrategies := [][]string{
		{"-l", "eng", "--psm", "6", "--oem", "3"},         // Uniform block
		{"-l", "eng+fra+deu", "--psm", "3", "--oem", "3"}, // Multi-lang auto
		{"-l", "eng", "--psm", "12", "--oem", "3"},        // Sparse text
		{"-l", "eng", "--psm", "1", "--oem", "3"},         // Auto OSD
		{"-l", "eng", "--psm", "4", "--oem", "3"},         // Column finding
		{"-l", "eng", "--psm", "8", "--oem", "3"},         // Word finding
	}

	for _, strategy := range psmStrategies {
		args := append([]string{imagePath, "stdout"}, strategy...)
		cmd := execCommand("tesseract", args...)
		output, err := cmd.Output()
		if err == nil {
			text := strings.TrimSpace(string(output))
			if len(text) > 10 && !isGarbledText(text) {
				allTexts = append(allTexts, text)
			}
		}
	}

	// Strategy 2: Try with different image preprocessing
	preprocessedPath := imagePath + "_enhanced"
	defer os.Remove(preprocessedPath) // Clean up

	// Create enhanced version
	cmd := execCommand("convert", imagePath,
		"-resize", "300%", // Higher resolution
		"-colorspace", "Gray", // Grayscale
		"-contrast-stretch", "0", // Auto contrast
		"-normalize",      // Normalize
		"-unsharp", "0x1", // Sharpen
		"-threshold", "60%", // Adaptive threshold
		preprocessedPath)

	if cmd.Run() == nil {
		// Try OCR on enhanced image
		cmd := execCommand("tesseract", preprocessedPath, "stdout", "-l", "eng", "--psm", "6")
		if output, err := cmd.Output(); err == nil {
			text := strings.TrimSpace(string(output))
			if len(text) > 10 && !isGarbledText(text) {
				allTexts = append(allTexts, text)
			}
		}
	}

	// Strategy 3: Try with different OCR engines if available
	// Could add support for different OCR engines here

	// Select best result
	var bestText string
	maxQuality := 0

	for _, text := range allTexts {
		quality := assessTextQuality(text)
		if quality > maxQuality {
			maxQuality = quality
			bestText = text
		}
	}

	if bestText == "" {
		return "", fmt.Errorf("all enhanced OCR strategies failed")
	}

	// Post-process the best result
	bestText = postProcessOCRText(bestText)
	zap.S().Info("Enhanced OCR completed", "strategies_tried", len(psmStrategies)+1, "quality_score", maxQuality)
	return bestText, nil
}

// extractTextFromPDFEnhanced - Enhanced PDF extraction
func extractTextFromPDFEnhanced(pdfPath string) (string, error) {
	zap.S().Info("Starting enhanced PDF extraction", "path", pdfPath)

	// Try pdftotext first
	cmd := execCommand("pdftotext", pdfPath, "-")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		text := strings.TrimSpace(string(output))
		if len(text) > 100 { // Good result
			return text, nil
		}
	}

	// Fallback to go-pdf library with enhanced processing
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var fullText strings.Builder
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		// Extract text from page
		pageText, _ := p.GetPlainText(nil)

		if pageText != "" {
			fullText.WriteString(pageText)
			fullText.WriteString("\n\n")
		}
	}

	extracted := strings.TrimSpace(fullText.String())
	if len(extracted) < 10 {
		return "", fmt.Errorf("enhanced PDF extraction yielded insufficient text")
	}

	return extracted, nil
}

// extractTextFromImageBasic - Basic OCR fallback
func extractTextFromImageBasic(imagePath string) (string, error) {
	zap.S().Info("Starting basic OCR extraction", "path", imagePath)
	cmd := execCommand("tesseract", imagePath, "stdout")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractTextFromPDFBasic - Basic PDF extraction fallback
func extractTextFromPDFBasic(pdfPath string) (string, error) {
	zap.S().Info("Starting basic PDF extraction", "path", pdfPath)

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

	// Store document in vector store for RAG
	if globalVectorStore != nil {
		docID := fmt.Sprintf("%s_%s", fileType, filePath)
		embedding := generateEmbedding(ctx, aiService, result.Summary+" "+result.Text)

		doc := vectorstore.Document{
			ID:      docID,
			Content: result.Summary,
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"file_type":   fileType,
				"category":    result.Category,
				"language":    result.Language,
				"ai_provider": result.AIProvider,
			},
			Vector: embedding,
		}

		err := globalVectorStore.AddDocuments(ctx, []vectorstore.Document{doc})
		if err != nil {
			zap.S().Error("Failed to store document in vector store", "error", err)
		} else {
			zap.S().Info("Document stored in vector store for RAG", "id", docID, "embedding_dim", len(embedding))
		}
	}

	return result
}

// generateEmbedding creates proper embeddings using AI service
func generateEmbedding(ctx context.Context, aiService ai.AIServiceInterface, text string) []float32 {
	if aiService == nil {
		zap.S().Warn("AI service unavailable, falling back to simple embeddings")
		return generateSimpleEmbedding(text)
	}

	// Try to use AI service for embeddings first
	embedding, err := generateAIEmbedding(ctx, aiService, text)
	if err != nil {
		zap.S().Warn("AI embedding failed, falling back to simple embeddings", "error", err)
		return generateSimpleEmbedding(text)
	}

	return embedding
}

// generateAIEmbedding creates embeddings using AI service (OpenAI/Sentence Transformers style)
func generateAIEmbedding(ctx context.Context, aiService ai.AIServiceInterface, text string) ([]float32, error) {
	// Use a prompt that asks for embeddings in a format we can parse
	prompt := fmt.Sprintf("Generate a 384-dimensional embedding vector for the following text as a JSON array of floats:\n\nText: %s\n\nReturn only the JSON array, nothing else.", text)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.0, // Deterministic for embeddings
	}

	var response strings.Builder
	err := aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, fmt.Errorf("AI embedding request failed: %w", err)
	}

	// Try to parse the response as JSON array
	result := response.String()
	result = strings.TrimSpace(result)

	// Clean up potential markdown formatting
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var embedding []float32
	if err := json.Unmarshal([]byte(result), &embedding); err != nil {
		return nil, fmt.Errorf("failed to parse AI embedding response: %w", err)
	}

	// Validate embedding dimensions
	if len(embedding) != 384 {
		zap.S().Warn("AI returned embedding with unexpected dimensions", "expected", 384, "got", len(embedding))
	}

	// Normalize the embedding
	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	zap.S().Info("Generated AI embedding", "dimensions", len(embedding))
	return embedding, nil
}

// generateSimpleEmbedding creates a basic embedding vector from text (fallback)
func generateSimpleEmbedding(text string) []float32 {
	// Simple hash-based embedding for demonstration
	// Used as fallback when AI embeddings fail
	const embeddingDim = 384 // Common embedding dimension
	embedding := make([]float32, embeddingDim)

	// Simple hash function to generate pseudo-random values
	for i, char := range text {
		hash := int(char) * (i + 1)
		idx := hash % embeddingDim
		embedding[idx] += float32(hash%100) / 100.0
	}

	// Normalize the embedding
	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding
}
