package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/middleware"
	"obsidian-automation/internal/ssh"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/telemetry"
)

// AppLogger wraps zap logger with color support
type AppLogger struct {
	logger       *zap.Logger
	enableColors bool
}

// NewAppLogger creates a new colored logger
func NewAppLogger(enableColors bool) *AppLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create Zap logger: %v\n", err)
		os.Exit(1)
	}
	return &AppLogger{
		logger:       logger,
		enableColors: enableColors,
	}
}

func (l *AppLogger) Info(msg string, fields ...zap.Field) {
	if l.enableColors {
		l.logger.Info(msg, fields...)
	} else {
		l.logger.Info(msg, fields...)
	}
}

func (l *AppLogger) Error(msg string, fields ...zap.Field) {
	if l.enableColors {
		l.logger.Error(msg, fields...)
	} else {
		l.logger.Error(msg, fields...)
	}
}

func (l *AppLogger) Success(msg string, fields ...zap.Field) {
	if l.enableColors {
		l.logger.Info(msg, fields...)
	} else {
		l.logger.Info(msg, fields...)
	}
}

// setupGracefulShutdown handles graceful shutdown
func setupGracefulShutdown(srv *http.Server, logger *AppLogger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		logger.Info("🛑 Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", zap.Error(err))
		} else {
			logger.Info("🎉 Server stopped gracefully")
		}
	}()
}

func initConfig() {
	config.LoadConfig()
}

func validateEnvironment(logger *AppLogger) {
	logger.Info("Validating environment configuration...")

	// Required environment variables
	requiredVars := map[string]string{
		"TELEGRAM_BOT_TOKEN": "Telegram Bot API token from @BotFather",
		"ENVIRONMENT_MODE":  "Environment mode (dev/prod/staging)",
		"SESSION_SECRET":    "Session secret for authentication (min 32 chars)",
	}

	missingVars := []string{}
	for varName, description := range requiredVars {
		if value := os.Getenv(varName); value == "" {
			missingVars = append(missingVars, fmt.Sprintf("%s: %s", varName, description))
		}
	}

	if len(missingVars) > 0 {
		logger.Error("Missing required environment variables", zap.Strings("missing", missingVars))
		logger.Error("Please set these variables in your .env file or environment")
		os.Exit(1)
	}

	// Validate specific variables
	envMode := os.Getenv("ENVIRONMENT_MODE")
	if envMode != "" {
		validModes := []string{"dev", "prod", "staging"}
		found := false
		for _, mode := range validModes {
			if envMode == mode {
				found = true
				break
			}
		}
		if !found {
			logger.Error("Invalid ENVIRONMENT_MODE", zap.String("mode", envMode), zap.Strings("valid", validModes))
			os.Exit(1)
		}
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if len(sessionSecret) < 32 {
		logger.Error("SESSION_SECRET too short", zap.Int("length", len(sessionSecret)), zap.Int("minimum", 32))
		os.Exit(1)
	}

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if !strings.Contains(telegramToken, ":") {
		logger.Error("TELEGRAM_BOT_TOKEN appears invalid (should contain ':')")
		os.Exit(1)
	}

	// Check for AI providers (at least one should be configured)
	aiProviders := []string{
		"GEMINI_API_KEY",
		"GROQ_API_KEY",
		"HUGGINGFACE_API_KEY",
		"OPENROUTER_API_KEY",
	}

	aiConfigured := 0
	for _, provider := range aiProviders {
		if os.Getenv(provider) != "" {
			aiConfigured++
		}
	}

	if aiConfigured == 0 {
		logger.Error("No AI providers configured", zap.Strings("providers", aiProviders))
		logger.Error("Please configure at least one AI provider for the bot to function")
		os.Exit(1)
	}

	logger.Info("Environment validation completed", zap.Int("ai_providers", aiConfigured))
}

func main() {
	logger := NewAppLogger(os.Getenv("ENABLE_COLORFUL_LOGS") == "true")

	initConfig()
	validateEnvironment(logger)
	initTelemetry(logger)

	logger.Info("Starting Obsidian Bot API Server")

	db := initDatabase(logger)
	defer db.Close()

	rcm := initRuntimeConfigManager(db, logger)

	aiService, authService, wsManager := initServices(context.Background(), db, rcm, logger)

	router := setupRouter(logger)

	initDashboard(router, aiService, rcm, db, authService, wsManager, logger)

	port := config.AppConfig.Dashboard.Port
	if port == 0 {
		port = 8080
	}
	logger.Info(fmt.Sprintf("Using port: %d", port))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	startServer(server, logger)
	startBot(db, aiService, rcm, wsManager, logger)

	setupGracefulShutdown(server, logger)

	select {}
}
	logger.Info(fmt.Sprintf("Using port: %d", port))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	startServer(server, logger)
	startBot(db, aiService, rcm, wsManager, logger)

	setupGracefulShutdown(server, logger)

	select {}
}
	logger.Info(fmt.Sprintf("Using port: %d", port))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	startServer(server, logger)
	startBot(dbClient.DB, aiService, rcm, wsManager, logger)

	setupGracefulShutdown(server, logger)

	select {}
}
