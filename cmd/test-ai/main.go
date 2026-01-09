package main

import (
	"context"
	"fmt"
	"log"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/telemetry"
)

func main() {
	// Initialize telemetry
	_, err := telemetry.Init("obsidian-bot-test")
	if err != nil {
		log.Fatal("Failed to initialize telemetry:", err)
	}

	// Test database connection
	dbClient := database.OpenDB()
	defer dbClient.DB.Close()

	// Test RuntimeConfigManager
	rcm, err := state.NewRuntimeConfigManager(dbClient.DB)
	if err != nil {
		log.Fatal("Failed to initialize RuntimeConfigManager:", err)
	}

	// Test AI Service initialization
	ctx := context.Background()
	_ = ai.NewAIService(ctx, rcm, config.AppConfig.ProviderProfiles, config.AppConfig.SwitchingRules)

	// Get configuration
	cfg := rcm.GetConfig(true)

	fmt.Printf("✅ Runtime Configuration:\n")
	fmt.Printf("  AI Enabled: %v\n", cfg.AIEnabled)
	fmt.Printf("  Active Provider: %s\n", cfg.ActiveProvider)
	fmt.Printf("  Providers: %d\n", len(cfg.Providers))
	fmt.Printf("  API Keys: %d\n", len(cfg.APIKeys))

	fmt.Printf("\n📋 Provider Details:\n")
	for name, provider := range cfg.Providers {
		fmt.Printf("  %s: Enabled=%v, Model=%s\n", name, provider.Enabled, provider.ModelName)
	}

	fmt.Printf("\n🔑 API Keys:\n")
	for id, key := range cfg.APIKeys {
		fmt.Printf("  %s: Provider=%s, Enabled=%v\n", id, key.Provider, key.Enabled)
	}

	fmt.Printf("\n🎯 Test completed!\n")
}
