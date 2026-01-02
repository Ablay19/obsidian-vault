package ai

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiProvider implements the ai.AIProvider interface for Google Gemini.
type GeminiProvider struct {
	client    *genai.Client
	modelName string
	key       string // Store the key for identification/logging if needed, but not for direct use
	mu        sync.Mutex
}

// NewGeminiProvider creates a new Gemini provider for a single API key.
func NewGeminiProvider(ctx context.Context, apiKey string, modelName string) *GeminiProvider {
	if apiKey == "" {
		log.Println("Gemini API key is empty. Gemini AI will be unavailable for this provider instance.")
		return nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client with key %s: %v", apiKey, err)
		return nil
	}

	return &GeminiProvider{
		client:    client,
		modelName: modelName,
		key:       apiKey,
	}
}

// GenerateContent streams a human-readable response from Gemini.
func (p *GeminiProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	var fullResponse strings.Builder

	model := p.client.GenerativeModel(modelType)
	var parts []genai.Part
	parts = append(parts, genai.Text(prompt))
	if len(imageData) > 0 {
		parts = append(parts, genai.ImageData("jpeg", imageData))
	}

	iter := model.GenerateContentStream(ctx, parts...)
	fullResponse.Reset()

	for {
		resp, streamErr := iter.Next()
		if streamErr == iterator.Done {
			return fullResponse.String(), nil
		}
		if streamErr != nil {
			// Check for 429 specifically for higher-level handling (key rotation)
			if gerr, ok := streamErr.(*googleapi.Error); ok && gerr.Code == 429 {
				return "", fmt.Errorf("gemini_rate_limit_exceeded: %w", streamErr)
			}
			return "", streamErr
		}
		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
				chunk := string(txt)
				fullResponse.WriteString(chunk)
				if streamCallback != nil {
					streamCallback(chunk)
				}
			}
		}
	}
}

// GenerateJSONData gets structured data in JSON format from Gemini.
func (p *GeminiProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.
The content of "topics" and "questions" fields should be in %s.
Text to analyze:
%s`, language, text)

	model := p.client.GenerativeModel(p.modelName)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))

	if err == nil {
		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
				jsonStr := string(txt)
				jsonStr = strings.TrimPrefix(jsonStr, "```json")
				jsonStr = strings.TrimSuffix(jsonStr, "```")
				jsonStr = strings.TrimSpace(jsonStr)
				return jsonStr, nil
			}
		}
		err = fmt.Errorf("no content generated from AI for JSON data")
	}

	// Check for 429 specifically for higher-level handling (key rotation)
	if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
		return "", fmt.Errorf("gemini_rate_limit_exceeded: %w", err)
	}

	return "", err
}

// GetModelInfo returns information about the model.
func (p *GeminiProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Gemini",
		ModelName:    p.modelName,
	}
}

// Process sends a request to the AI service and returns a stream of responses.
func (p *GeminiProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	model := p.client.GenerativeModel(p.modelName)
	var parts []genai.Part
	if system != "" {
		parts = append(parts, genai.Text(system))
	}
	parts = append(parts, genai.Text(prompt))

	iter := model.GenerateContentStream(ctx, parts...)
	for {
		resp, streamErr := iter.Next()
		if streamErr == iterator.Done {
			return nil
		}
		if streamErr != nil {
			// Check for 429 specifically for higher-level handling (key rotation)
			if gerr, ok := streamErr.(*googleapi.Error); ok && gerr.Code == 429 {
				return fmt.Errorf("gemini_rate_limit_exceeded: %w", streamErr)
			}
			return streamErr
		}
		if len(resp.Candidates) > 0 {
			candidate := resp.Candidates[0]
			if candidate.Content != nil {
				for _, part := range candidate.Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						fmt.Fprint(w, txt)
					}
				}
			}
		}
	}
}
