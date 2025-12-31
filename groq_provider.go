package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/magicx-ai/groq-go/groq"
)

// GroqProvider implements the AIProvider interface for Groq.
type GroqProvider struct {
	client groq.Client
}

// NewGroqProvider creates a new Groq provider.
func NewGroqProvider(ctx context.Context) *GroqProvider {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Println("GROQ_API_KEY environment variable not set. Groq AI will be unavailable.")
		return nil
	}

	client := groq.NewClient(apiKey, &http.Client{})

	return &GroqProvider{client: client}
}

// GenerateContent generates a non-streaming response from Groq.
func (p *GroqProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	req := groq.ChatCompletionRequest{
		Model: groq.ModelIDLLAMA370B,
		Messages: []groq.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := p.client.CreateChatCompletion(req)
	if err != nil {
		return "", fmt.Errorf("groq content generation failed: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content generated from Groq")
}

// GenerateJSONData gets structured data in JSON format from Groq.
func (p *GroqProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.
The content of "topics" and "questions" fields should be in %s.
Text to analyze:
%s`, language, text)

	req := groq.ChatCompletionRequest{
		Model: groq.ModelIDLLAMA370B,
		Messages: []groq.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := p.client.CreateChatCompletion(req)
	if err != nil {
		return "", fmt.Errorf("groq json generation failed: %w", err)
	}

	if len(resp.Choices) > 0 {
		jsonStr := resp.Choices[0].Message.Content
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
		jsonStr = strings.TrimSpace(jsonStr)
		return jsonStr, nil
	}

	return "", fmt.Errorf("no content generated from Groq for JSON data")
}

// ProviderName returns the name of the provider.
func (p *GroqProvider) ProviderName() string {
	return "Groq"
}