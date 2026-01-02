package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/state" // Import the new state package
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.LoadConfig() // Still load config for initial setup of things like dashboard port etc.

	db := database.OpenDB()
	defer db.Close()

	database.ApplySchemaAndMigrations(db)

	runtimeConfigManager, err := state.NewRuntimeConfigManager(db)
	if err != nil {
		log.Fatalf("Failed to initialize state manager: %v", err)
	}

	ctx := context.Background()
	// Pass runtimeConfigManager instead of appConfig
	aiService := ai.NewAIService(ctx, runtimeConfigManager)

	router := http.NewServeMux()
	bot.StartHealthServer(router) // This needs to be updated to use runtimeConfigManager later
	// Pass runtimeConfigManager instead of aiService and db
	dash := dashboard.NewDashboard(runtimeConfigManager)
	dash.RegisterRoutes(router)

	dashboardPort := config.AppConfig.Dashboard.Port // Dashboard port still comes from AppConfig
	go func() {
		addr := fmt.Sprintf(":%d", dashboardPort)
		log.Printf("Starting HTTP server on %s for health and dashboard...", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()

	// bot.Run also needs runtimeConfigManager now
	if err := bot.Run(db, aiService, runtimeConfigManager); err != nil {
		log.Fatal(err)
	}
}
