package ai

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hupe1980/go-huggingface"
)

// HuggingFaceProvider implements the AIProvider interface for Hugging Face models.
type HuggingFaceProvider struct {
	client *huggingface.InferenceClient
	model  string
}

// NewHuggingFaceProvider creates a new HuggingFaceProvider.
func NewHuggingFaceProvider(apiKey, model string) *HuggingFaceProvider {
	if apiKey == "" {
		return nil // Return nil if API key is empty
	}
	client := huggingface.NewInferenceClient(apiKey)
	return &HuggingFaceProvider{
		client: client,
		model:  model,
	}
}

// Process sends a request to the Hugging Face service and returns a stream of responses.
func (p *HuggingFaceProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	fullPrompt := prompt
	if system != "" {
		fullPrompt = fmt.Sprintf("%s\n\n%s", system, prompt)
	}

	res, err := p.client.TextGeneration(ctx, &huggingface.TextGenerationRequest{
		Model:  p.model,
		Inputs: fullPrompt,
	})
	if err != nil {
		return err
	}

	if len(res) > 0 {
		_, err := fmt.Fprint(w, res[0].GeneratedText)
		return err
	}

	return fmt.Errorf("no text generated")
}

// GenerateContent streams a human-readable response from Hugging Face.
func (p *HuggingFaceProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	res, err := p.client.TextGeneration(ctx, &huggingface.TextGenerationRequest{
		Model:  p.model,
		Inputs: prompt,
	})
	if err != nil {
		return "", err
	}

	if len(res) > 0 {
		text := res[0].GeneratedText
		if streamCallback != nil {
			streamCallback(text)
		}
		return text, nil
	}

	return "", fmt.Errorf("no text generated")
}

// GenerateJSONData gets structured data in JSON format from Hugging Face.
func (p *HuggingFaceProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- "category": a single category from the list [physics, math, chemistry, admin, general].
- "topics": a list of 3-5 key topics.
- "questions": a list of 2-3 review questions based on the text.
The content of "topics" and "questions" fields should be in %s.
Text to analyze:
%s`, language, text)

	res, err := p.client.TextGeneration(ctx, &huggingface.TextGenerationRequest{
		Model:  p.model,
		Inputs: prompt,
	})
	if err != nil {
		return "", err
	}

	if len(res) > 0 {
		jsonStr := res[0].GeneratedText
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
		jsonStr = strings.TrimSpace(jsonStr)
		return jsonStr, nil
	}

	return "", fmt.Errorf("no text generated from Hugging Face for JSON data")
}

// GetModelInfo returns information about the model.
func (p *HuggingFaceProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Hugging Face",
		ModelName:    p.model,
	}
}
