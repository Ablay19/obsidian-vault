package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	ModelFlashSearch = "gemini-3-flash-preview"
	ModelProComplex  = "gemini-3-pro-preview"
	ModelImageGen    = "gemini-3-flash-preview"
)

type AIService struct {
	clients         []*genai.Client
	currentKeyIndex int
	mu              sync.Mutex
}

func NewAIService(ctx context.Context) *AIService {
	apiKeysStr := os.Getenv("GEMINI_API_KEYS")
	if apiKeysStr == "" {
		log.Println("GEMINI_API_KEYS environment variable not set. Gemini AI will be unavailable.")
		return nil
	}

	apiKeys := strings.Split(apiKeysStr, ",")
	var clients []*genai.Client

	for _, key := range apiKeys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey != "" {
			client, err := genai.NewClient(ctx, option.WithAPIKey(trimmedKey))
			if err != nil {
				log.Printf("Error creating Gemini client with a key: %v", err)
				continue
			}
			clients = append(clients, client)
		}
	}

	if len(clients) == 0 {
		log.Println("No valid Gemini clients could be created. Gemini AI will be unavailable.")
		return nil
	}

	return &AIService{clients: clients}
}

func (s *AIService) getClient() *genai.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clients[s.currentKeyIndex]
}

func (s *AIService) switchToNextKey() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentKeyIndex = (s.currentKeyIndex + 1) % len(s.clients)
	log.Printf("Switched to Gemini API Key #%d", s.currentKeyIndex+1)
}

// SwitchKey manually switches to the next key and returns its index.
func (s *AIService) SwitchKey() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentKeyIndex = (s.currentKeyIndex + 1) % len(s.clients)
	log.Printf("Manually switched to Gemini API Key #%d", s.currentKeyIndex+1)
	return s.currentKeyIndex
}

// GenerateContent is for streaming a human-readable response.
func (s *AIService) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	var fullResponse strings.Builder
	var err error

	for i := 0; i < len(s.clients); i++ {
		client := s.getClient()
		model := client.GenerativeModel(modelType)
		var parts []genai.Part
		parts = append(parts, genai.Text(prompt))
		if len(imageData) > 0 {
			parts = append(parts, genai.ImageData("jpeg", imageData))
		}

		iter := model.GenerateContentStream(ctx, parts...)
		fullResponse.Reset()

		// Streaming loop
		for {
			resp, streamErr := iter.Next()
			if streamErr == iterator.Done {
				err = nil
				break
			}
			if streamErr != nil {
				err = streamErr
				break
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

		if err == nil {
			return fullResponse.String(), nil // Success
		}

		// Check if the error is a quota error
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
			log.Printf("Gemini API Key #%d quota exceeded. Trying next key.", s.currentKeyIndex+1)
			s.switchToNextKey()
			continue // Retry with the next key
		}

		// For other errors, don't retry
		return "", err
	}

	return "", fmt.Errorf("all Gemini API keys failed: %w", err)
}

// GenerateJSONData is for getting structured data in JSON format. Non-streaming.
func (s *AIService) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	var err error

	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.
The content of "topics" and "questions" fields should be in %s.
Text to analyze:
%s`, language, text)

	for i := 0; i < len(s.clients); i++ {
		client := s.getClient()
		model := client.GenerativeModel(ModelProComplex)

		var resp *genai.GenerateContentResponse
		resp, err = model.GenerateContent(ctx, genai.Text(prompt))

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
			err = fmt.Errorf("no content generated from AI for JSON data") // Set error to retry
		}
		
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
			log.Printf("Gemini API Key #%d quota exceeded. Trying next key.", s.currentKeyIndex+1)
			s.switchToNextKey()
			continue
		}
		
		return "", err
	}

	return "", fmt.Errorf("all Gemini API keys failed for JSON data: %w", err)
}
