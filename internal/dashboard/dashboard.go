package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url" // New import
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/gcp"
	"obsidian-automation/internal/security" // New import
	"obsidian-automation/internal/state" 
	"obsidian-automation/internal/status"
	"os"
	"strconv"
	"strings"
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
	aiService   *ai.AIService
	rcm         *state.RuntimeConfigManager
	db          *sql.DB
	authService *auth.AuthService
	wsManager   *ws.Manager
	rateLimiter *security.RateLimiter
}

// NewDashboard creates a new Dashboard instance.
func NewDashboard(aiService *ai.AIService, rcm *state.RuntimeConfigManager, db *sql.DB, authService *auth.AuthService, wsManager *ws.Manager) *Dashboard {
	return &Dashboard{
		aiService:   aiService,
		rcm:         rcm,
		db:          db,
		authService: authService,
		wsManager:   wsManager,
		rateLimiter: security.NewRateLimiter(100, time.Minute), // 100 req/min default
	}
}

// RegisterRoutes registers the dashboard's HTTP handlers on the provided router.
func (d *Dashboard) RegisterRoutes(router *http.ServeMux) {
	// Auth routes (unprotected)
	router.HandleFunc("/auth/login", d.handleLoginPage)
	router.HandleFunc("/auth/google/login", d.handleGoogleLogin)
	router.HandleFunc("/auth/dev/login", d.handleDevLogin)
	router.HandleFunc("/auth/google/callback", d.handleGoogleCallback)
	router.HandleFunc("/auth/logout", d.handleLogout)

	// GCP Discovery routes
	router.Handle("/api/auth/google/list-projects", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGCPListProjects)))
	router.Handle("/api/auth/google/list-keys", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGCPListKeys)))

	// Telegram Auth Webhook
	router.Handle("/api/v1/auth/telegram/webhook", d.rateLimiter.Middleware(http.HandlerFunc(d.handleTelegramWebhook)))

	// Protected routes
	router.HandleFunc("/", d.handleDashboardRedirect)
	router.HandleFunc("/dashboard/overview", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "overview") })
	router.HandleFunc("/dashboard/status", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "status") })
	router.HandleFunc("/dashboard/providers", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "providers") })
	router.HandleFunc("/dashboard/keys", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "keys") })
	router.HandleFunc("/dashboard/history", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "history") })
	router.HandleFunc("/dashboard/chat", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "chat") })
	router.HandleFunc("/dashboard/whatsapp", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "whatsapp") })
	router.HandleFunc("/dashboard/env", func(w http.ResponseWriter, r *http.Request) { d.renderDashboardPage(w, r, "env") })

	router.HandleFunc("/ws", d.wsManager.HandleWebSocket)
	router.Handle("/api/services/status", d.rateLimiter.Middleware(http.HandlerFunc(d.handleServicesStatus)))

	// New routes for provider management
	router.Handle("/api/ai/providers", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGetAIProviders)))             // GET all provider info
	router.Handle("/api/ai/provider/config", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGetProviderConfig)))    // GET specific provider config
	router.Handle("/api/ai/provider/set", d.rateLimiter.Middleware(http.HandlerFunc(d.handleSetAIProvider)))           // POST set active provider (may be redundant with new config)
	router.Handle("/api/ai/provider/toggle", d.rateLimiter.Middleware(http.HandlerFunc(d.handleToggleProviderStatus))) // POST enable/disable/pause provider

	// New routes for API key management
	router.Handle("/api/ai/keys", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGetAPIKeys)))               // GET all API keys for a provider
	router.Handle("/api/ai/key/add", d.rateLimiter.Middleware(http.HandlerFunc(d.handleAddAPIKey)))             // POST add new API key
	router.Handle("/api/ai/key/remove", d.rateLimiter.Middleware(http.HandlerFunc(d.handleRemoveAPIKey)))       // POST remove API key
	router.Handle("/api/ai/key/toggle", d.rateLimiter.Middleware(http.HandlerFunc(d.handleToggleAPIKeyStatus))) // POST enable/disable/block API key
	router.Handle("/api/ai/key/rotate", d.rateLimiter.Middleware(http.HandlerFunc(d.handleRotateAPIKey)))       // POST rotate API key

	// New routes for environment control
	router.Handle("/api/env", d.rateLimiter.Middleware(http.HandlerFunc(d.handleGetEnvironmentState)))     // GET current environment state
	router.Handle("/api/env/set", d.rateLimiter.Middleware(http.HandlerFunc(d.handleSetEnvironmentState))) // POST set environment state

	// Serve static files (CSS, JS)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/dashboard/static"))))

	// Panel rendering routes
	router.HandleFunc("/dashboard/panels/overview", d.handleOverviewPanel)
	router.HandleFunc("/dashboard/panels/file_processing", d.handleFileProcessingPanel)
	router.HandleFunc("/dashboard/panels/users", d.handleUsersPanel)
	router.HandleFunc("/dashboard/panels/db_config", d.handleDbConfigPanel)
	router.HandleFunc("/dashboard/panels/api_keys", d.handleAPIKeysPanel)
	router.HandleFunc("/dashboard/panels/service_status", d.handleServiceStatusPanel)
	router.HandleFunc("/dashboard/panels/ai_providers", d.handleAIProvidersPanel)
	router.HandleFunc("/dashboard/panels/stats", d.handleStatsPanel)
	router.HandleFunc("/dashboard/panels/chat_history", d.handleChatHistoryPanel)
	router.HandleFunc("/dashboard/panels/environment", d.handleEnvironmentPanel)
	router.HandleFunc("/dashboard/panels/whatsapp", d.handleWhatsAppPanel)
	router.HandleFunc("/dashboard/panels/qa_console", d.handleQAConsolePanel)
	router.Handle("/api/qa", d.rateLimiter.Middleware(http.HandlerFunc(d.handleQA)))
}

// handleQA handles the Q&A requests from the dashboard console.
func (d *Dashboard) handleQA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	question := r.FormValue("question")
	if question == "" {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a simple chat request
	req := &ai.RequestModel{
		UserPrompt: question,
	}

	// Stream the response back
	slog.Info("Handling QA request", "question", question)
	err := d.aiService.Chat(r.Context(), req, func(chunk string) {
		fmt.Fprint(w, chunk)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})

	if err != nil {
		slog.Error("QA request failed", "error", err)
		fmt.Fprintf(w, "<div class='text-red-500'>Error: %v</div>", err)
	}
}

// handleServiceStatusPanel serves the ServiceStatusPanel HTML fragment.
func (d *Dashboard) handleServiceStatusPanel(w http.ResponseWriter, r *http.Request) {
	services := status.GetServicesStatus(d.aiService, d.rcm)
	ServiceStatusPanel(services).Render(r.Context(), w)
}

// handleAIProvidersPanel serves the AIProviderManagementPanel HTML fragment.
func (d *Dashboard) handleAIProvidersPanel(w http.ResponseWriter, r *http.Request) {
	config := d.rcm.GetConfig()
	AIProviderManagementPanel(config).Render(r.Context(), w)
}

// handleStatsPanel serves the StatsPanel HTML fragment.
func (d *Dashboard) handleStatsPanel(w http.ResponseWriter, r *http.Request) {
	stats := status.GetStats(d.rcm)
	StatsPanel(stats).Render(r.Context(), w)
}

// handleChatHistoryPanel serves the ChatHistoryPanel HTML fragment.
func (d *Dashboard) handleChatHistoryPanel(w http.ResponseWriter, r *http.Request) {
	// For now, let's fetch global history or just a default user
	// In a real scenario, we might want a user selector
	history, err := database.GetChatHistory(0, 50) // UserID 0 as placeholder for "all" or just demo
	if err != nil {
		// Fallback if user_id 0 doesn't exist or we want all
		// Let's try to fetch all if userID filter is problematic
		rows, err := d.db.Query("SELECT id, user_id, chat_id, message_id, direction, content_type, text_content, file_path, created_at FROM chat_history ORDER BY created_at DESC LIMIT 100")
		if err == nil {
			defer rows.Close()
			var messages []database.ChatMessage
			for rows.Next() {
				var msg database.ChatMessage
				if err := rows.Scan(&msg.ID, &msg.UserID, &msg.ChatID, &msg.MessageID, &msg.Direction, &msg.ContentType, &msg.TextContent, &msg.FilePath, &msg.CreatedAt); err != nil {
					slog.Error("Failed to scan chat history row", "error", err)
					continue
				}
				messages = append(messages, msg)
			}
			ChatHistoryPanel(messages).Render(r.Context(), w)
			return
		}
	}
	ChatHistoryPanel(history).Render(r.Context(), w)
}

// handleEnvironmentPanel serves the EnvironmentPanel HTML fragment.
func (d *Dashboard) handleEnvironmentPanel(w http.ResponseWriter, r *http.Request) {
	config := d.rcm.GetConfig()
	EnvironmentPanel(config.Environment).Render(r.Context(), w)
}

// handleWhatsAppPanel serves the WhatsAppPanel HTML fragment.
func (d *Dashboard) handleWhatsAppPanel(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling WhatsApp panel request")
	WhatsAppPanel().Render(r.Context(), w)
}

// handleQAConsolePanel serves the QAConsolePanel HTML fragment.
func (d *Dashboard) handleQAConsolePanel(w http.ResponseWriter, r *http.Request) {
	QAConsolePanel().Render(r.Context(), w)
}

// handleAPIKeysPanel serves the APIKeysPanel HTML fragment.
func (d *Dashboard) handleAPIKeysPanel(w http.ResponseWriter, r *http.Request) {
	config := d.rcm.GetConfig()
	var apiKeysSlice []state.APIKeyState
	for _, key := range config.APIKeys {
		apiKeysSlice = append(apiKeysSlice, key)
	}

	// List of all supported providers to ensure they are available in the dropdown
	providers := []string{"Gemini", "Groq", "Hugging Face", "OpenRouter"}

	APIKeysPanel(apiKeysSlice, providers).Render(r.Context(), w)
}

// handleDashboardRedirect redirects / to /dashboard/overview
func (d *Dashboard) handleDashboardRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard/overview", http.StatusFound)
}

// renderDashboardPage renders the main App component with the specified active tab.
func (d *Dashboard) renderDashboardPage(w http.ResponseWriter, r *http.Request, tab string) {
	App(tab).Render(r.Context(), w)
}

// handleOverviewPanel serves the OverviewPanel HTML fragment.
func (d *Dashboard) handleOverviewPanel(w http.ResponseWriter, r *http.Request) {
	services := status.GetServicesStatus(d.aiService, d.rcm)
	providers := d.getAIProviders()
	session := getSessionUser(r.Context())
	OverviewPanel(services, providers, session).Render(r.Context(), w)
}

// handleFileProcessingPanel serves the FileProcessingPanel HTML fragment.
func (d *Dashboard) handleFileProcessingPanel(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50" // Default safe limit
	}
	query := fmt.Sprintf("SELECT id, hash, category, timestamp, extracted_text, summary, topics, questions, ai_provider FROM processed_files ORDER BY timestamp DESC LIMIT %s", limit)
	rows, err := d.db.Query(query)
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

	// Refresh AI service
	d.aiService.RefreshProviders(context.Background())

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

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		// Assume form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		req.ProviderName = r.FormValue("providerName")
		req.KeyValue = r.FormValue("keyValue")
		req.Enabled = true // Default to enabled for form submissions
	}

	if req.ProviderName == "" || strings.TrimSpace(req.KeyValue) == "" {
		http.Error(w, "Provider name and API key are required", http.StatusBadRequest)
		return
	}

	keyID, err := d.rcm.AddAPIKey(req.ProviderName, strings.TrimSpace(req.KeyValue), req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Return 400 for validation errors
		return
	}

	// Refresh AI service
	d.aiService.RefreshProviders(context.Background())

	// If it's an HTMX request, we might want to return the updated panel instead of JSON
	if r.Header.Get("HX-Request") == "true" {
		d.handleAPIKeysPanel(w, r)
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

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		// Assume form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		req.KeyID = r.FormValue("keyID")
	}

	if req.KeyID == "" {
		http.Error(w, "Key ID is required", http.StatusBadRequest)
		return
	}

	err := d.rcm.RemoveAPIKey(req.KeyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Refresh AI service
	d.aiService.RefreshProviders(context.Background())

	// If it's an HTMX request, we might want to return the updated panel instead of JSON
	if r.Header.Get("HX-Request") == "true" {
		d.handleAPIKeysPanel(w, r)
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

	// Refresh AI service
	d.aiService.RefreshProviders(context.Background())

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

	// Refresh AI service
	d.aiService.RefreshProviders(context.Background())

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

func (d *Dashboard) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	LoginPage().Render(r.Context(), w)
}

func (d *Dashboard) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := d.authService.GetLoginURL("state-token") // TODO: secure state
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (d *Dashboard) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("ENVIRONMENT_MODE") != "dev" {
		http.Error(w, "Dev login only available in dev mode", http.StatusForbidden)
		return
	}

	session, _ := d.authService.CreateDevSession()
	cookie, err := d.authService.CreateSessionCookie(session)
	if err != nil {
		http.Error(w, "Failed to create dev session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (d *Dashboard) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	session, err := d.authService.HandleCallback(r.Context(), code)
	if err != nil {
		http.Error(w, "Auth failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cookie, err := d.authService.CreateSessionCookie(session)
	if err != nil {
		http.Error(w, "Session creation failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (d *Dashboard) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

func (d *Dashboard) handleGCPListProjects(w http.ResponseWriter, r *http.Request) {
	session := getSessionUser(r.Context())
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	httpClient, err := d.authService.GetClientForUser(r.Context(), session.GoogleID)
	if err != nil {
		http.Error(w, "Failed to get GCP client", http.StatusInternalServerError)
		return
	}

	gcpClient := gcp.NewClient(httpClient)
	projects, err := gcpClient.ListProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (d *Dashboard) handleGCPListKeys(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectId")
	if projectID == "" {
		http.Error(w, "Missing projectId", http.StatusBadRequest)
		return
	}

	session := getSessionUser(r.Context())
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	httpClient, err := d.authService.GetClientForUser(r.Context(), session.GoogleID)
	if err != nil {
		http.Error(w, "Failed to get GCP client", http.StatusInternalServerError)
		return
	}

	gcpClient := gcp.NewClient(httpClient)
	keys, err := gcpClient.ListAPIKeys(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (d *Dashboard) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	// Telegram Login Widget callback logic
	// In a production app, we would verify the hash/signature.
	// For this implementation, we take the telegram_id and email from the session.
	
	telegramIDStr := r.URL.Query().Get("id")
	if telegramIDStr == "" {
		http.Error(w, "Missing telegram id", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid telegram id", http.StatusBadRequest)
		return
	}

	session := getSessionUser(r.Context())
	if session == nil {
		// If not logged in, we might need a temporary token flow
		http.Error(w, "Session required to link account", http.StatusUnauthorized)
		return
	}

	slog.Info("Linking accounts via webhook", "email", session.Email, "telegram_id", telegramID)
	
	if err := database.LinkTelegramToEmail(telegramID, session.Email); err != nil {
		slog.Error("Webhook linking failed", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?linked=true", http.StatusFound)
}


