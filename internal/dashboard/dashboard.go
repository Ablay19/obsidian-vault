package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/gcp"
	"obsidian-automation/internal/security"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/telemetry"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
func (d *Dashboard) RegisterRoutes(router *gin.Engine) {
	// Static files
	router.Static("/static", "internal/dashboard/static")

	// Auth routes (unprotected)
	authRoutes := router.Group("/auth")
	{
		authRoutes.GET("/login", d.handleLoginPage)
		authRoutes.GET("/google/login", d.handleGoogleLogin)
		authRoutes.GET("/dev/login", d.handleDevLogin)
		authRoutes.GET("/google/callback", d.handleGoogleCallback)
		authRoutes.GET("/logout", d.handleLogout)
	}

	// API routes
	api := router.Group("/api")
	api.Use(d.rateLimiter.GinMiddleware())
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/auth/telegram/webhook", d.handleTelegramWebhook)
		}

		// GCP Discovery routes
		gcpRoutes := api.Group("/auth/google")
		gcpRoutes.Use(d.authService.GinMiddleware())
		{
			gcpRoutes.GET("/list-projects", d.handleGCPListProjects)
			gcpRoutes.GET("/list-keys", d.handleGCPListKeys)
		}

		api.GET("/services/status", d.handleServicesStatus)

		aiRoutes := api.Group("/ai")
		aiRoutes.Use(d.authService.GinMiddleware())
		{
			aiRoutes.GET("/providers", d.handleGetAIProviders)
			aiRoutes.GET("/provider/config", d.handleGetProviderConfig)
			aiRoutes.POST("/provider/set", d.handleSetAIProvider)
			aiRoutes.POST("/provider/toggle", d.handleToggleProviderStatus)
			aiRoutes.GET("/keys", d.handleGetAPIKeys)
			aiRoutes.POST("/key/add", d.handleAddAPIKey)
			aiRoutes.POST("/key/remove", d.handleRemoveAPIKey)
			aiRoutes.POST("/key/toggle", d.handleToggleAPIKeyStatus)
			aiRoutes.POST("/key/rotate", d.handleRotateAPIKey)
		}

		envRoutes := api.Group("/env")
		envRoutes.Use(d.authService.GinMiddleware())
		{
			envRoutes.GET("/", d.handleGetEnvironmentState)
			envRoutes.POST("/set", d.handleSetEnvironmentState)
		}

		api.POST("/qa", d.authService.GinMiddleware(), d.handleQA)
	}

	// WebSocket
	router.GET("/ws", gin.WrapH(http.HandlerFunc(d.wsManager.HandleWebSocket)))

	// Protected dashboard routes
	dash := router.Group("/")
	dash.Use(d.authService.GinMiddleware())
	{
		dash.GET("/", d.handleDashboardRedirect)
		dash.GET("/dashboard/overview", func(c *gin.Context) { d.renderDashboardPage(c, "overview") })
		dash.GET("/dashboard/status", func(c *gin.Context) { d.renderDashboardPage(c, "status") })
		dash.GET("/dashboard/providers", func(c *gin.Context) { d.renderDashboardPage(c, "providers") })
		dash.GET("/dashboard/keys", func(c *gin.Context) { d.renderDashboardPage(c, "keys") })
		dash.GET("/dashboard/history", func(c *gin.Context) { d.renderDashboardPage(c, "history") })
		dash.GET("/dashboard/chat", func(c *gin.Context) { d.renderDashboardPage(c, "chat") })
		dash.GET("/dashboard/whatsapp", func(c *gin.Context) { d.renderDashboardPage(c, "whatsapp") })
		dash.GET("/dashboard/env", func(c *gin.Context) { d.renderDashboardPage(c, "env") })

		panels := dash.Group("/dashboard/panels")
		{
			panels.GET("/overview", d.handleOverviewPanel)
			panels.GET("/file_processing", d.handleFileProcessingPanel)
			panels.GET("/users", d.handleUsersPanel)
			panels.GET("/db_config", d.handleDbConfigPanel)
			panels.GET("/api_keys", d.handleAPIKeysPanel)
			panels.GET("/service_status", d.handleServiceStatusPanel)
			panels.GET("/ai_providers", d.handleAIProvidersPanel)
			panels.GET("/stats", d.handleStatsPanel)
			panels.GET("/chat_history", d.handleChatHistoryPanel)
			panels.GET("/environment", d.handleEnvironmentPanel)
			panels.GET("/whatsapp", d.handleWhatsAppPanel)
			panels.GET("/qa_console", d.handleQAConsolePanel)
		}
	}
}

// handleQA handles the Q&A requests from the dashboard console.
func (d *Dashboard) handleQA(c *gin.Context) {
	question := c.PostForm("question")
	if question == "" {
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Create a simple chat request
	req := &ai.RequestModel{
		UserPrompt: question,
	}

	// Stream the response back
	telemetry.ZapLogger.Sugar().Info("Handling QA request", "question", question)
	err := d.aiService.Chat(c.Request.Context(), req, func(chunk string) {
		fmt.Fprint(c.Writer, chunk)
		c.Writer.Flush()
	})

	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("QA request failed", "error", err)
		fmt.Fprintf(c.Writer, "<div class='text-red-500'>Error: %v</div>", err)
	}
}

// handleServiceStatusPanel serves the ServiceStatusPanel HTML fragment.
func (d *Dashboard) handleServiceStatusPanel(c *gin.Context) {
	services := status.GetServicesStatus(d.aiService, d.rcm)
	ServiceStatusPanel(services).Render(c.Request.Context(), c.Writer)
}

// handleAIProvidersPanel serves the AIProviderManagementPanel HTML fragment.
func (d *Dashboard) handleAIProvidersPanel(c *gin.Context) {
	config := d.rcm.GetConfig()
	AIProviderManagementPanel(config).Render(c.Request.Context(), c.Writer)
}

// handleStatsPanel serves the StatsPanel HTML fragment.
func (d *Dashboard) handleStatsPanel(c *gin.Context) {
	stats := status.GetStats(d.rcm)
	StatsPanel(stats).Render(c.Request.Context(), c.Writer)
}

// handleChatHistoryPanel serves the ChatHistoryPanel HTML fragment.
func (d *Dashboard) handleChatHistoryPanel(c *gin.Context) {
	history, err := database.GetChatHistory(0, 50) // UserID 0 as placeholder for "all" or just demo
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to get chat history", "error", err)
	}
	ChatHistoryPanel(history).Render(c.Request.Context(), c.Writer)
}

// handleEnvironmentPanel serves the EnvironmentPanel HTML fragment.
func (d *Dashboard) handleEnvironmentPanel(c *gin.Context) {
	config := d.rcm.GetConfig()
	EnvironmentPanel(config.Environment).Render(c.Request.Context(), c.Writer)
}

// handleWhatsAppPanel serves the WhatsAppPanel HTML fragment.
func (d *Dashboard) handleWhatsAppPanel(c *gin.Context) {
	telemetry.ZapLogger.Sugar().Info("Handling WhatsApp panel request")
	WhatsAppPanel().Render(c.Request.Context(), c.Writer)
}

// handleQAConsolePanel serves the QAConsolePanel HTML fragment.
func (d *Dashboard) handleQAConsolePanel(c *gin.Context) {
	QAConsolePanel().Render(c.Request.Context(), c.Writer)
}

// handleAPIKeysPanel serves the APIKeysPanel HTML fragment.
func (d *Dashboard) handleAPIKeysPanel(c *gin.Context) {
	config := d.rcm.GetConfig()
	var apiKeysSlice []state.APIKeyState
	for _, key := range config.APIKeys {
		apiKeysSlice = append(apiKeysSlice, key)
	}

	providers := []string{"Gemini", "Groq", "Hugging Face", "OpenRouter"}

	APIKeysPanel(apiKeysSlice, providers).Render(c.Request.Context(), c.Writer)
}

// handleDashboardRedirect redirects / to /dashboard/overview
func (d *Dashboard) handleDashboardRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "/dashboard/overview")
}

// renderDashboardPage renders the main App component with the specified active tab.
func (d *Dashboard) renderDashboardPage(c *gin.Context, tab string) {
	App(tab).Render(c.Request.Context(), c.Writer)
}

// handleOverviewPanel serves the OverviewPanel HTML fragment.
func (d *Dashboard) handleOverviewPanel(c *gin.Context) {
	services := status.GetServicesStatus(d.aiService, d.rcm)
	providers := d.getAIProviders()
	session, _ := c.Get("session")
	OverviewPanel(services, providers, session.(*auth.UserSession)).Render(c.Request.Context(), c.Writer)
}

// handleFileProcessingPanel serves the FileProcessingPanel HTML fragment.
func (d *Dashboard) handleFileProcessingPanel(c *gin.Context) {
	limit := c.Query("limit")
	if limit == "" {
		limit = "50" // Default safe limit
	}
	query := fmt.Sprintf("SELECT id, hash, category, timestamp, extracted_text, summary, topics, questions, ai_provider FROM processed_files ORDER BY timestamp DESC LIMIT %s", limit)
	rows, err := d.db.Query(query)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to query database")
		return
	}
	defer rows.Close()

	var files []ProcessedFile
	for rows.Next() {
		var file ProcessedFile
		if err := rows.Scan(&file.ID, &file.Hash, &file.Category, &file.Timestamp, &file.ExtractedText, &file.Summary, &file.Topics, &file.Questions, &file.AiProvider); err != nil {
			c.String(http.StatusInternalServerError, "Failed to scan database rows")
			return
		}
		files = append(files, file)
	}

	FileProcessingPanel(files).Render(c.Request.Context(), c.Writer)
}

// handleUsersPanel serves the UsersPanel HTML fragment.
func (d *Dashboard) handleUsersPanel(c *gin.Context) {
	rows, err := d.db.Query("SELECT id, username, first_name, last_name, language_code, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to query database")
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.LanguageCode, &user.CreatedAt); err != nil {
			c.String(http.StatusInternalServerError, "Failed to scan database rows")
			return
		}
		users = append(users, user)
	}

	UsersPanel(users).Render(c.Request.Context(), c.Writer)
}

// handleDbConfigPanel serves the DbConfigPanel HTML fragment.
func (d *Dashboard) handleDbConfigPanel(c *gin.Context) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	u, err := url.Parse(dbURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse database URL")
		return
	}

	dbType := u.Scheme
	dbHost := u.Host

	DbConfigPanel(dbType, dbHost).Render(c.Request.Context(), c.Writer)
}

// handleServicesStatus provides the status of all monitored services.
func (d *Dashboard) handleServicesStatus(c *gin.Context) {
	statuses := status.GetServicesStatus(d.aiService, d.rcm)
	c.JSON(http.StatusOK, statuses)
}

// handleGetAIProviders returns the available and active AI providers.
func (d *Dashboard) handleGetAIProviders(c *gin.Context) {
	providers := d.getAIProviders()
	c.JSON(http.StatusOK, providers)
}

func (d *Dashboard) getAIProviders() struct {
	Available []string `json:"available"`
	Active    string   `json:"active"`
} {
	if d.aiService == nil {
		return struct {
			Available []string `json:"available"`
			Active    string   `json:"active"`
		}{}
	}

	config := d.rcm.GetConfig()
	var availableProviders []string
	for name, ps := range config.Providers {
		if ps.Enabled {
			availableProviders = append(availableProviders, name)
		}
	}
	activeProviderName := d.aiService.GetActiveProviderName()
	return struct {
		Available []string `json:"available"`
		Active    string   `json:"active"`
	}{
		Available: availableProviders,
		Active:    activeProviderName,
	}
}

// handleSetAIProvider sets the active AI provider.
func (d *Dashboard) handleSetAIProvider(c *gin.Context) {
	if d.aiService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service not available"})
		return
	}

	var req struct {
		Provider string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := d.aiService.SetProvider(req.Provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("AI provider set to %s", req.Provider)})
}

// handleGetProviderConfig returns the configuration for a specific provider.
func (d *Dashboard) handleGetProviderConfig(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'provider' query parameter"})
		return
	}

	config := d.rcm.GetConfig()
	providerState, providerExists := config.Providers[providerName]
	if !providerExists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider '%s' not found", providerName)})
		return
	}

	var apiKeys []state.APIKeyState
	for _, keyState := range config.APIKeys {
		if keyState.Provider == providerName {
			apiKeys = append(apiKeys, keyState)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": providerState,
		"apiKeys":  apiKeys,
	})
}

// handleToggleProviderStatus handles POST requests to toggle the status of an AI provider.
func (d *Dashboard) handleToggleProviderStatus(c *gin.Context) {
	var req struct {
		Provider      string `json:"provider"`
		Enabled       bool   `json:"enabled"`
		Paused        bool   `json:"paused"`
		Blocked       bool   `json:"blocked"`
		BlockedReason string `json:"blockedReason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := d.rcm.SetProviderState(req.Provider, req.Enabled, req.Paused, req.Blocked, req.BlockedReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.aiService.InitializeProviders(context.Background())
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("Provider '%s' status updated.", req.Provider)})
}

// handleGetAPIKeys returns all API keys, optionally filtered by provider.
func (d *Dashboard) handleGetAPIKeys(c *gin.Context) {
	providerName := c.Query("provider")
	config := d.rcm.GetConfig()
	var apiKeys []state.APIKeyState
	for _, keyState := range config.APIKeys {
		if providerName == "" || keyState.Provider == providerName {
			apiKeys = append(apiKeys, keyState)
		}
	}
	c.JSON(http.StatusOK, apiKeys)
}

// handleAddAPIKey handles POST requests to add a new API key.
func (d *Dashboard) handleAddAPIKey(c *gin.Context) {
	var req struct {
		ProviderName string `form:"providerName" json:"providerName"`
		KeyValue     string `form:"keyValue" json:"keyValue"`
		Enabled      bool   `form:"enabled" json:"enabled"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.ProviderName == "" || strings.TrimSpace(req.KeyValue) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider name and API key are required"})
		return
	}

	keyID, err := d.rcm.AddAPIKey(req.ProviderName, strings.TrimSpace(req.KeyValue), req.Enabled)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d.aiService.InitializeProviders(context.Background())

	if c.GetHeader("HX-Request") == "true" {
		d.handleAPIKeysPanel(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "API key added", "keyID": keyID})
}

// handleRemoveAPIKey handles POST requests to remove an API key.
func (d *Dashboard) handleRemoveAPIKey(c *gin.Context) {
	var req struct {
		KeyID string `form:"keyID" json:"keyID"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.KeyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key ID is required"})
		return
	}

	err := d.rcm.RemoveAPIKey(req.KeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.aiService.InitializeProviders(context.Background())

	if c.GetHeader("HX-Request") == "true" {
		d.handleAPIKeysPanel(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("API key '%s' removed.", req.KeyID)})
}

// handleToggleAPIKeyStatus handles POST requests to toggle the status of an API key.
func (d *Dashboard) handleToggleAPIKeyStatus(c *gin.Context) {
	var req struct {
		KeyID         string `json:"keyID"`
		Enabled       bool   `json:"enabled"`
		Blocked       bool   `json:"blocked"`
		BlockedReason string `json:"blockedReason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := d.rcm.SetAPIKeyStatus(req.KeyID, req.Enabled, req.Blocked, req.BlockedReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.aiService.InitializeProviders(context.Background())
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("API key '%s' status updated.", req.KeyID)})
}

// handleRotateAPIKey handles POST requests to rotate the active API key for a provider.
func (d *Dashboard) handleRotateAPIKey(c *gin.Context) {
	var req struct {
		ProviderName string `json:"providerName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := d.rcm.RotateAPIKey(req.ProviderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	d.aiService.InitializeProviders(context.Background())
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("API key for provider '%s' rotated.", req.ProviderName)})
}

// handleGetEnvironmentState handles GET requests to retrieve the current environment state.
func (d *Dashboard) handleGetEnvironmentState(c *gin.Context) {
	config := d.rcm.GetConfig()
	envState := config.Environment
	c.JSON(http.StatusOK, envState)
}

// handleSetEnvironmentState handles POST requests to update the bot's operational environment.
func (d *Dashboard) handleSetEnvironmentState(c *gin.Context) {
	var req struct {
		Mode             string `json:"mode"`
		BackendHost      string `json:"backendHost"`
		IsolationEnabled bool   `json:"isolationEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := d.rcm.SetEnvironmentState(req.Mode, req.BackendHost, req.IsolationEnabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("Environment state updated to mode '%s'.", req.Mode)})
}

func (d *Dashboard) handleLoginPage(c *gin.Context) {
	LoginPage().Render(c.Request.Context(), c.Writer)
}

func (d *Dashboard) handleGoogleLogin(c *gin.Context) {
	url := d.authService.GetLoginURL("state-token") // TODO: secure state
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (d *Dashboard) handleDevLogin(c *gin.Context) {
	if os.Getenv("ENVIRONMENT_MODE") != "dev" {
		c.String(http.StatusForbidden, "Dev login only available in dev mode")
		return
	}

	session, _ := d.authService.CreateDevSession()
	cookie, err := d.authService.CreateSessionCookie(session)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create dev session")
		return
	}

	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/")
}

func (d *Dashboard) handleGoogleCallback(c *gin.Context) {
	code := c.Query("code")
	session, err := d.authService.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Auth failed: "+err.Error())
		return
	}

	cookie, err := d.authService.CreateSessionCookie(session)
	if err != nil {
		c.String(http.StatusInternalServerError, "Session creation failed")
		return
	}

	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/")
}

func (d *Dashboard) handleLogout(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, cookie)
	c.Redirect(http.StatusFound, "/auth/login")
}

func (d *Dashboard) handleGCPListProjects(c *gin.Context) {
	session, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	httpClient, err := d.authService.GetClientForUser(c.Request.Context(), session.(*auth.UserSession).GoogleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GCP client"})
		return
	}

	gcpClient := gcp.NewClient(httpClient)
	projects, err := gcpClient.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (d *Dashboard) handleGCPListKeys(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing projectId"})
		return
	}

	session, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	httpClient, err := d.authService.GetClientForUser(c.Request.Context(), session.(*auth.UserSession).GoogleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GCP client"})
		return
	}

	gcpClient := gcp.NewClient(httpClient)
	keys, err := gcpClient.ListAPIKeys(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, keys)
}

func (d *Dashboard) handleTelegramWebhook(c *gin.Context) {
	telegramIDStr := c.Query("id")
	if telegramIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing telegram id"})
		return
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telegram id"})
		return
	}

	session, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session required to link account"})
		return
	}
	userSession := session.(*auth.UserSession)

	telemetry.ZapLogger.Sugar().Info("Linking accounts via webhook", "email", userSession.Email, "telegram_id", telegramID)

	if err := database.LinkTelegramToEmail(telegramID, userSession.Email); err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Webhook linking failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.Redirect(http.StatusFound, "/?linked=true")
}
