package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/ai/local"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/pkg/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg := config.AppConfig
	if err := config.LoadConfig(); err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := utils.NewLogger(utils.LoggerConfig{
		Level:  cfg.LogLevel,
		File:   cfg.LogFile,
		Format: "text",
	})

	logger.Info("Starting Telegram AI Bot", 
		"version", "1.0.0",
		"environment", func() string {
			if cfg.IsDevelopment() { return "development" }
			return "production"
		}(),
	)

	// Create bot instance
	bot, err := bot.NewBot(cfg.TelegramBotToken, logger)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	// Initialize AI services
	modelManager := local.NewModelManager(logger)
	gpt2Provider, err := local.NewGPT2Provider("./models", logger)
	if err != nil {
		slog.Error("Failed to initialize GPT-2 provider", "error", err)
		os.Exit(1)
	}
	modelManager.AddModel("gpt2", gpt2Provider)

	// Initialize context manager
	contextManager := ai.NewContextManager(10, logger)

	// Initialize message handler
	messageHandler := bot.NewMessageHandler(bot, gpt2Provider, contextManager, logger)

	// Register handlers
	bot.RegisterHandler("message", messageHandler)
	
	// Initialize simple command handler
	commandHandler := bot.NewCommandHandler(bot, logger)
	bot.RegisterHandler("command", commandHandler)

	// Get bot info
	botUsername, err := bot.GetBotInfo()
	if err != nil {
		slog.Error("Failed to get bot info", "error", err)
	} else {
		slog.Info("Bot initialized", "username", botUsername)
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in goroutine
	go func() {
		slog.Info("Bot starting...")
		if err := bot.Start(); err != nil {
			slog.Error("Bot failed to start", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	slog.Info("Shutting down bot...")

	// Graceful shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	// Wait for graceful shutdown or timeout
	select {
	case <-sigChan:
		slog.Info("Received shutdown signal")
	case <-time.After(5 * time.Second):
		slog.Info("Shutdown timeout reached")
	case <-shutdownCtx.Done():
		slog.Info("Graceful shutdown completed")
	}

	// Stop bot
	bot.Stop()
	slog.Info("Bot stopped successfully")
}
