package ai

import (
	"context"
	"fmt"
	"io"

	"github.com/knights-analytics/hugot"
)

// ONNXProvider implements the AIProvider interface for local ONNX models.
type ONNXProvider struct {
	pipeline     *hugot.SequenceClassificationPipeline
	modelPath    string
}

// NewONNXProvider creates a new provider for a local ONNX model.
func NewONNXProvider(modelPath string) (*ONNXProvider, error) {	
	pipeline, err := hugot.NewSequenceClassificationPipeline(modelPath, "onnx", "ort")
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX pipeline: %w", err)
	}

	return &ONNXProvider{
		pipeline:  pipeline,
		modelPath: modelPath,
	};
}

// Process is adapted for a classification model. It runs prediction on the prompt.
func (p *ONNXProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	inputs := []string{prompt}

	// Run prediction
	predictions, err := p.pipeline.Predict(inputs, nil)
	if err != nil {
		return fmt.Errorf("ONNX prediction failed: %w", err)
	}

	if len(predictions) == 0 || len(predictions[0]) == 0 {
		return fmt.Errorf("received no prediction from ONNX model")
	}

	topPrediction := predictions[0][0]
	result := fmt.Sprintf("Classification Result:\n- Label: %s\n- Score: %.4f", topPrediction.Label, topPrediction.Score)
	
	_, err = io.WriteString(w, result)
	return err
}

// GenerateContent runs classification and returns the result.
func (p *ONNXProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	inputs := []string{prompt}

	predictions, err := p.pipeline.Predict(inputs, nil)
	if err != nil {
		return "", fmt.Errorf("ONNX prediction failed: %w", err)
	}

	if len(predictions) == 0 || len(predictions[0]) == 0 {
		return "", fmt.Errorf("received no prediction from ONNX model")
	}

	topPrediction := predictions[0][0]
	result := fmt.Sprintf("Classification Result:\n- Label: %s\n- Score: %.4f", topPrediction.Label, topPrediction.Score)

	if streamCallback != nil {
		streamCallback(result)
	}

	return result, nil
}

// GenerateJSONData is not implemented for this classification model.
func (p *ONNXProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	return "", fmt.Errorf("JSON data generation is not supported by this ONNX provider")
}

// GetModelInfo returns information about the model.
func (p *ONNXProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "ONNX",
		ModelName:    p.modelPath,
	}
}
