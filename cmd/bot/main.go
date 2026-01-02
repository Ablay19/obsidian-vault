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
	"time"
)

func main() {
	config.LoadConfig() // Still load config for initial setup of things like dashboard port etc.

	db := database.OpenDB()
	defer db.Close()

	database.RunMigrations(db)

	if err := database.CheckExistingInstance(db); err != nil {
		log.Fatalf("Error checking for existing instance: %v", err)
	}

	if err := database.AddInstance(db); err != nil {
		log.Fatalf("Error adding instance: %v", err)
	}

	// Start a goroutine to periodically update the instance heartbeat
	heartbeatTicker := time.NewTicker(database.HEARTBEAT_THRESHOLD / 2) // Update every half of the threshold
	go func() {
		for range heartbeatTicker.C {
			if err := database.UpdateInstanceHeartbeat(db); err != nil {
				log.Printf("Error updating instance heartbeat: %v", err)
			}
		}
	}()

	runtimeConfigManager, err := state.NewRuntimeConfigManager(db)
	if err != nil {
		log.Fatalf("Failed to initialize state manager: %v", err)
	}

	ctx := context.Background()
	// Pass runtimeConfigManager instead of appConfig
	aiService := ai.NewAIService(ctx, runtimeConfigManager)
	if aiService == nil {
		log.Println("AI Service failed to initialize. No AI providers available or configured. Proceeding without AI features.")
	}

	router := http.NewServeMux()
	bot.StartHealthServer(router) // This needs to be updated to use runtimeConfigManager later
	// Pass aiService and db
	dash := dashboard.NewDashboard(aiService, runtimeConfigManager, db)
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
		log.Println("Gracefully shutting down...")
		heartbeatTicker.Stop() // Stop the heartbeat ticker
		database.RemoveInstance(db)
		os.Exit(0)
	}()

	// bot.Run also needs runtimeConfigManager now
	if err := bot.Run(db, aiService, runtimeConfigManager); err != nil {
		log.Fatal(err)
	}
}
