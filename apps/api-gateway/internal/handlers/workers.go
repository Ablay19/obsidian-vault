package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway/internal/models"
	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

func ListWorkers(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	logger.Info("List workers request received")

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	limitInt := 20
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			limitInt = l
		}
	}

	offsetInt := 0
	if offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			offsetInt = o
		}
	}

	workers := models.GetAllWorkers()

	// Apply pagination
	total := len(workers)
	if offsetInt > total {
		offsetInt = 0
	}
	end := offsetInt + limitInt
	if end > total {
		end = total
	}
	paginatedWorkers := workers[offsetInt:end]

	response := types.APIResponse{
		Status:  "ok",
		Data:    paginatedWorkers,
		Message: "Workers retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	duration := time.Since(startTime)
	logger.Info("List workers completed", "count", len(paginatedWorkers), "duration_ms", duration.Milliseconds())
}

func WorkerDetail(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	workerID := r.URL.Path[len("/api/v1/workers/"):]
	logger.Info("Get worker detail request received", "worker_id", workerID)

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	worker, found := models.GetWorkerByID(workerID)
	if !found {
		writeError(w, http.StatusNotFound, "Worker not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(worker); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	duration := time.Since(startTime)
	logger.Info("Get worker detail completed", "worker_id", workerID, "duration_ms", duration.Milliseconds())
}

func CreateWorker(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	logger.Info("Create worker request received")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var worker types.WorkerModule
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := models.CreateWorker(worker); err != nil {
		logger.Error("Failed to create worker", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to create worker")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(worker); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	duration := time.Since(startTime)
	logger.Info("Create worker completed", "worker_id", worker.ID, "duration_ms", duration.Milliseconds())
}

func UpdateWorker(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	workerID := r.URL.Path[len("/api/v1/workers/"):]
	logger.Info("Update worker request received", "worker_id", workerID)

	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var worker types.WorkerModule
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updated, err := models.UpdateWorker(workerID, worker)
	if err != nil {
		logger.Error("Failed to update worker", "error", err, "worker_id", workerID)
		if err.Error() == "worker not found" {
			writeError(w, http.StatusNotFound, "Worker not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to update worker")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updated); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	duration := time.Since(startTime)
	logger.Info("Update worker completed", "worker_id", workerID, "duration_ms", duration.Milliseconds())
}

func DeleteWorker(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	workerID := r.URL.Path[len("/api/v1/workers/"):]
	logger.Info("Delete worker request received", "worker_id", workerID)

	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := models.DeleteWorker(workerID); err != nil {
		logger.Error("Failed to delete worker", "error", err, "worker_id", workerID)
		if err.Error() == "worker not found" {
			writeError(w, http.StatusNotFound, "Worker not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete worker")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

	duration := time.Since(startTime)
	logger.Info("Delete worker completed", "worker_id", workerID, "duration_ms", duration.Milliseconds())
}

func InvokeWorker(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	workerID := r.URL.Path[len("/api/v1/workers/"):]
	if idx := strings.LastIndex(workerID, "/invoke"); idx != -1 {
		workerID = workerID[:idx]
	}
	logger.Info("Invoke worker request", "worker_id", workerID)

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	worker, found := models.GetWorkerByID(workerID)
	if !found {
		writeError(w, http.StatusNotFound, "Worker not found")
		return
	}

	// Proxy to worker URL - for now using a placeholder URL
	// In real deployment, this should be configured per worker or discovered
	workerURL := "https://your-deployed-ai-worker.workers.dev"
	if worker.ID == "ai-worker-001" {
		workerURL = "http://localhost:8787" // Default Cloudflare Worker port for local testing
	}

	resp, err := http.Post(workerURL, "application/json", r.Body)
	if err != nil {
		logger.Error("Failed to invoke worker", "error", err, "worker_url", workerURL)
		writeError(w, http.StatusInternalServerError, "Failed to invoke worker")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		logger.Error("Failed to copy response body", "error", copyErr)
	}

	duration := time.Since(startTime)
	logger.Info("Invoke worker completed", "worker_id", workerID, "duration_ms", duration.Milliseconds())
}
