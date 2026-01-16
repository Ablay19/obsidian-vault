package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var goApps = []types.GoApplication{
	{
		ID:          "goapp-001",
		Name:        "api-gateway",
		Version:     "1.0.0",
		Description: "Main API gateway service for routing requests",
		ModulePath:  "github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway",
		EntryPoint:  "cmd/main.go",
		Port:        8080,
		Database:    types.DatabaseConfig{},
		APIs: []types.APIEndpoint{
			{
				Path:   "/health",
				Method: "GET",
			},
		},
		Dependencies: []types.GoDependency{},
		Resources: types.ResourceLimits{
			CPU:     "500m",
			Memory:  "512Mi",
			Storage: "1Gi",
		},
		Status:    "production",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	},
}

func ListGoApps(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"applications": goApps,
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   len(goApps),
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetKV(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	key := r.URL.Path[len("/api/v1/resources/kv/"):]
	logger.Info("Get KV request", "key", key)

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: Implement actual KV storage using Cloudflare KV API
	// For now, return a mock value
	value := ""
	switch key {
	case "whatsapp_session":
		value = "mock_session_data"
	default:
		writeError(w, http.StatusNotFound, "Key not found")
		return
	}

	response := map[string]interface{}{
		"key":    key,
		"value":  value,
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	duration := time.Since(startTime)
	logger.Info("Get KV completed", "key", key, "duration_ms", duration.Milliseconds())
}

func SetKV(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	key := r.URL.Path[len("/api/v1/resources/kv/"):]
	logger.Info("Set KV request", "key", key)

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	value, ok := data["value"].(string)
	if !ok {
		writeError(w, http.StatusBadRequest, "Value is required")
		return
	}

	// TODO: Implement actual KV storage using Cloudflare KV API
	// For now, just log it
	logger.Info("Setting KV value", "key", key, "value", value)

	response := map[string]interface{}{
		"key":    key,
		"value":  value,
		"status": "set",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	duration := time.Since(startTime)
	logger.Info("Set KV completed", "key", key, "duration_ms", duration.Milliseconds())
}

func GetWhatsAppQR(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	logger.Info("Get WhatsApp QR request")

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Proxy to whatsapp-go service
	whatsappURL := "http://whatsapp-go:3001/qr"
	resp, err := http.Get(whatsappURL)
	if err != nil {
		logger.Error("Failed to reach WhatsApp service", "error", err)
		writeError(w, http.StatusInternalServerError, "WhatsApp service unavailable")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		logger.Error("Failed to copy QR response", "error", copyErr)
	}

	duration := time.Since(startTime)
	logger.Info("Get WhatsApp QR completed", "duration_ms", duration.Milliseconds())
}

func GetWhatsAppStatus(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	logger.Info("Get WhatsApp status request")

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Proxy to whatsapp-go service
	whatsappURL := "http://whatsapp-go:3001/status"
	resp, err := http.Get(whatsappURL)
	if err != nil {
		logger.Error("Failed to reach WhatsApp service", "error", err)
		writeError(w, http.StatusInternalServerError, "WhatsApp service unavailable")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		logger.Error("Failed to copy status response", "error", copyErr)
	}

	duration := time.Since(startTime)
	logger.Info("Get WhatsApp status completed", "duration_ms", duration.Milliseconds())
}

func GoAppDetail(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Path[len("/api/v1/go-applications/"):]
	for _, app := range goApps {
		if app.ID == appID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(app)
			return
		}
	}
	writeError(w, http.StatusNotFound, "Application not found")
}

func ListPipelines(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"pipelines": []types.DeploymentPipeline{},
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   0,
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func PipelineDetail(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "Pipeline not found")
}

func ListSharedPackages(w http.ResponseWriter, r *http.Request) {
	packages := []types.SharedPackage{
		{
			ID:        "pkg-001",
			Name:      "shared-types",
			Version:   "1.0.0",
			Type:      "types",
			Languages: []string{"go", "typescript"},
			Exports:   []string{"types.go", "index.ts"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:        "pkg-002",
			Name:      "api-contracts",
			Version:   "1.0.0",
			Type:      "contracts",
			Languages: []string{"openapi"},
			Exports:   []string{"openapi.yaml"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:        "pkg-003",
			Name:      "communication",
			Version:   "1.0.0",
			Type:      "communication",
			Languages: []string{"go", "typescript"},
			Exports:   []string{"client.go", "client.ts"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
	response := map[string]interface{}{
		"packages": packages,
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   len(packages),
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
