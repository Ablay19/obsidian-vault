package utils

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerConfig holds configuration for logging
type LoggerConfig struct {
	Level  string
	File   string
	Format string
}

// Logger provides structured logging capabilities
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config LoggerConfig) *Logger {
	// Create log directory if it doesn't exist
	if config.File != "" {
		logDir := filepath.Dir(config.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.Error("Failed to create log directory", "error", err)
		}
	}

	// Create multi-writer for both file and console
	var writers []io.Writer

	// Console writer
	writers = append(writers, os.Stdout)

	// File writer with rotation
	if config.File != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   config.File,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)

	// Parse log level
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create handler
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(multiWriter, opts)
	} else {
		handler = slog.NewTextHandler(multiWriter, opts)
	}

	// Create logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &Logger{Logger: logger}
}

// With creates a new logger with additional attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithGroup creates a new logger with a group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{Logger: l.Logger.WithGroup(name)}
}

// Request logs a request with context
func (l *Logger) Request(method, path string, userID int64) {
	l.Info("Request received",
		"method", method,
		"path", path,
		"user_id", userID,
	)
}

// Response logs a response with context
func (l *Logger) Response(statusCode int, duration int64, userID int64) {
	l.Info("Response sent",
		"status_code", statusCode,
		"duration_ms", duration,
		"user_id", userID,
	)
}

// Error logs an error with context
func (l *Logger) ErrorWithContext(err error, message string, context map[string]any) {
	args := []any{"error", err}
	for k, v := range context {
		args = append(args, k, v)
	}
	l.Logger.Error(message, args...)
}

// AIRequest logs an AI request
func (l *Logger) AIRequest(provider, model string, inputTokens int, userID int64) {
	l.Info("AI request",
		"provider", provider,
		"model", model,
		"input_tokens", inputTokens,
		"user_id", userID,
	)
}

// AIResponse logs an AI response
func (l *Logger) AIResponse(provider, model string, outputTokens int, duration int64, userID int64) {
	l.Info("AI response",
		"provider", provider,
		"model", model,
		"output_tokens", outputTokens,
		"duration_ms", duration,
		"user_id", userID,
	)
}

// RateLimit logs a rate limit event
func (l *Logger) RateLimit(userID int64, limitType string, resetTime int64) {
	l.Warn("Rate limit exceeded",
		"user_id", userID,
		"limit_type", limitType,
		"reset_time", resetTime,
	)
}

// DatabaseOperation logs a database operation
func (l *Logger) DatabaseOperation(operation, table string, duration int64) {
	l.Debug("Database operation",
		"operation", operation,
		"table", table,
		"duration_ms", duration,
	)
}

// CacheOperation logs a cache operation
func (l *Logger) CacheOperation(operation, key string, hit bool) {
	l.Debug("Cache operation",
		"operation", operation,
		"key", key,
		"hit", hit,
	)
}
