package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/apps/auth-service/internal/handlers"
	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var (
	logger *slog.Logger
	port   = ":8081"
)

func init() {
	logger = types.NewColoredLogger("auth-service")

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(logLevel),
	}

	handler := types.NewColoredJSONHandler(os.Stdout, opts)
	logger = slog.New(handler).With("service", "auth-service", "version", "1.0.0")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	types.LogInfo(logger, "Starting Auth Service", "port", port)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/api/v1/auth/register", handlers.Register)
	mux.HandleFunc("/api/v1/auth/login", handlers.Login)
	mux.HandleFunc("/api/v1/auth/refresh", handlers.RefreshToken)
	mux.HandleFunc("/api/v1/auth/validate", handlers.ValidateToken)

	server := &http.Server{
		Addr:           port,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		types.LogInfo(logger, "Server listening", "port", port, "addr", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			types.LogError(logger, err, "Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	types.LogInfo(logger, "Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		types.LogError(logger, err, "Server forced to shutdown")
	}

	types.LogInfo(logger, "Server exited")
}
