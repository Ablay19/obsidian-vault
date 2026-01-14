package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var logger = types.NewStructuredLogger("api-gateway.handlers")

func Health(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"service":   "api-gateway",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode health response", "error", err)
	}

	duration := time.Since(startTime)
	logger.Debug("Health check completed", "duration_ms", duration.Milliseconds(), "status", http.StatusOK)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	response := types.ErrorResponse{
		Error:   "validation_error",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

	logger.Warn("Error response", "status", statusCode, "message", message)
}
