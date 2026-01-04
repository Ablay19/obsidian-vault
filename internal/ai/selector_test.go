package ai

import (
	"obsidian-automation/internal/config"
	"testing"
)

func TestSelectProvider(t *testing.T) {
	profiles := map[string]config.ProviderConfig{
		"gemini": {
			ProviderName:         "gemini",
			InputCostPerToken:    0.0001,
			OutputCostPerToken:   0.0002,
			MaxInputTokens:       8000,
			MaxOutputTokens:      2048,
			LatencyMsThreshold:   2000,
			AccuracyPctThreshold: 0.95,
		},
		"groq": {
			ProviderName:         "groq",
			InputCostPerToken:    0.00005,
			OutputCostPerToken:   0.0001,
			MaxInputTokens:       4000,
			MaxOutputTokens:      1024,
			LatencyMsThreshold:   500,
			AccuracyPctThreshold: 0.90,
		},
	}

	rules := config.SwitchingRules{
		DefaultProvider:     "gemini",
		LatencyTarget:       1000,
		AccuracyThreshold:   0.92,
	}

	// Test case 1: Groq is cheaper and meets latency/accuracy
	selected := select_provider(3000, 1, 0.5, profiles, rules)
	if selected != "groq" {
		t.Errorf("Expected groq, got %s", selected)
	}

	// Test case 2: Groq is too slow, Gemini is selected
	rules.LatencyTarget = 400
	selected = select_provider(3000, 1, 0.5, profiles, rules)
	if selected != "gemini" {
		t.Errorf("Expected gemini, got %s", selected)
	}

	// Test case 3: Neither meets accuracy, default is selected
	rules.LatencyTarget = 1000
	rules.AccuracyThreshold = 0.98
	selected = select_provider(3000, 1, 0.5, profiles, rules)
	if selected != "gemini" {
		t.Errorf("Expected gemini (default), got %s", selected)
	}
}
