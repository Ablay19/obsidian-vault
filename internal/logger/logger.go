package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	appLogger *slog.Logger
	dbLogger  *slog.Logger
)

// Setup initializes the application and database loggers with distinct rotation policies.
func Setup() {
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		panic("failed to create logs directory: " + err.Error())
	}

	// Application Logger Configuration
	// Policy: 10MB max size, keep 5 backups, max age 30 days
	appRotator := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "app.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	appLogger = slog.New(slog.NewTextHandler(appRotator, nil))
	slog.SetDefault(appLogger) // Set as default for the application

	// Database Logger Configuration
	// Policy: 20MB max size, keep 10 backups, max age 60 days
	dbRotator := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "db.log"),
		MaxSize:    20, // megabytes
		MaxBackups: 10,
		MaxAge:     60, // days
		Compress:   true,
	}

	dbLogger = slog.New(slog.NewTextHandler(dbRotator, nil))
}

// GetDBLogger returns the specialized logger for database operations.
func GetDBLogger() *slog.Logger {
	if dbLogger == nil {
		// Fallback if Setup hasn't been called, though it should be.
		return slog.Default()
	}
	return dbLogger
}
