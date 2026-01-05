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
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/ssh"
)

// AppLogger wraps zap logger with color support
type AppLogger struct {
	logger       *zap.Logger
	enableColors bool
}

// NewAppLogger creates a new colored logger
func NewAppLogger(enableColors bool) *AppLogger {
	logger := zap.NewProduction()
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

func main() {
	logger := NewAppLogger(os.Getenv("ENABLE_COLORFUL_LOGS") == "true")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Starting Obsidian Bot API Server", zap.String("version", cfg.Version))

	// Initialize database
	db, err := database.NewConnection(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Database migrations completed")

	// Initialize router
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

	// Register dashboard routes
	dashboard.RegisterRoutes(router, db, logger)
	// Register SSH server routes
	ssh.RegisterRoutes(router, db, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server starting on port %s", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()

	setupGracefulShutdown(server, logger)
}
