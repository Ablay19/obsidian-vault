package dashboard

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url" // New import
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/state" // Import the state package
	"obsidian-automation/internal/status"
	"os" // New import
	"time"
)

type ProcessedFile struct {
	ID            int64
	Hash          string
	Category      string
	Timestamp     time.Time
	ExtractedText sql.NullString
	Summary       sql.NullString
	Topics        sql.NullString
	Questions     sql.NullString
	AiProvider    sql.NullString
}

// Dashboard holds dependencies for the dashboard server.
type Dashboard struct {
	aiService *ai.AIService
	rcm       *state.RuntimeConfigManager
	db        *sql.DB
}

// NewDashboard creates a new Dashboard instance.
func NewDashboard(aiService *ai.AIService, rcm *state.RuntimeConfigManager, db *sql.DB) *Dashboard {
	return &Dashboard{
		aiService: aiService,
		rcm:       rcm,
		db:        db,
	}
}

// RegisterRoutes registers the dashboard's HTTP handlers on the provided router.
func (d *Dashboard) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/", d.handleDashboard)
	router.HandleFunc("/api/services/status", d.handleServicesStatus)

	// New routes for provider management
	router.HandleFunc("/api/ai/providers", d.handleGetAIProviders)             // GET all provider info
	router.HandleFunc("/api/ai/provider/config", d.handleGetProviderConfig)    // GET specific provider config
	router.HandleFunc("/api/ai/provider/set", d.handleSetAIProvider)           // POST set active provider (may be redundant with new config)
	router.HandleFunc("/api/ai/provider/toggle", d.handleToggleProviderStatus) // POST enable/disable/pause provider

	// New routes for API key management
	router.HandleFunc("/api/ai/keys", d.handleGetAPIKeys)               // GET all API keys for a provider
	router.HandleFunc("/api/ai/key/add", d.handleAddAPIKey)             // POST add new API key
	router.HandleFunc("/api/ai/key/remove", d.handleRemoveAPIKey)       // POST remove API key
	router.HandleFunc("/api/ai/key/toggle", d.handleToggleAPIKeyStatus) // POST enable/disable/block API key
	router.HandleFunc("/api/ai/key/rotate", d.handleRotateAPIKey)       // POST rotate API key

	// New routes for environment control
	router.HandleFunc("/api/env", d.handleGetEnvironmentState)     // GET current environment state
	router.HandleFunc("/api/env/set", d.handleSetEnvironmentState) // POST set environment state

	// Serve static files (CSS, JS)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/dashboard/static"))))

	// Panel rendering routes
	router.HandleFunc("/dashboard/panels/overview", d.handleOverviewPanel)
	router.HandleFunc("/dashboard/panels/file_processing", d.handleFileProcessingPanel)
	router.HandleFunc("/dashboard/panels/users", d.handleUsersPanel)
	router.HandleFunc("/dashboard/panels/db_config", d.handleDbConfigPanel)
	router.HandleFunc("/dashboard/panels/api_keys", d.handleAPIKeysPanel) // New route for API Keys panel
}

// handleAPIKeysPanel serves the APIKeysPanel HTML fragment.
func (d *Dashboard) handleAPIKeysPanel(w http.ResponseWriter, r *http.Request) {
	config := d.rcm.GetConfig()
	var apiKeysSlice []state.APIKeyState
	for _, key := range config.APIKeys {
		apiKeysSlice = append(apiKeysSlice, key)
	}
	APIKeysPanel(apiKeysSlice).Render(r.Context(), w)
}

// handleDashboard serves the main dashboard HTML page.
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	App().Render(r.Context(), w)
}

// handleOverviewPanel serves the OverviewPanel HTML fragment.
func (d *Dashboard) handleOverviewPanel(w http.ResponseWriter, r *http.Request) {
	services := status.GetServicesStatus(d.aiService, d.rcm)
	providers := d.getAIProviders()
	OverviewPanel(services, providers).Render(r.Context(), w)
}

// handleFileProcessingPanel serves the FileProcessingPanel HTML fragment.
func (d *Dashboard) handleFileProcessingPanel(w http.ResponseWriter, r *http.Request) {
	rows, err := d.db.Query("SELECT id, hash, category, timestamp, extracted_text, summary, topics, questions, ai_provider FROM processed_files ORDER BY timestamp DESC LIMIT 10")
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []ProcessedFile
	for rows.Next() {
		var file ProcessedFile
		if err := rows.Scan(&file.ID, &file.Hash, &file.Category, &file.Timestamp, &file.ExtractedText, &file.Summary, &file.Topics, &file.Questions, &file.AiProvider); err != nil {
			http.Error(w, "Failed to scan database rows", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	FileProcessingPanel(files).Render(r.Context(), w)
}

// handleUsersPanel serves the UsersPanel HTML fragment.
func (d *Dashboard) handleUsersPanel(w http.ResponseWriter, r *http.Request) {
	rows, err := d.db.Query("SELECT id, username, first_name, last_name, language_code, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.LanguageCode, &user.CreatedAt); err != nil {
			http.Error(w, "Failed to scan database rows", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	UsersPanel(users).Render(r.Context(), w)
}

// handleDbConfigPanel serves the DbConfigPanel HTML fragment.
func (d *Dashboard) handleDbConfigPanel(w http.ResponseWriter, r *http.Request) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	u, err := url.Parse(dbURL)
	if err != nil {
		http.Error(w, "Failed to parse database URL", http.StatusInternalServerError)
		return
	}

	dbType := u.Scheme
	dbHost := u.Host

	DbConfigPanel(dbType, dbHost).Render(r.Context(), w)
}

// handleServicesStatus provides the status of all monitored services.
func (d *Dashboard) handleServicesStatus(w http.ResponseWriter, r *http.Request) {
	statuses := status.GetServicesStatus(d.aiService, d.rcm)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		http.Error(w, "Failed to encode service statuses", http.StatusInternalServerError)
		return
	}
}

// handleGetAIProviders returns the available and active AI providers.
func (d *Dashboard) handleGetAIProviders(w http.ResponseWriter, r *http.Request) {
	providers := d.getAIProviders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (d *Dashboard) getAIProviders() struct {
	Available []string `json:"available"`
	Active    string   `json:"active"`
} {
	if d.aiService == nil {
		return struct {
			Available []string `json:"available"`
			Active    string   `json:"active"`
		}{
			Available: []string{},
			Active:    "",
		}
	}

	config := d.rcm.GetConfig() // Get current runtime config

	var availableProviders []string
	for name, ps := range config.Providers {
		if ps.Enabled { // Only consider enabled providers
			availableProviders = append(availableProviders, name)
		}
	}

	activeProviderName := d.aiService.GetActiveProviderName() // aiService still has the logic for active provider

	return struct {
		Available []string `json:"available"`
		Active    string   `json:"active"`
	}{
		Available: availableProviders,
		Active:    activeProviderName,
	}
}

// handleSetAIProvider sets the active AI provider.
func (d *Dashboard) handleSetAIProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	// The aiService is still needed to update the 'active' provider, if that concept remains.
	// However, the dashboard's primary interaction for configuration should be with RCM.
	// This handler's logic might need re-evaluation or removal based on the final UX.
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

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"AI provider set to %s"}`, req.Provider)
}

// handleGetProviderConfig returns the configuration for a specific provider, including all its keys.
func (d *Dashboard) handleGetProviderConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		http.Error(w, "Missing 'provider' query parameter", http.StatusBadRequest)
		return
	}

	config := d.rcm.GetConfig()
	providerState, providerExists := config.Providers[providerName]

	if !providerExists {
		http.Error(w, fmt.Sprintf("Provider '%s' not found", providerName), http.StatusNotFound)
		return
	}

	var apiKeys []state.APIKeyState
	for _, keyState := range config.APIKeys {
		if keyState.Provider == providerName {
			apiKeys = append(apiKeys, keyState)
		}
	}

	response := struct {
		Provider state.ProviderState `json:"provider"`
		APIKeys  []state.APIKeyState `json:"apiKeys"`
	}{
		Provider: providerState,
		APIKeys:  apiKeys,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode provider configuration", http.StatusInternalServerError)
		return
	}
}

// handleToggleProviderStatus handles POST requests to toggle the status of an AI provider.
func (d *Dashboard) handleToggleProviderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Provider      string `json:"provider"`
		Enabled       bool   `json:"enabled"`
		Paused        bool   `json:"paused"`
		Blocked       bool   `json:"blocked"`
		BlockedReason string `json:"blockedReason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := d.rcm.SetProviderState(req.Provider, req.Enabled, req.Paused, req.Blocked, req.BlockedReason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"Provider '%s' status updated."}`, req.Provider)
}

// handleGetAPIKeys returns all API keys, optionally filtered by provider.
func (d *Dashboard) handleGetAPIKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	providerName := r.URL.Query().Get("provider")
	config := d.rcm.GetConfig()

	var apiKeys []state.APIKeyState
	for _, keyState := range config.APIKeys {
		if providerName == "" || keyState.Provider == providerName {
			apiKeys = append(apiKeys, keyState)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apiKeys); err != nil {
		http.Error(w, "Failed to encode API keys", http.StatusInternalServerError)
		return
	}
}

// handleAddAPIKey handles POST requests to add a new API key.
func (d *Dashboard) handleAddAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProviderName string `json:"providerName"`
		KeyValue     string `json:"keyValue"`
		Enabled      bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	keyID, err := d.rcm.AddAPIKey(req.ProviderName, req.KeyValue, req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"API key added", "keyID":"%s"}`, keyID)
}

// handleRemoveAPIKey handles POST requests to remove an API key.
func (d *Dashboard) handleRemoveAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		KeyID string `json:"keyID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := d.rcm.RemoveAPIKey(req.KeyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"API key '%s' removed."}`, req.KeyID)
}

// handleToggleAPIKeyStatus handles POST requests to toggle the status of an API key.
func (d *Dashboard) handleToggleAPIKeyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		KeyID         string `json:"keyID"`
		Enabled       bool   `json:"enabled"`
		Blocked       bool   `json:"blocked"`
		BlockedReason string `json:"blockedReason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := d.rcm.SetAPIKeyStatus(req.KeyID, req.Enabled, req.Blocked, req.BlockedReason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"API key '%s' status updated."}`, req.KeyID)
}

// handleRotateAPIKey handles POST requests to rotate the active API key for a provider.
func (d *Dashboard) handleRotateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProviderName string `json:"providerName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := d.rcm.RotateAPIKey(req.ProviderName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"API key for provider '%s' rotated."}`, req.ProviderName)
}

// handleGetEnvironmentState handles GET requests to retrieve the current environment state.
func (d *Dashboard) handleGetEnvironmentState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	config := d.rcm.GetConfig()
	envState := config.Environment

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(envState); err != nil {
		http.Error(w, "Failed to encode environment state", http.StatusInternalServerError)
		return
	}
}

// handleSetEnvironmentState handles POST requests to update the bot's operational environment.
func (d *Dashboard) handleSetEnvironmentState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mode             string `json:"mode"`
		BackendHost      string `json:"backendHost"`
		IsolationEnabled bool   `json:"isolationEnabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := d.rcm.SetEnvironmentState(req.Mode, req.BackendHost, req.IsolationEnabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "message":"Environment state updated to mode '%s'."}`, req.Mode)
}
