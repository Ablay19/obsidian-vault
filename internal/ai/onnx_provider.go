package ai

import (
	"context"
	"fmt"
	"io"
	"log"
	
	ort "github.com/yalue/onnxruntime_go"
)

type ONNXProvider struct {
	session    *ort.DynamicAdvancedSession
	modelPath  string
	inputNames  []string
	outputNames []string
}

func NewONNXProvider(modelPath string) (*ONNXProvider, error) {
	// Initialize ONNX Runtime
	err := ort.InitializeEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX environment: %w", err)
	}

	// Create session options
	options, err := ort.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create session options: %w", err)
	}
	defer options.Destroy()

	// Set number of threads
	err = options.SetIntraOpNumThreads(4)
	if err != nil {
		return nil, fmt.Errorf("failed to set threads: %w", err)
	}

	// In a real scenario, you'd inspect the ONNX model to get exact names.
	// For now, assuming "input_ids" for input and "logits" for output for a DistilBERT-like model.
	inputNames := []string{"input_ids"} 
	outputNames := []string{"logits"}    

	// Create session
	session, err := ort.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX session: %w", err)
	}

	log.Printf("✅ ONNX model loaded: %s", modelPath)

	return &ONNXProvider{
		session:    session,
		modelPath:  modelPath,
		inputNames:  inputNames,
		outputNames: outputNames,
	}, nil
}

func (p *ONNXProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	result, err := p.GenerateContent(ctx, prompt, nil, "", nil)
	if err != nil {
		return err
	}
	
	_, err = w.Write([]byte(result))
	return err
}

func (p *ONNXProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	// This is a simplified example - you'll need to adapt based on your model
	// Most text classification models expect tokenized input
	
	// For demonstration, we'll use a simple approach
	// In production, you'd need proper tokenization and potentially attention masks
	
	inputValues, err := p.prepareInput(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to prepare input: %w", err)
	}
	defer func() {
		for _, v := range inputValues {
			v.Destroy()
		}
	}()

	// Pre-allocate output values
	outputValues := make([]ort.Value, len(p.outputNames))

	// Run inference
	err = p.session.Run(inputValues, outputValues) // p.session.Run expects input and output slices
	if err != nil {
		return "", fmt.Errorf("ONNX inference failed: %w", err)
	}
	defer func() {
		for _, output := range outputValues { // Iterate over outputValues, not outputs
			output.Destroy()
		}
	}()

	// Process output
	result, err := p.processOutput(outputValues) // Pass outputValues to processOutput
	if err != nil {
		return "", err
	}

	if streamCallback != nil {
		streamCallback(result)
	}

	return result, nil
}

func (p *ONNXProvider) prepareInput(text string) ([]ort.Value, error) {
	// For the distilbert model, it expects input_ids, and optionally attention_mask.
	// Tokenizer will be crucial here. For now, creating a dummy input_ids.
	
	// Example: assuming a fixed input size for now
	inputData := make([]int64, 128) // Dummy input_ids
	for i := range inputData {
		inputData[i] = int64(i) // Fill with some dummy data
	}
	
	inputShape := ort.NewShape(1, int64(len(inputData)))
	
	inputTensor, err := ort.NewTensor(inputShape, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input tensor: %w", err)
	}

	// ort.Tensor implements ort.Value directly.
	return []ort.Value{inputTensor}, nil
}

func (p *ONNXProvider) processOutput(outputs []ort.Value) (string, error) {
	if len(outputs) == 0 {
		return "", fmt.Errorf("no outputs from model")
	}

	// Get output tensor value
	outputValue := outputs[0]
	
	// Assuming the output is a tensor of float32 (e.g., logits)
	// We need to type assert ort.Value to *ort.Tensor to call GetData()
	outputTensor, ok := outputValue.(*ort.Tensor[float32])
	if !ok {
		return "", fmt.Errorf("output is not a float32 tensor")
	}

	outputData := outputTensor.GetData()
	
	// For classification, you might have logits or probabilities
	// Process them accordingly
	result := fmt.Sprintf("ONNX Model Output: %v", outputData)
	
	return result, nil
}

func (p *ONNXProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	// ONNX models typically don't generate JSON directly
	// You'd need to structure the output yourself
	return "", fmt.Errorf("JSON generation not supported by ONNX provider")
}

func (p *ONNXProvider) GetType() string {
	return "onnx"
}

func (p *ONNXProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "ONNX",
		ModelName:    p.modelPath,
	}
}

func (p *ONNXProvider) Destroy() {
	if p.session != nil {
		p.session.Destroy()
	}
}
