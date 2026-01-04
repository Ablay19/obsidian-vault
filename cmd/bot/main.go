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

	db := database.OpenDB()
	defer db.Close()

	database.RunMigrations(db)

	for {
		if err := database.CheckExistingInstance(db); err != nil {
			telemetry.ZapLogger.Sugar().Info("Another instance is running, retrying in 15 seconds...", "error", err)
			time.Sleep(15 * time.Second)
		} else {
			break
		}
	}

	if err := database.AddInstance(db); err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Error adding instance", "error", err)
	}

	heartbeatTicker := time.NewTicker(database.HEARTBEAT_THRESHOLD / 2)
	go func() {
		for range heartbeatTicker.C {
			if err := database.UpdateInstanceHeartbeat(db); err != nil {
				telemetry.ZapLogger.Sugar().Errorw("Error updating instance heartbeat", "error", err)
			}
		}
	}()

	runtimeConfigManager, err := state.NewRuntimeConfigManager(db)
	if err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Failed to initialize state manager", "error", err)
	}

	ctx := context.Background()
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
	dash := dashboard.NewDashboard(aiService, runtimeConfigManager, db, authService, wsManager)
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
		database.RemoveInstance(db)
		os.Exit(0)
	}()

	if err := bot.Run(db, aiService, runtimeConfigManager, wsManager); err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("Bot failed to run", "error", err)
	}
}

