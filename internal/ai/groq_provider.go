package ai

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/magicx-ai/groq-go/groq"
)

// GroqProvider implements the AIProvider interface for Groq.
type GroqProvider struct {
	client    groq.Client
	modelName string
	key       string // Store the key for identification/logging if needed
}

// NewGroqProvider creates a new Groq provider for a single API key.
func NewGroqProvider(apiKey string, modelName string) *GroqProvider {
	if apiKey == "" {
		log.Println("Groq API key is empty. Groq AI will be unavailable for this provider instance.")
		return nil
	}

	client := groq.NewClient(apiKey, &http.Client{Timeout: 60 * time.Second})

	return &GroqProvider{
		client:    client,
		modelName: modelName,
		key:       apiKey,
	}
}

// GenerateContent generates a non-streaming response from Groq.
func (p *GroqProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	req := groq.ChatCompletionRequest{
		Model: groq.ModelID(p.modelName), // Cast to groq.ModelID
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
		fullResponse := resp.Choices[0].Message.Content
		if streamCallback != nil {
			streamCallback(fullResponse)
		}
		return fullResponse, nil
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
		Model: groq.ModelID(p.modelName), // Cast to groq.ModelID
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

// GetModelInfo returns information about the model.
func (p *GroqProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Groq",
		ModelName:    p.modelName,
	}
}

// Process sends a request to the AI service and returns a stream of responses.
func (p *GroqProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	req := groq.ChatCompletionRequest{
		Model: groq.ModelID(p.modelName), // Cast to groq.ModelID
		Messages: []groq.Message{
			{
				Role:    "system",
				Content: system,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	respCh, _, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("groq content generation failed: %w", err)
	}

	for res := range respCh {
		if res.Error != nil {
			if res.Error == io.EOF {
				return nil
			}
			return fmt.Errorf("error occurred during stream: %v", res.Error)
		}
		if len(res.Response.Choices) > 0 {
			fmt.Fprint(w, res.Response.Choices[0].Delta.Content)
		}
	}
	return nil
}
