package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	ModelFlashSearch = "gemini-3-flash-preview"
	ModelProComplex  = "gemini-3-pro-preview"
	ModelImageGen    = "gemini-3-flash-preview"
)

type AIService struct {
	client *genai.Client
}

func NewAIService(ctx context.Context) *AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("GEMINI_API_KEY environment variable not set. Gemini AI will be unavailable.")
		return nil
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		return nil
	}
	return &AIService{client: client}
}

// GenerateContent is for streaming a human-readable response.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	model := s.client.GenerativeModel(modelType)
	var parts []genai.Part
	parts = append(parts, genai.Text(prompt))

	if len(imageData) > 0 {
		parts = append(parts, genai.ImageData("jpeg", imageData))
	}

	iter := model.GenerateContentStream(ctx, parts...)
	var fullResponse strings.Builder
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", err
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
	return fullResponse.String(), nil
}

// GenerateJSONData is for getting structured data in JSON format. Non-streaming.
func (s *AIService) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	model := s.client.GenerativeModel(ModelProComplex)
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.

The content of "topics" and "questions" fields should be in %s.

Text to analyze:
%s`, language, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			jsonStr := string(txt)
			jsonStr = strings.TrimPrefix(jsonStr, "```json")
			jsonStr = strings.TrimSuffix(jsonStr, "```")
			jsonStr = strings.TrimSpace(jsonStr)
			return jsonStr, nil
		}
	}
	return "", fmt.Errorf("no content generated from AI for JSON data")
}
