package main

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/telemetry"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var version = "dev" // This will be overwritten by the build process

const HEARTBEAT_THRESHOLD = 30 * time.Second // Define locally after database refactor

func main() {
	// Initialize OpenTelemetry and the Zap logger
	tp, err := telemetry.Init("obsidian-automation-bot")
	if err != nil {
		panic(fmt.Sprintf("failed to initialize telemetry: %v", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			telemetry.ZapLogger.Sugar().Errorf("Error shutting down tracer provider: %v", err)
		}
	}()

	telemetry.ZapLogger.Sugar().Info("Starting bot", "version", version)

	config.LoadConfig()

	dbClient := database.OpenDB() // Renamed for clarity
	defer dbClient.DB.Close()

	database.RunMigrations(dbClient.DB) // Pass the raw *sql.DB for migrations

	ctx := context.Background()

	for {
		if err := database.CheckExistingInstance(ctx); err != nil {
			telemetry.ZapLogger.Sugar().Info("Another instance is running, retrying in 15 seconds...", "error", err)
			time.Sleep(15 * time.Second)
		} else {
			break
		}
	}

	if err := database.AddInstance(ctx, os.Getpid()); err != nil { // Pass PID
		telemetry.ZapLogger.Sugar().Fatalw("Error adding instance", "error", err)
	}

	heartbeatTicker := time.NewTicker(HEARTBEAT_THRESHOLD / 2)
	go func() {
		for range heartbeatTicker.C {
			if err := database.UpdateInstanceHeartbeat(ctx); err != nil { // Pass context
				telemetry.ZapLogger.Sugar().Errorw("Error updating instance heartbeat", "error", err)
			}
		}
	}()

	runtimeConfigManager, err := state.NewRuntimeConfigManager(dbClient.DB) // Pass raw *sql.DB
	if err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Failed to initialize state manager", "error", err)
	}

	aiService := ai.NewAIService(ctx, runtimeConfigManager, config.AppConfig.ProviderProfiles, config.AppConfig.SwitchingRules)
	if aiService == nil {
		telemetry.ZapLogger.Sugar().Info("AI Service failed to initialize. No AI providers available or configured. Proceeding without AI features.")
	}

	authService := auth.NewAuthService(config.AppConfig)
	wsManager := ws.NewManager()
	go wsManager.Start()

	router := gin.Default()
	router.Use(otelgin.Middleware("obsidian-automation"))

	router.POST("/api/v1/whatsapp/webhook", gin.WrapF(bot.WhatsAppWebhookHandler))
	bot.StartHealthServer(router)
	dash := dashboard.NewDashboard(aiService, runtimeConfigManager, dbClient.DB, authService, wsManager) // Pass raw *sql.DB
	dash.RegisterRoutes(router)

	dashboardPort := config.AppConfig.Dashboard.Port
	go func() {
		addr := fmt.Sprintf(":%d", dashboardPort)
		telemetry.ZapLogger.Sugar().Info("Starting HTTP server for health and dashboard...", "addr", addr)
		if err := router.Run(addr); err != nil {
			telemetry.ZapLogger.Sugar().Fatalw("HTTP server failed to start", "error", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		telemetry.ZapLogger.Sugar().Info("Gracefully shutting down...")
		heartbeatTicker.Stop()
		database.RemoveInstance(ctx) // Pass context
		os.Exit(0)
	}()

	if err := bot.Run(dbClient.DB, aiService, runtimeConfigManager, wsManager); err != nil { // Pass raw *sql.DB
		telemetry.ZapLogger.Sugar().Fatalw("Bot failed to run", "error", err)
	}
}
