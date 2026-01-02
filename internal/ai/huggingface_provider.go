package ai

import (
	"context"
	"fmt"
	"io"

	"github.com/hupe1980/go-huggingface"
)

// HuggingFaceProvider implements the AIProvider interface for Hugging Face models.
type HuggingFaceProvider struct {
	client *huggingface.InferenceClient
	model  string
}

// NewHuggingFaceProvider creates a new HuggingFaceProvider.
func NewHuggingFaceProvider(apiKey, model string) *HuggingFaceProvider {
	client := huggingface.NewInferenceClient(apiKey)
	return &HuggingFaceProvider{
		client: client,
		model:  model,
	}
}

// Process sends a request to the Hugging Face service and returns a stream of responses.
func (p *HuggingFaceProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	// Not implemented for now
	return fmt.Errorf("process is not implemented for huggingface provider")
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
	// Not implemented for now
	return "", fmt.Errorf("generate a JSON data is not implemented for huggingface provider")
}

// GetModelInfo returns information about the model.
func (p *HuggingFaceProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "Hugging Face",
		ModelName:    p.model,
	}
}
