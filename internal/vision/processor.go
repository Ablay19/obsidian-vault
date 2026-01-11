package vision

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/mathocr"
	"obsidian-automation/internal/ocr"
	"strings"

	"go.uber.org/zap"
)

// VisionEncoder interface for encoding images into embeddings
type VisionEncoder interface {
	EncodeImage(ctx context.Context, imagePath string) ([]float32, error)
	EncodeWithText(ctx context.Context, imagePath, text string) (MultimodalEmbedding, error)
	IsAvailable() bool
	GetName() string
}

// MultimodalEmbedding represents fused vision and text embeddings
type MultimodalEmbedding struct {
	ImageVector []float32
	TextVector  []float32
	FusedVector []float32
	Confidence  float64
}

// FusionMethod represents different fusion strategies
type FusionMethod string

const (
	FusionConcatenation  FusionMethod = "concatenation"
	FusionCrossAttention FusionMethod = "cross_attention"
	FusionAverage        FusionMethod = "average"
	FusionWeighted       FusionMethod = "weighted"
)

// VisionProcessor handles multimodal image processing
type VisionProcessor struct {
	encoder       VisionEncoder
	fusionMethod  FusionMethod
	aiService     ai.AIServiceInterface
	minConfidence float64
	documentOCR   *ocr.DeepSeekOCR
	mathProcessor *mathocr.MathOCRProcessor
}

// NewVisionProcessor creates a new vision processor
func NewVisionProcessor(aiService ai.AIServiceInterface) *VisionProcessor {
	// Try to create the best available encoder
	var encoder VisionEncoder

	// Priority: API-based encoders (DeepSeek, Gemini with vision)
	if isProviderVisionCapable("deepseek", aiService) {
		encoder = NewDeepSeekVisionEncoder(aiService)
	} else if isProviderVisionCapable("gemini", aiService) {
		encoder = NewGeminiVisionEncoder(aiService)
	} else if isProviderVisionCapable("openai", aiService) {
		encoder = NewOpenAIVisionEncoder(aiService)
	}

	if encoder == nil || !encoder.IsAvailable() {
		zap.S().Warn("No vision encoder available, vision processing will be disabled")
		return nil
	}

	fusionMethod := FusionMethod(config.AppConfig.Vision.FusionMethod)
	if fusionMethod == "" {
		fusionMethod = FusionCrossAttention // Default
	}

	// Initialize DeepSeek OCR for document processing
	var documentOCR *ocr.DeepSeekOCR
	if isProviderVisionCapable("deepseek", aiService) {
		documentOCR = ocr.NewDeepSeekOCR(aiService)
	}

	// Initialize math OCR processor
	mathProcessor := mathocr.NewMathOCRProcessor()

	return &VisionProcessor{
		encoder:       encoder,
		fusionMethod:  fusionMethod,
		aiService:     aiService,
		minConfidence: config.AppConfig.Vision.MinConfidence,
		documentOCR:   documentOCR,
		mathProcessor: mathProcessor,
	}
}

// ProcessImage processes an image using vision capabilities
func (vp *VisionProcessor) ProcessImage(ctx context.Context, imagePath string, extractedText string) (MultimodalEmbedding, error) {
	if vp == nil {
		return MultimodalEmbedding{}, fmt.Errorf("vision processor not available")
	}

	zap.S().Info("Processing image with vision encoder", "encoder", vp.encoder.GetName(), "path", imagePath)

	// Enhanced text extraction for documents
	enhancedText := extractedText
	if vp.documentOCR != nil && vp.isDocumentImage(imagePath, extractedText) {
		zap.S().Info("Detected document image, using DeepSeek OCR for enhanced extraction")

		layout, err := vp.documentOCR.ProcessDocument(ctx, imagePath)
		if err == nil && layout.Confidence > 0.7 {
			enhancedText = vp.documentOCR.ExtractText(layout)

			// Apply mathematical OCR enhancements
			enhancedText, formulas := vp.mathProcessor.EnhanceOCROutput(enhancedText)
			if len(formulas) > 0 {
				zap.S().Info("Mathematical formulas detected and enhanced", "count", len(formulas))
			}

			zap.S().Info("Enhanced OCR completed", "original_len", len(extractedText), "enhanced_len", len(enhancedText), "confidence", layout.Confidence)
		} else {
			zap.S().Warn("DeepSeek OCR failed or low confidence, using basic text", "error", err)
		}
	} else {
		// Even for non-document images, apply basic mathematical enhancements
		var formulas []mathocr.MathFormula
		enhancedText, formulas = vp.mathProcessor.EnhanceOCROutput(extractedText)
		if len(formulas) > 0 {
			zap.S().Info("Mathematical content enhanced in general image", "formulas", len(formulas))
		}
	}

	// Generate multimodal embedding with enhanced text
	embedding, err := vp.encoder.EncodeWithText(ctx, imagePath, enhancedText)
	if err != nil {
		return MultimodalEmbedding{}, fmt.Errorf("vision encoding failed: %w", err)
	}

	// Apply fusion method
	embedding.FusedVector = vp.fuseEmbeddings(embedding.ImageVector, embedding.TextVector)

	// Calculate confidence based on embedding quality
	embedding.Confidence = vp.calculateEmbeddingConfidence(embedding)

	zap.S().Info("Vision processing completed",
		"image_dim", len(embedding.ImageVector),
		"text_dim", len(embedding.TextVector),
		"fused_dim", len(embedding.FusedVector),
		"confidence", embedding.Confidence)

	return embedding, nil
}

// fuseEmbeddings applies the configured fusion method
func (vp *VisionProcessor) fuseEmbeddings(imageVec, textVec []float32) []float32 {
	switch vp.fusionMethod {
	case FusionConcatenation:
		return vp.concatenateEmbeddings(imageVec, textVec)
	case FusionCrossAttention:
		return vp.crossAttentionFusion(imageVec, textVec)
	case FusionAverage:
		return vp.averageEmbeddings(imageVec, textVec)
	case FusionWeighted:
		return vp.weightedFusion(imageVec, textVec)
	default:
		zap.S().Warn("Unknown fusion method, using concatenation", "method", vp.fusionMethod)
		return vp.concatenateEmbeddings(imageVec, textVec)
	}
}

// concatenateEmbeddings simply concatenates image and text embeddings
func (vp *VisionProcessor) concatenateEmbeddings(imageVec, textVec []float32) []float32 {
	result := make([]float32, len(imageVec)+len(textVec))
	copy(result[:len(imageVec)], imageVec)
	copy(result[len(imageVec):], textVec)
	return result
}

// crossAttentionFusion implements simplified cross-attention fusion
func (vp *VisionProcessor) crossAttentionFusion(imageVec, textVec []float32) []float32 {
	// Simplified attention mechanism - in practice, this would use transformer layers
	minDim := len(imageVec)
	if len(textVec) < minDim {
		minDim = len(textVec)
	}

	result := make([]float32, minDim)

	// Compute attention weights (simplified dot-product attention)
	for i := 0; i < minDim; i++ {
		// Weighted combination based on relative magnitudes
		imageWeight := float32(0.6) // Bias toward visual features
		textWeight := float32(0.4)

		if len(imageVec) > i && len(textVec) > i {
			result[i] = imageWeight*imageVec[i] + textWeight*textVec[i]
		} else if len(imageVec) > i {
			result[i] = imageVec[i]
		} else if len(textVec) > i {
			result[i] = textVec[i]
		}
	}

	return result
}

// averageEmbeddings takes the element-wise average
func (vp *VisionProcessor) averageEmbeddings(imageVec, textVec []float32) []float32 {
	minDim := len(imageVec)
	if len(textVec) < minDim {
		minDim = len(textVec)
	}

	result := make([]float32, minDim)
	for i := 0; i < minDim; i++ {
		count := 0
		sum := float32(0)

		if i < len(imageVec) {
			sum += imageVec[i]
			count++
		}
		if i < len(textVec) {
			sum += textVec[i]
			count++
		}

		if count > 0 {
			result[i] = sum / float32(count)
		}
	}

	return result
}

// weightedFusion applies confidence-based weighting
func (vp *VisionProcessor) weightedFusion(imageVec, textVec []float32) []float32 {
	imageConf := vp.calculateVectorConfidence(imageVec)
	textConf := vp.calculateVectorConfidence(textVec)

	totalWeight := imageConf + textConf
	if totalWeight == 0 {
		return vp.averageEmbeddings(imageVec, textVec)
	}

	minDim := len(imageVec)
	if len(textVec) < minDim {
		minDim = len(textVec)
	}

	result := make([]float32, minDim)
	for i := 0; i < minDim; i++ {
		imageVal := float32(0)
		textVal := float32(0)

		if i < len(imageVec) {
			imageVal = imageVec[i]
		}
		if i < len(textVec) {
			textVal = textVec[i]
		}

		result[i] = (float32(imageConf)*imageVal + float32(textConf)*textVal) / float32(totalWeight)
	}

	return result
}

// calculateEmbeddingConfidence estimates embedding quality
func (vp *VisionProcessor) calculateEmbeddingConfidence(embedding MultimodalEmbedding) float64 {
	imageConf := vp.calculateVectorConfidence(embedding.ImageVector)
	textConf := vp.calculateVectorConfidence(embedding.TextVector)
	fusedConf := vp.calculateVectorConfidence(embedding.FusedVector)

	// Weighted average favoring the fused result
	return 0.5*fusedConf + 0.25*imageConf + 0.25*textConf
}

// calculateVectorConfidence calculates confidence for a single vector
func (vp *VisionProcessor) calculateVectorConfidence(vec []float32) float64 {
	if len(vec) == 0 {
		return 0.0
	}

	// Measure vector "usefulness" by variance and magnitude
	var sum, sumSq float32
	for _, v := range vec {
		sum += v
		sumSq += v * v
	}

	mean := sum / float32(len(vec))
	variance := (sumSq / float32(len(vec))) - (mean * mean)

	// Normalize variance to 0-1 range (rough heuristic)
	confidence := float64(variance * 10)
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1 // Minimum confidence
	}

	return confidence
}

// IsAvailable checks if vision processing is available
func (vp *VisionProcessor) IsAvailable() bool {
	return vp != nil && vp.encoder != nil && vp.encoder.IsAvailable()
}

// GetEncoderName returns the name of the current encoder
func (vp *VisionProcessor) GetEncoderName() string {
	if vp == nil || vp.encoder == nil {
		return "none"
	}
	return vp.encoder.GetName()
}

// isDocumentImage determines if an image is likely a document based on text content
func (vp *VisionProcessor) isDocumentImage(imagePath, extractedText string) bool {
	if len(extractedText) < 100 {
		return false
	}

	text := strings.ToLower(extractedText)
	lines := strings.Split(text, "\n")

	// Document indicators
	documentIndicators := 0

	// Check for structured text patterns
	if strings.Contains(text, "chapter") || strings.Contains(text, "section") {
		documentIndicators++
	}

	// Check for paragraphs (multiple lines with substantial content)
	paragraphCount := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 50 {
			paragraphCount++
		}
	}
	if paragraphCount >= 3 {
		documentIndicators++
	}

	// Check for formal language patterns
	formalWords := []string{"therefore", "however", "furthermore", "consequently", "accordingly", "moreover", "hence"}
	formalCount := 0
	for _, word := range formalWords {
		if strings.Contains(text, word) {
			formalCount++
		}
	}
	if formalCount >= 2 {
		documentIndicators++
	}

	// Check for list structures
	listPatterns := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "•") ||
			strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			listPatterns++
		}
	}
	if listPatterns >= 3 {
		documentIndicators++
	}

	// Consider it a document if we have multiple indicators
	return documentIndicators >= 2
}

// isProviderVisionCapable checks if an AI provider supports vision
func isProviderVisionCapable(providerName string, aiService ai.AIServiceInterface) bool {
	if aiService == nil {
		return false
	}

	activeProvider := strings.ToLower(aiService.GetActiveProviderName())

	// Get provider profile from config
	providerKey := strings.ToLower(activeProvider)
	if profile, exists := config.AppConfig.ProviderProfiles[providerKey]; exists {
		return profile.SupportsVision
	}

	// Fallback to name-based detection
	switch strings.ToLower(providerName) {
	case "deepseek":
		return strings.Contains(activeProvider, "deepseek")
	case "gemini":
		return strings.Contains(activeProvider, "gemini") || strings.Contains(activeProvider, "google")
	case "openai":
		return strings.Contains(activeProvider, "openai") || strings.Contains(activeProvider, "gpt")
	default:
		return false
	}
}
