package telemetry

import (
	"context"
	"log/slog"
	"os"
)

var logger *slog.Logger

func Init(serviceName string) {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func GetLogger() *slog.Logger {
	if logger == nil {
		Init("default")
	}
	return logger
}

func Info(msg string, args ...any)  { GetLogger().Info(msg, args...) }
func Warn(msg string, args ...any)  { GetLogger().Warn(msg, args...) }
func Error(msg string, args ...any) { GetLogger().Error(msg, args...) }
func Debug(msg string, args ...any) { GetLogger().Debug(msg, args...) }
func Fatal(msg string, args ...any) { GetLogger().Error(msg, args...); os.Exit(1) }

func WithContext(ctx context.Context, args ...any) *slog.Logger {
	return GetLogger().With(args...)
}
