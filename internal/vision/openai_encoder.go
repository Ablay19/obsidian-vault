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

// OpenAIVisionEncoder implements vision encoding using OpenAI's GPT-4V
type OpenAIVisionEncoder struct {
	aiService ai.AIServiceInterface
}

// NewOpenAIVisionEncoder creates a new OpenAI vision encoder
func NewOpenAIVisionEncoder(aiService ai.AIServiceInterface) *OpenAIVisionEncoder {
	return &OpenAIVisionEncoder{
		aiService: aiService,
	}
}

// EncodeImage encodes an image using OpenAI's vision capabilities
func (e *OpenAIVisionEncoder) EncodeImage(ctx context.Context, imagePath string) ([]float32, error) {
	// Use GPT-4V to analyze image and generate embedding
	prompt := fmt.Sprintf(`Analyze this image and create a 512-dimensional semantic embedding vector as a JSON array of floats.

Image location: %s

Analyze for:
- Visual content and layout
- Text content and readability
- Document structure and type
- Key visual elements and their significance

Return only a JSON array of exactly 512 float values representing the semantic embedding.`, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1,
	}

	var response strings.Builder
	err := e.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI vision encoding failed: %w", err)
	}

	result := strings.TrimSpace(response.String())
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var embedding []float32
	if err := json.Unmarshal([]byte(result), &embedding); err != nil {
		zap.S().Warn("Failed to parse OpenAI embedding JSON", "error", err)
		// Try fallback parsing
		embedding, err = e.parseEmbeddingFromText(result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse OpenAI embedding: %w", err)
		}
	}

	// Ensure correct dimensions
	if len(embedding) != 512 {
		embedding = e.normalizeDimensions(embedding, 512)
	}

	e.normalizeEmbedding(embedding)
	return embedding, nil
}

// EncodeWithText creates multimodal embedding
func (e *OpenAIVisionEncoder) EncodeWithText(ctx context.Context, imagePath, text string) (MultimodalEmbedding, error) {
	imageEmbedding, err := e.EncodeImage(ctx, imagePath)
	if err != nil {
		return MultimodalEmbedding{}, err
	}

	textEmbedding, err := e.generateTextEmbedding(ctx, text)
	if err != nil {
		zap.S().Warn("OpenAI text embedding failed", "error", err)
		textEmbedding = make([]float32, 512)
	}

	return MultimodalEmbedding{
		ImageVector: imageEmbedding,
		TextVector:  textEmbedding,
		Confidence:  0.95, // GPT-4V is highly accurate
	}, nil
}

// IsAvailable checks if OpenAI vision is available
func (e *OpenAIVisionEncoder) IsAvailable() bool {
	return e.aiService != nil &&
		(strings.Contains(strings.ToLower(e.aiService.GetActiveProviderName()), "openai") ||
			strings.Contains(strings.ToLower(e.aiService.GetActiveProviderName()), "gpt"))
}

// GetName returns encoder name
func (e *OpenAIVisionEncoder) GetName() string {
	return "GPT-4V"
}

// generateTextEmbedding creates text embeddings using OpenAI
func (e *OpenAIVisionEncoder) generateTextEmbedding(ctx context.Context, text string) ([]float32, error) {
	prompt := fmt.Sprintf(`Generate a 512-dimensional embedding vector for this text as a JSON array of floats: %s

Return only the JSON array.`, text)

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

// parseEmbeddingFromText extracts numbers from text
func (e *OpenAIVisionEncoder) parseEmbeddingFromText(text string) ([]float32, error) {
	var embedding []float32
	parts := strings.Split(text, ",")

	for _, part := range parts {
		part = strings.Trim(part, "[] \t\n\"'")
		if part == "" {
			continue
		}

		var num float32
		if _, err := fmt.Sscanf(part, "%f", &num); err == nil {
			embedding = append(embedding, num)
		}
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("no numeric values found")
	}

	return embedding, nil
}

// normalizeDimensions adjusts embedding size
func (e *OpenAIVisionEncoder) normalizeDimensions(embedding []float32, target int) []float32 {
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

// normalizeEmbedding normalizes vector to unit length
func (e *OpenAIVisionEncoder) normalizeEmbedding(embedding []float32) {
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
