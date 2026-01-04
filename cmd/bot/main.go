package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/logger"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger.Setup()

	// Perform external binary check at startup
	if err := util.CheckExternalBinaries(); err != nil {
		slog.Error("Startup check failed", "error", err)
		os.Exit(1)
	}

	config.LoadConfig() // Still load config for initial setup of things like dashboard port etc.

	db := database.OpenDB()
	defer db.Close()

	database.RunMigrations(db)

	for {
		if err := database.CheckExistingInstance(db); err != nil {
			slog.Info("Another instance is running, retrying in 15 seconds...", "error", err)
			time.Sleep(15 * time.Second)
		} else {
			break // No other instance found, proceed with startup
		}
	}

	if err := database.AddInstance(db); err != nil {
		slog.Error("Error adding instance", "error", err)
		os.Exit(1)
	}

	// Start a goroutine to periodically update the instance heartbeat
	heartbeatTicker := time.NewTicker(database.HEARTBEAT_THRESHOLD / 2) // Update every half of the threshold
	go func() {
		for range heartbeatTicker.C {
			if err := database.UpdateInstanceHeartbeat(db); err != nil {
				slog.Error("Error updating instance heartbeat", "error", err)
			}
		}
	}()

	runtimeConfigManager, err := state.NewRuntimeConfigManager(db)
	if err != nil {
		slog.Error("Failed to initialize state manager", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	// Pass provider profiles and switching rules to NewAIService
	aiService := ai.NewAIService(ctx, runtimeConfigManager, config.AppConfig.ProviderProfiles, config.AppConfig.SwitchingRules)
	if aiService == nil {
		slog.Info("AI Service failed to initialize. No AI providers available or configured. Proceeding without AI features.")
	}

	authService := auth.NewAuthService(config.AppConfig)

	// WebSocket Manager
	wsManager := ws.NewManager()
	go wsManager.Start()

	router := http.NewServeMux()
	// Pass aiService to StartHealthServer
	bot.StartHealthServer(router)
	// Pass aiService, db, and wsManager
	dash := dashboard.NewDashboard(aiService, runtimeConfigManager, db, authService, wsManager)
	dash.RegisterRoutes(router)

	// Wrap with Middleware
	protectedRouter := authService.Middleware(router)

	dashboardPort := config.AppConfig.Dashboard.Port // Dashboard port still comes from AppConfig
	go func() {
		addr := fmt.Sprintf(":%d", dashboardPort)
		slog.Info("Starting HTTP server for health and dashboard...", "addr", addr)
		if err := http.ListenAndServe(addr, protectedRouter); err != nil {
			slog.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		slog.Info("Gracefully shutting down...")
		heartbeatTicker.Stop() // Stop the heartbeat ticker
		database.RemoveInstance(db)
		os.Exit(0)
	}()

	// bot.Run also needs runtimeConfigManager now
	if err := bot.Run(db, aiService, runtimeConfigManager, wsManager); err != nil {
		slog.Error("Bot failed to run", "error", err)
		os.Exit(1)
	}
}