package main

import (
	"context"
	"testing"
)

func TestAIService(t *testing.T) {
	// Create mock providers
	geminiProvider := &MockGeminiProvider{}
	groqProvider := &MockGroqProvider{}

	// Create a new AIService with the mock providers
	aiService := &AIService{
		providers: map[string]AIProvider{
			"Gemini": geminiProvider,
			"Groq":   groqProvider,
		},
		activeProvider: geminiProvider,
	}

	t.Run("Initial Provider", func(t *testing.T) {
		if aiService.GetActiveProviderName() != "Gemini" {
			t.Errorf("Expected initial provider to be Gemini, but got %s", aiService.GetActiveProviderName())
		}
	})

	t.Run("Set Provider", func(t *testing.T) {
		err := aiService.SetProvider("Groq")
		if err != nil {
			t.Errorf("Error setting provider: %v", err)
		}
		if aiService.GetActiveProviderName() != "Groq" {
			t.Errorf("Expected active provider to be Groq, but got %s", aiService.GetActiveProviderName())
		}

		err = aiService.SetProvider("InvalidProvider")
		if err == nil {
			t.Error("Expected error when setting invalid provider, but got nil")
		}
	})

	t.Run("GenerateContent", func(t *testing.T) {
		aiService.SetProvider("Gemini")
		resp, err := aiService.GenerateContent(context.Background(), "test", nil, "", nil)
		if err != nil {
			t.Errorf("Error generating content: %v", err)
		}
		if resp != "Gemini response" {
			t.Errorf("Expected Gemini response, but got %s", resp)
		}

		aiService.SetProvider("Groq")
		resp, err = aiService.GenerateContent(context.Background(), "test", nil, "", nil)
		if err != nil {
			t.Errorf("Error generating content: %v", err)
		}
		if resp != "Groq response" {
			t.Errorf("Expected Groq response, but got %s", resp)
		}
	})

	t.Run("GenerateJSONData", func(t *testing.T) {
		aiService.SetProvider("Gemini")
		resp, err := aiService.GenerateJSONData(context.Background(), "test", "English")
		if err != nil {
			t.Errorf("Error generating JSON data: %v", err)
		}
		if resp != `{"category": "gemini"}` {
			t.Errorf("Expected Gemini JSON response, but got %s", resp)
		}

		aiService.SetProvider("Groq")
		resp, err = aiService.GenerateJSONData(context.Background(), "test", "English")
		if err != nil {
			t.Errorf("Error generating JSON data: %v", err)
		}
		if resp != `{"category": "groq"}` {
			t.Errorf("Expected Groq JSON response, but got %s", resp)
		}
	})
}
