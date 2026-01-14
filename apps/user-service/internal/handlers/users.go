package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var logger = types.NewStructuredLogger("user-service.handlers")

func Health(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"service":   "user-service",
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

func ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]interface{}{
		"status": "ok",
		"data": []map[string]interface{}{
			{
				"id":         "user_123",
				"email":      "user@example.com",
				"name":       "John Doe",
				"created_at": "2024-01-01T00:00:00Z",
			},
			{
				"id":         "user_456",
				"email":      "jane@example.com",
				"name":       "Jane Doe",
				"created_at": "2024-01-02T00:00:00Z",
			},
		},
		"meta": map[string]interface{}{
			"page":     1,
			"per_page": 20,
			"total":    2,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.URL.Path[len("/api/v1/users/"):]

	response := map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"id":         userID,
			"email":      "user@example.com",
			"name":       "John Doe",
			"avatar_url": "https://example.com/avatar.png",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"user_id":     "user_123",
			"bio":         "Software engineer",
			"preferences": map[string]interface{}{},
			"settings":    map[string]interface{}{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"message": "Profile updated successfully",
		"data": map[string]interface{}{
			"user_id":    "user_123",
			"updated":    req,
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
