package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// OpenRouterProvider implements the AIProvider interface for OpenRouter.
type OpenRouterProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider.
func NewOpenRouterProvider(apiKey string, modelName string) *OpenRouterProvider {
	if apiKey == "" {
		slog.Info("OpenRouter API key is empty. OpenRouter AI will be unavailable.")
		return nil
	}
	return &OpenRouterProvider{
		apiKey:    apiKey,
		modelName: modelName,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterChatRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
}

type openRouterChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type openRouterStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Process sends a request to the OpenRouter service and returns a stream of responses.
func (p *OpenRouterProvider) Process(ctx context.Context, w io.Writer, system, prompt string, images []string) error {
	messages := []openRouterMessage{}
	if system != "" {
		messages = append(messages, openRouterMessage{Role: "system", Content: system})
	}
	messages = append(messages, openRouterMessage{Role: "user", Content: prompt})

	reqBody, err := json.Marshal(openRouterChatRequest{
		Model:    p.modelName,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/obsidian-automation-bot") // Optional, but recommended by OpenRouter
	req.Header.Set("X-Title", "Obsidian Automation Bot")                         // Optional

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openrouter api error (status %d): %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp openRouterStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			slog.Error("Error unmarshaling OpenRouter stream response", "error", err, "data", data)
			continue
		}

		if len(streamResp.Choices) > 0 {
			content := streamResp.Choices[0].Delta.Content
			fmt.Fprint(w, content)
		}
	}

	return nil
}

// GenerateContent streams a human-readable response from OpenRouter.
func (p *OpenRouterProvider) GenerateContent(ctx context.Context, prompt string, imageData []byte, modelType string, streamCallback func(string)) (string, error) {
	// OpenRouter/OpenAI compatible vision models use a different message format for images.
	// For now, let's implement basic text completion.

	reqBody, err := json.Marshal(openRouterChatRequest{
		Model: p.modelName,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter api error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp openRouterChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("openrouter api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) > 0 {
		content := chatResp.Choices[0].Message.Content
		if streamCallback != nil {
			streamCallback(content)
		}
		return content, nil
	}

	return "", fmt.Errorf("no content generated from OpenRouter")
}

// GenerateJSONData gets structured data in JSON format from OpenRouter.
func (p *OpenRouterProvider) GenerateJSONData(ctx context.Context, text, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following text and return ONLY a JSON object with the following fields:
- \"category\": a single category from the list [physics, math, chemistry, admin, general].
- \"topics\": a list of 3-5 key topics.
- \"questions\": a list of 2-3 review questions based on the text.
The content of \"topics\" and \"questions\" fields should be in %s.
Text to analyze:
%s`, language, text)

	reqBody, err := json.Marshal(openRouterChatRequest{
		Model: p.modelName,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter api error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp openRouterChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) > 0 {
		jsonStr := chatResp.Choices[0].Message.Content
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
		jsonStr = strings.TrimSpace(jsonStr)
		return jsonStr, nil
	}

	return "", fmt.Errorf("no content generated from OpenRouter for JSON data")
}

// GetModelInfo returns information about the model.
func (p *OpenRouterProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		ProviderName: "OpenRouter",
		ModelName:    p.modelName,
	}
}
