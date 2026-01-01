package ai

import (
	"context"
	"fmt"
	"io"
	"log"
	"obsidian-automation/internal/config"
	"os"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GeminiProvider implements the ai.AIProvider interface for Google Gemini.
type GeminiProvider struct {
	clients         []*genai.Client
	currentKeyIndex int
	modelName       string
	mu              sync.Mutex
}

// NewGeminiProvider creates a new Gemini provider with API key rotation.
func NewGeminiProvider(ctx context.Context) *GeminiProvider {
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

	return &GeminiProvider{
		clients:   clients,
		modelName: config.AppConfig.Providers.Gemini.Model,
	}
}

func (p *GeminiProvider) getClient() *genai.Client {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.clients[p.currentKeyIndex]
}

func (p *GeminiProvider) switchToNextKey() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentKeyIndex = (p.currentKeyIndex + 1) % len(p.clients)
	log.Printf("Switched to Gemini API Key #%d", p.currentKeyIndex+1)
}

// SwitchKey manually switches to the next key and returns its index.
func (p *GeminiProvider) SwitchKey() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentKeyIndex = (p.currentKeyIndex + 1) % len(p.clients)
	log.Printf("Manually switched to Gemini API Key #%d", p.currentKeyIndex+1)
	return p.currentKeyIndex
}

// GenerateContent streams a human-readable response from Gemini.
func (p *GeminiProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	var fullResponse strings.Builder
	var err error

	for i := 0; i < len(p.clients); i++ {
		client := p.getClient()
		model := client.GenerativeModel(modelType)
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
			return fullResponse.String(), nil
		}

		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
			log.Printf("Gemini API Key #%d quota exceeded. Trying next key.", p.currentKeyIndex+1)
			p.switchToNextKey()
			continue
		}

		return "", err
	}

	return "", fmt.Errorf("all Gemini API keys failed: %w", err)
}

// GenerateJSONData gets structured data in JSON format from Gemini.
func (p *GeminiProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	var err error

	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.
The content of "topics" and "questions" fields should be in %s.
Text to analyze:
%s`, language, text)

	for i := 0; i < len(p.clients); i++ {
		client := p.getClient()
		model := client.GenerativeModel(p.modelName) // Use p.modelName here

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
			err = fmt.Errorf("no content generated from AI for JSON data")
		}

		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
			log.Printf("Gemini API Key #%d quota exceeded. Trying next key.", p.currentKeyIndex+1)
			p.switchToNextKey()
			continue
		}

		return "", err
	}

	return "", fmt.Errorf("all Gemini API keys failed for JSON data: %w", err)
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
	client := p.getClient()
	model := client.GenerativeModel(p.modelName)
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


