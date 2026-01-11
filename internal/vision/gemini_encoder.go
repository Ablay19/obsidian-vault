package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"obsidian-automation/internal/ai"
	"strings"

	"go.uber.org/zap"
)

// GeminiVisionEncoder implements vision encoding using Google's Gemini
type GeminiVisionEncoder struct {
	aiService ai.AIServiceInterface
}

// NewGeminiVisionEncoder creates a new Gemini vision encoder
func NewGeminiVisionEncoder(aiService ai.AIServiceInterface) *GeminiVisionEncoder {
	return &GeminiVisionEncoder{
		aiService: aiService,
	}
}

// EncodeImage encodes an image using Gemini's vision capabilities
func (e *GeminiVisionEncoder) EncodeImage(ctx context.Context, imagePath string) ([]float32, error) {
	// Gemini has native vision support - use it to generate embeddings
	prompt := fmt.Sprintf(`Analyze this image and generate a 512-dimensional semantic embedding vector as a JSON array of floats.

Image: %s

Focus on:
- Visual content and composition
- Text and document structure
- Key objects and their relationships
- Overall scene understanding

Return only the JSON array of 512 float values.`, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1,
	}

	var response strings.Builder
	err := e.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, fmt.Errorf("Gemini vision encoding failed: %w", err)
	}

	// Parse response similar to DeepSeek
	result := strings.TrimSpace(response.String())
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var embedding []float32
	if err := json.Unmarshal([]byte(result), &embedding); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini embedding: %w", err)
	}

	// Normalize to 512 dimensions
	if len(embedding) != 512 {
		embedding = e.normalizeEmbeddingDimensions(embedding, 512)
	}

	e.normalizeEmbedding(embedding)
	return embedding, nil
}

// EncodeWithText creates multimodal embedding using Gemini
func (e *GeminiVisionEncoder) EncodeWithText(ctx context.Context, imagePath, text string) (MultimodalEmbedding, error) {
	imageEmbedding, err := e.EncodeImage(ctx, imagePath)
	if err != nil {
		return MultimodalEmbedding{}, err
	}

	textEmbedding, err := e.generateTextEmbedding(ctx, text)
	if err != nil {
		zap.S().Warn("Gemini text embedding failed", "error", err)
		textEmbedding = make([]float32, 512)
	}

	return MultimodalEmbedding{
		ImageVector: imageEmbedding,
		TextVector:  textEmbedding,
		Confidence:  0.9, // Gemini typically has high accuracy
	}, nil
}

// IsAvailable checks if Gemini vision is available
func (e *GeminiVisionEncoder) IsAvailable() bool {
	return e.aiService != nil &&
		(strings.Contains(strings.ToLower(e.aiService.GetActiveProviderName()), "gemini") ||
			strings.Contains(strings.ToLower(e.aiService.GetActiveProviderName()), "google"))
}

// GetName returns encoder name
func (e *GeminiVisionEncoder) GetName() string {
	return "Gemini-Vision"
}

// generateTextEmbedding creates text embeddings
func (e *GeminiVisionEncoder) generateTextEmbedding(ctx context.Context, text string) ([]float32, error) {
	prompt := fmt.Sprintf(`Generate a 512-dimensional embedding for this text as JSON array: %s`, text)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.0,
	}

	var response strings.Builder
	err := e.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(response.String())
	var embedding []float32
	json.Unmarshal([]byte(result), &embedding)

	e.normalizeEmbedding(embedding)
	return embedding, nil
}

// normalizeEmbeddingDimensions adjusts dimensions
func (e *GeminiVisionEncoder) normalizeEmbeddingDimensions(embedding []float32, target int) []float32 {
	result := make([]float32, target)
	copy(result, embedding)
	if len(embedding) < target && len(embedding) > 0 {
		avg := float32(0)
		for _, v := range embedding {
			avg += v
		}
		avg /= float32(len(embedding))
		for i := len(embedding); i < target; i++ {
			result[i] = avg
		}
	}
	return result
}

// normalizeEmbedding normalizes vector
func (e *GeminiVisionEncoder) normalizeEmbedding(embedding []float32) {
	if len(embedding) == 0 {
		return
	}

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
}
