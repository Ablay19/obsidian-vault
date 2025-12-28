package main

import (
	"context"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	// ModelFlashSearch represents the 'gemini-3-flash-preview' model.
	ModelFlashSearch = "gemini-3-flash-preview"
	// ModelProComplex represents the 'gemini-3-pro-preview' model.
	ModelProComplex = "gemini-3-pro-preview"
	// ModelImageGen represents the 'gemini-3-flash-preview' model.
	ModelImageGen = "gemini-3-flash-preview"
)

// AIService provides methods for interacting with the Gemini AI.
type AIService struct {
	client *genai.Client
}

// NewAIService creates a new AIService.
func NewAIService(ctx context.Context) *AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	return &AIService{
		client: client,
	}
}

// GenerateContent generates content using the specified model.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string) (string, error) {
	model := s.client.GenerativeModel(modelType)
	var parts []genai.Part
	parts = append(parts, genai.Text(prompt))

	if len(imageData) > 0 {
		parts = append(parts, genai.ImageData("jpeg", imageData))
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}

	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(txt), nil
	}
	return "", nil
}

// GenerateImage generates an image using the image generation model.
func (s *AIService) GenerateImage(ctx context.Context, prompt string) (string, error) {
	model := s.client.GenerativeModel(ModelImageGen)
	fullPrompt := "High quality scientific illustration, clean background, detailed, educational: " + prompt
	
	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(txt), nil
		}
	}
	
	return "", nil
}
