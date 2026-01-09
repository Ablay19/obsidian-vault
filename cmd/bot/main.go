package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

func initTelemetry(logger *AppLogger) {
	if _, err := telemetry.Init("obsidian-bot"); err != nil {
		logger.Error("Failed to initialize telemetry", zap.Error(err))
		os.Exit(1)
	}
}

func initDatabase(logger *AppLogger) *database.DBClient {
	dbClient := database.OpenDB()
	logger.Info("Database connected successfully")
	return dbClient
}

func initRuntimeConfigManager(db *gorm.DB, logger *AppLogger) *state.RuntimeConfigManager {
	rcm, err := state.NewRuntimeConfigManager(db)
	if err != nil {
		logger.Error("Failed to initialize RuntimeConfigManager", zap.Error(err))
		os.Exit(1)
	}
	return rcm
}

func initServices(ctx context.Context, db *gorm.DB, rcm *state.RuntimeConfigManager, logger *AppLogger) (*ai.AIService, *auth.AuthService, *ws.Manager) {
	aiService := ai.NewAIService(ctx, rcm, config.AppConfig.ProviderProfiles, config.AppConfig.SwitchingRules)
	authService := auth.NewAuthService(config.AppConfig)
	wsManager := ws.NewManager()
	go wsManager.Start()
	database.RunMigrations(db)
	logger.Info("Database migrations completed")
	return aiService, authService, wsManager
}

func setupRouter(logger *AppLogger) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Add Google Cloud logging middleware (if enabled)
	if os.Getenv("ENABLE_GOOGLE_LOGGING") == "true" {
		router.Use(middleware.GoogleCloudLoggingMiddleware())
		logger.Info("Google Cloud logging enabled")
	}

	// Add request logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		logger.Info("API Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		)
	})

	return router
}

func initDashboard(router *gin.Engine, aiService *ai.AIService, rcm *state.RuntimeConfigManager, db *gorm.DB, authService *auth.AuthService, wsManager *ws.Manager, logger *AppLogger) {
	dashboardService := dashboard.NewDashboard(aiService, rcm, db, authService, wsManager)
	dashboardService.RegisterRoutes(router)
	ssh.RegisterRoutes(router, db, logger.logger)
}

func startServer(server *http.Server, logger *AppLogger) {
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %d", config.AppConfig.Dashboard.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()
}

func startBot(db *gorm.DB, aiService *ai.AIService, rcm *state.RuntimeConfigManager, wsManager *ws.Manager, logger *AppLogger) {
	go func() {
		logger.Info("Starting Telegram bot...")
		if err := bot.Run(db, aiService, rcm, wsManager); err != nil {
			logger.Error("Failed to start Telegram bot", zap.Error(err))
		}
	}()
}

func main() {
	logger := NewAppLogger(os.Getenv("ENABLE_COLORFUL_LOGS") == "true")

	initConfig()
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
	startBot(dbClient.DB, aiService, rcm, wsManager, logger)

	setupGracefulShutdown(server, logger)

	select {}
}
