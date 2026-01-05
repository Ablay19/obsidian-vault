package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"obsidian-automation/internal/ai"
)

// TestCloudflareIntegration tests the Cloudflare AI provider
func TestCloudflareIntegration() {
	workerURL := os.Getenv("CLOUDFLARE_WORKER_URL")
	if workerURL == "" {
		log.Fatal("CLOUDFLARE_WORKER_URL environment variable not set")
	}

	// Create Cloudflare provider
	provider := ai.NewCloudflareProvider(workerURL)

	fmt.Printf("🤖 Testing Cloudflare AI integration...\n")
	fmt.Printf("📍 Worker URL: %s\n", workerURL)

	// Test 1: Health check
	fmt.Println("\n🏥 Health Check:")
	ctx := context.Background()

	if err := provider.CheckHealth(ctx); err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
	} else {
		fmt.Println("✅ Health check passed")
	}

	// Test 2: Basic AI request
	fmt.Println("\n📝 AI Request Test:")
	req := &ai.RequestModel{
		SystemPrompt: "You are a helpful assistant.",
		UserPrompt:   "What is Cloudflare Workers?",
	}

	resp, err := provider.GenerateCompletion(ctx, req)
	if err != nil {
		fmt.Printf("❌ AI request failed: %v\n", err)
		return
	}

	fmt.Printf("✅ AI Response: %s\n", resp.Content)
	fmt.Printf("📊 Tokens Used: %d\n", resp.Usage.TotalTokens)
	fmt.Printf("🏷️ Provider: %s\n", resp.ProviderInfo.ProviderName)

	// Test 3: Model info
	fmt.Println("\nℹ️ Model Info:")
	modelInfo := provider.GetModelInfo()
	fmt.Printf("Model: %s\n", modelInfo.ModelName)
	fmt.Printf("Enabled: %t\n", modelInfo.Enabled)
}

func main() {
	TestCloudflareIntegration()
}
