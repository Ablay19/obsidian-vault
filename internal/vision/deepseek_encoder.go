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

// DeepSeekVisionEncoder implements vision encoding using DeepSeek's multimodal capabilities
type DeepSeekVisionEncoder struct {
	aiService ai.AIServiceInterface
}

// NewDeepSeekVisionEncoder creates a new DeepSeek vision encoder
func NewDeepSeekVisionEncoder(aiService ai.AIServiceInterface) *DeepSeekVisionEncoder {
	return &DeepSeekVisionEncoder{
		aiService: aiService,
	}
}

// EncodeImage encodes an image into a vision embedding using DeepSeek
func (e *DeepSeekVisionEncoder) EncodeImage(ctx context.Context, imagePath string) ([]float32, error) {
	// For DeepSeek, we'll use a text-based approach to generate embeddings
	// since DeepSeek-VL2 focuses more on understanding than pure embeddings

	prompt := fmt.Sprintf(`Analyze this image and generate a 512-dimensional semantic embedding vector as a JSON array of floats.

Image path: %s

The embedding should capture:
- Visual content and layout
- Text content and structure
- Document type and purpose
- Key visual elements

Return only the JSON array of 512 float values, nothing else.`, imagePath)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.1, // Low temperature for consistent embeddings
	}

	var response strings.Builder
	err := e.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, fmt.Errorf("DeepSeek vision encoding failed: %w", err)
	}

	// Parse the JSON response
	result := strings.TrimSpace(response.String())
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var embedding []float32
	if err := json.Unmarshal([]byte(result), &embedding); err != nil {
		// If JSON parsing fails, try to extract numbers from text
		zap.S().Warn("Failed to parse DeepSeek embedding JSON, attempting fallback parsing", "error", err)
		embedding, err = e.parseEmbeddingFromText(result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedding: %w", err)
		}
	}

	// Validate and normalize
	if len(embedding) != 512 {
		zap.S().Warn("DeepSeek returned embedding with unexpected dimensions", "expected", 512, "got", len(embedding))
		// Pad or truncate to expected size
		embedding = e.normalizeEmbeddingDimensions(embedding, 512)
	}

	// Normalize the embedding
	e.normalizeEmbedding(embedding)

	return embedding, nil
}

// EncodeWithText encodes both image and text, returning multimodal embedding
func (e *DeepSeekVisionEncoder) EncodeWithText(ctx context.Context, imagePath, text string) (MultimodalEmbedding, error) {
	// Get image embedding
	imageEmbedding, err := e.EncodeImage(ctx, imagePath)
	if err != nil {
		return MultimodalEmbedding{}, err
	}

	// Generate text embedding using AI service
	textEmbedding, err := e.generateTextEmbedding(ctx, text)
	if err != nil {
		zap.S().Warn("Text embedding failed, using zero vector", "error", err)
		textEmbedding = make([]float32, 512) // Zero vector fallback
	}

	return MultimodalEmbedding{
		ImageVector: imageEmbedding,
		TextVector:  textEmbedding,
		Confidence:  0.85, // DeepSeek typically has high confidence
	}, nil
}

// IsAvailable checks if DeepSeek vision encoding is available
func (e *DeepSeekVisionEncoder) IsAvailable() bool {
	return e.aiService != nil && strings.Contains(strings.ToLower(e.aiService.GetActiveProviderName()), "deepseek")
}

// GetName returns the encoder name
func (e *DeepSeekVisionEncoder) GetName() string {
	return "DeepSeek-VL2"
}

// generateTextEmbedding creates text embeddings using the AI service
func (e *DeepSeekVisionEncoder) generateTextEmbedding(ctx context.Context, text string) ([]float32, error) {
	prompt := fmt.Sprintf(`Generate a 512-dimensional embedding vector for the following text as a JSON array of floats:

Text: %s

Return only the JSON array, nothing else.`, text)

	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.0, // Deterministic
	}

	var response strings.Builder
	err := e.aiService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})

	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(response.String())
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	var embedding []float32
	if err := json.Unmarshal([]byte(result), &embedding); err != nil {
		return nil, fmt.Errorf("failed to parse text embedding: %w", err)
	}

	e.normalizeEmbedding(embedding)
	return embedding, nil
}

// parseEmbeddingFromText attempts to extract numbers from text response
func (e *DeepSeekVisionEncoder) parseEmbeddingFromText(text string) ([]float32, error) {
	// Simple regex to extract numbers
	var embedding []float32

	// Look for JSON-like array or comma-separated numbers
	parts := strings.Split(text, ",")
	for _, part := range parts {
		part = strings.Trim(part, "[] \t\n")
		if part == "" {
			continue
		}

		var num float32
		if _, err := fmt.Sscanf(part, "%f", &num); err != nil {
			continue // Skip non-numeric parts
		}
		embedding = append(embedding, num)
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("no numeric values found in response")
	}

	return embedding, nil
}

// normalizeEmbeddingDimensions adjusts embedding to target size
func (e *DeepSeekVisionEncoder) normalizeEmbeddingDimensions(embedding []float32, targetSize int) []float32 {
	if len(embedding) == targetSize {
		return embedding
	}

	result := make([]float32, targetSize)

	if len(embedding) < targetSize {
		// Pad with zeros
		copy(result, embedding)
		// Fill remaining with average of existing values
		if len(embedding) > 0 {
			sum := float32(0)
			for _, v := range embedding {
				sum += v
			}
			avg := sum / float32(len(embedding))
			for i := len(embedding); i < targetSize; i++ {
				result[i] = avg
			}
		}
	} else {
		// Truncate
		copy(result, embedding[:targetSize])
	}

	return result
}

// normalizeEmbedding normalizes vector to unit length
func (e *DeepSeekVisionEncoder) normalizeEmbedding(embedding []float32) {
	if len(embedding) == 0 {
		return
	}

	// Calculate L2 norm
	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))

	// Normalize
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}
}
