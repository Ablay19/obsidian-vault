package logger

import (
	"log/slog"
	"os"
	"time"
)

// Setup initializes the structured logger.
func Setup() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// LogProviderSelection logs the provider selection details.
func LogProviderSelection(provider string, estimatedCost float64) {
	slog.Info("Provider selected",
		"provider", provider,
		"estimated_cost", estimatedCost,
	)
}

// LogPerformanceMetrics logs the performance metrics of an AI call.
func LogPerformanceMetrics(provider string, latency time.Duration, accuracy float64) {
	slog.Info("Performance metrics",
		"provider", provider,
		"latency_ms", latency.Milliseconds(),
		"accuracy", accuracy,
	)
}