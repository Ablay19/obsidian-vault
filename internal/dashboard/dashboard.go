package dashboard

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/status" // Import the new status package
	"path"
)

//go:embed all:static
var staticFiles embed.FS

// Dashboard holds dependencies for the dashboard server.
type Dashboard struct {
	aiService *ai.AIService
	db        *sql.DB
}

// NewDashboard creates a new Dashboard instance.
func NewDashboard(aiService *ai.AIService, db *sql.DB) *Dashboard {
	return &Dashboard{
		aiService: aiService,
		db:        db,
	}
}

// RegisterRoutes registers the dashboard's HTTP handlers on the provided router.
func (d *Dashboard) RegisterRoutes(router *http.ServeMux) {
	fs := http.FileServer(http.FS(staticFiles))
	router.HandleFunc("/", d.handleDashboard)
	router.Handle("/static/", fs)
	router.HandleFunc("/api/services/status", d.handleServicesStatus)
	router.HandleFunc("/api/ai/providers", d.handleGetAIProviders)
	router.HandleFunc("/api/ai/provider/set", d.handleSetAIProvider)
}

// handleDashboard serves the main dashboard HTML page.
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.FileServer(http.FS(staticFiles)).ServeHTTP(w, r)
        return
    }
    content, err := staticFiles.ReadFile(path.Join("static", "index.html"))
    if err != nil {
        http.Error(w, "Could not load dashboard page.", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    w.Write(content)
}

// handleServicesStatus provides the status of all monitored services.
func (d *Dashboard) handleServicesStatus(w http.ResponseWriter, r *http.Request) {
	statuses := status.GetServicesStatus(d.aiService, d.db)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		http.Error(w, "Failed to encode service statuses", http.StatusInternalServerError)
		return
	}
}

// handleGetAIProviders returns the available and active AI providers.
func (d *Dashboard) handleGetAIProviders(w http.ResponseWriter, r *http.Request) {
	if d.aiService == nil {
		http.Error(w, "AI service not available", http.StatusInternalServerError)
		return
	}
	providers := struct {
		Available []string `json:"available"`
		Active    string   `json:"active"`
	}{
		Available: d.aiService.GetAvailableProviders(),
		Active:    d.aiService.GetActiveProviderName(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

// handleSetAIProvider sets the active AI provider.
func (d *Dashboard) handleSetAIProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	if d.aiService == nil {
		http.Error(w, "AI service not available", http.StatusInternalServerError)
		return
	}

	var req struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := d.aiService.SetProvider(req.Provider); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"success", "message":"AI provider set to %s"}`, req.Provider)
}
