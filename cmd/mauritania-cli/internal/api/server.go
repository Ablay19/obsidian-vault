package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// Server handles HTTP requests for the CLI
type Server struct {
	db            *database.DB
	config        *utils.Config
	logger        *log.Logger
	commandParser *utils.CommandParser
	authValidator *utils.AuthValidator
	port          string
}

// NewServer creates a new HTTP server
func NewServer(db *database.DB, config *utils.Config, logger *log.Logger) *Server {
	// Create crypto manager for authentication
	cryptoManager, err := utils.NewCryptoManager("default-key", utils.DefaultEncryptionConfig())
	if err != nil {
		logger.Printf("Warning: Failed to create crypto manager: %v", err)
		cryptoManager = nil
	}

	return &Server{
		db:            db,
		config:        config,
		logger:        logger,
		commandParser: utils.NewCommandParser(),
		authValidator: utils.NewAuthValidator(config, cryptoManager),
		port:          "3001", // Default port
	}
}

// Start starts the HTTP server

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/commands/execute", s.handleExecuteCommand)
	mux.HandleFunc("/api/v1/commands/", s.handleCommandOperations)
	mux.HandleFunc("/api/v1/queue", s.handleQueueOperations)
	mux.HandleFunc("/api/v1/transports", s.handleGetTransports)
	mux.HandleFunc("/api/v1/network/routes", s.handleGetNetworkRoutes)
	mux.HandleFunc("/api/v1/sessions", s.handleSessionOperations)

	// Webhook routes
	mux.HandleFunc("/webhooks/social-media", s.handleSocialMediaWebhook)
	mux.HandleFunc("/webhooks/whatsapp", s.handleWhatsAppWebhook)
	mux.HandleFunc("/webhooks/telegram", s.handleTelegramWebhook)
	mux.HandleFunc("/webhooks/facebook", s.handleFacebookWebhook)

	// Health check
	mux.HandleFunc("/health", s.handleHealthCheck)

	s.logger.Printf("Starting HTTP server on port %s", s.port)

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server.ListenAndServe()
}

// handleExecuteCommand handles command execution requests
func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Command   string `json:"command"`
		Transport string `json:"transport"`
		Priority  string `json:"priority,omitempty"`
		Timeout   int    `json:"timeout,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse and validate command
	command, err := s.commandParser.ParseCommand(req.Command)
	if err != nil {
		s.logger.Printf("Command parsing failed: %v", err)
		http.Error(w, fmt.Sprintf("Invalid command: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.commandParser.ValidateCommand(command); err != nil {
		s.logger.Printf("Command validation failed: %v", err)
		http.Error(w, fmt.Sprintf("Command validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Create command record
	cmd := &models.SocialMediaCommand{
		ID:        generateID(),
		Command:   command.Command,
		Priority:  command.Priority,
		Status:    models.StatusQueued,
		Timestamp: time.Now(),
	}

	// Save to database
	if err := s.db.SaveCommand(*cmd); err != nil {
		s.logger.Printf("Failed to save command: %v", err)
		http.Error(w, "Failed to queue command", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"commandId":     cmd.ID,
		"status":        cmd.Status,
		"estimatedTime": 30, // Default estimate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCommandOperations handles command status and result queries
func (s *Server) handleCommandOperations(w http.ResponseWriter, r *http.Request) {
	// Extract command ID from URL path
	path := r.URL.Path
	if len(path) < len("/api/v1/commands/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	commandID := path[len("/api/v1/commands/"):]
	if commandID == "" {
		http.Error(w, "Command ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetCommandStatus(w, r, commandID)
	case http.MethodDelete:
		s.handleCancelCommand(w, r, commandID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetCommandStatus returns command status or result
func (s *Server) handleGetCommandStatus(w http.ResponseWriter, r *http.Request, commandID string) {
	cmd, err := s.db.GetCommand(commandID)
	if err != nil {
		s.logger.Printf("Failed to get command %s: %v", commandID, err)
		http.Error(w, "Command not found", http.StatusNotFound)
		return
	}

	if cmd.Status == models.StatusCompleted || cmd.Status == models.StatusFailed {
		// Return full result
		result, err := s.db.GetCommandResult(commandID)
		if err != nil {
			s.logger.Printf("Failed to get result for command %s: %v", commandID, err)
			http.Error(w, "Result not available", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		// Return status only
		status := map[string]interface{}{
			"commandId": cmd.ID,
			"status":    cmd.Status,
			"createdAt": cmd.Timestamp,
			"progress":  50, // Mock progress
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

// handleCancelCommand cancels a queued command
func (s *Server) handleCancelCommand(w http.ResponseWriter, r *http.Request, commandID string) {
	if err := s.db.DeleteCommand(commandID); err != nil {
		s.logger.Printf("Failed to cancel command %s: %v", commandID, err)
		http.Error(w, "Failed to cancel command", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleQueueOperations manages the command queue
func (s *Server) handleQueueOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetQueue(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetQueue returns queued commands
func (s *Server) handleGetQueue(w http.ResponseWriter, r *http.Request) {
	commands, err := s.db.GetQueuedCommands()
	if err != nil {
		s.logger.Printf("Failed to get queued commands: %v", err)
		http.Error(w, "Failed to get queue", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"commands": commands,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetTransports returns available transport mechanisms
func (s *Server) handleGetTransports(w http.ResponseWriter, r *http.Request) {
	transports := []map[string]interface{}{
		{
			"type":        "social_media",
			"name":        "Social Media",
			"available":   true,
			"costPerMB":   0.0,
			"latency":     5000,
			"reliability": 85.0,
		},
		{
			"type":        "sm_apos",
			"name":        "SM APOS Shipper",
			"available":   true,
			"costPerMB":   0.10,
			"latency":     3000,
			"reliability": 95.0,
		},
		{
			"type":        "nrt",
			"name":        "NRT Routing",
			"available":   false,
			"costPerMB":   0.05,
			"latency":     2000,
			"reliability": 90.0,
		},
	}

	response := map[string]interface{}{
		"transports": transports,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetNetworkRoutes returns network routes
func (s *Server) handleGetNetworkRoutes(w http.ResponseWriter, r *http.Request) {
	routes := []models.NetworkRoute{
		{
			ID:          "social-media-1",
			Type:        "social_media",
			Provider:    "WhatsApp",
			CostPerMB:   0.0,
			Bandwidth:   1.0,
			Latency:     5000,
			Reliability: 85.0,
			IsActive:    true,
		},
		{
			ID:          "shipper-1",
			Type:        "sm_apos",
			Provider:    "Mauritel",
			CostPerMB:   0.10,
			Bandwidth:   2.0,
			Latency:     3000,
			Reliability: 95.0,
			IsActive:    true,
		},
	}

	response := map[string]interface{}{
		"routes": routes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSessionOperations manages shipper sessions
func (s *Server) handleSessionOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateSession(w, r)
	case http.MethodGet:
		s.handleGetSessions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateSession creates a new shipper session
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Credentials map[string]string `json:"credentials"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Mock session creation
	session := models.ShipperSession{
		ID:          generateID(),
		UserID:      "user123",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Permissions: []string{"execute", "read"},
	}

	response := map[string]interface{}{
		"session": session,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetSessions returns active sessions
func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []models.ShipperSession{
		{
			ID:          "session1",
			UserID:      "user123",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			ExpiresAt:   time.Now().Add(23 * time.Hour),
			Permissions: []string{"execute", "read"},
		},
	}

	response := map[string]interface{}{
		"sessions": sessions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Webhook handlers
func (s *Server) handleSocialMediaWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleGenericWebhook(w, r, "social_media")
}

func (s *Server) handleWhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleGenericWebhook(w, r, "whatsapp")
}

func (s *Server) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleGenericWebhook(w, r, "telegram")
}

func (s *Server) handleFacebookWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleGenericWebhook(w, r, "facebook")
}

func (s *Server) handleGenericWebhook(w http.ResponseWriter, r *http.Request, platform string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhookData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
		s.logger.Printf("Failed to decode webhook: %v", err)
		http.Error(w, "Invalid webhook data", http.StatusBadRequest)
		return
	}

	s.logger.Printf("Received webhook from %s: %+v", platform, webhookData)

	// Extract command from webhook data (platform-specific)
	commandText, ok := s.extractCommandFromWebhook(webhookData, platform)
	if !ok {
		s.logger.Printf("No command found in %s webhook", platform)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create command
	cmd := &models.SocialMediaCommand{
		ID:        generateID(),
		Platform:  platform,
		Command:   commandText,
		Status:    models.StatusReceived,
		Timestamp: time.Now(),
	}

	// Save to database
	if err := s.db.SaveCommand(*cmd); err != nil {
		s.logger.Printf("Failed to save command from webhook: %v", err)
		http.Error(w, "Failed to process command", http.StatusInternalServerError)
		return
	}

	s.logger.Printf("Processed command from %s webhook: %s", platform, cmd.ID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received", "commandId": cmd.ID})
}

func (s *Server) extractCommandFromWebhook(data map[string]interface{}, platform string) (string, bool) {
	// Platform-specific extraction logic would go here
	// For now, return a mock command
	if msg, ok := data["message"].(string); ok && msg != "" {
		return msg, true
	}
	return "", false
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateID generates a simple ID (in production, use UUID)
func generateID() string {
	return fmt.Sprintf("cmd_%d", time.Now().UnixNano())
}
