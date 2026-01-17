package transports

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// SMAposShipperTransport implements SM APOS Shipper transport
type SMAposShipperTransport struct {
	config      *utils.Config
	logger      *log.Logger
	httpClient  *http.Client
	baseURL     string
	sessions    map[string]*models.ShipperSession
	rateLimiter *utils.RateLimiter
}

// NewSMAposShipperTransport creates a new SM APOS Shipper transport
func NewSMAposShipperTransport(config *utils.Config, logger *log.Logger) (*SMAposShipperTransport, error) {
	// Get shipper config
	shipperConfig := config.Transports.Shipper

	transport := &SMAposShipperTransport{
		config:   config,
		logger:   logger,
		baseURL:  shipperConfig.BaseURL,
		sessions: make(map[string]*models.ShipperSession),
	}

	// Configure HTTP client with appropriate timeouts and TLS settings
	transport.httpClient = &http.Client{
		Timeout: 60 * time.Second, // Longer timeout for network operations
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, // Enable TLS verification
			},
		},
	}

	// Initialize rate limiter (SM APOS Shipper allows 100 requests per hour)
	transport.rateLimiter = utils.NewRateLimiter(100, time.Hour, logger)

	// If no endpoint configured, use default
	if transport.baseURL == "" {
		transport.baseURL = "https://api.mauritania-shipper.mr/v1"
	}

	return transport, nil
}

// Authenticate establishes a session with SM APOS Shipper
func (s *SMAposShipperTransport) Authenticate(credentials map[string]string) (*models.ShipperSession, error) {
	// Check rate limit
	if s.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	authPayload := map[string]interface{}{
		"username": credentials["username"],
		"password": credentials["password"],
		"service":  "command_execution",
	}

	jsonData, err := json.Marshal(authPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	url := fmt.Sprintf("%s/auth/login", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mauritania-CLI/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("Auth failed with status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var authResponse struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
		Session struct {
			ID          string   `json:"id"`
			UserID      string   `json:"user_id"`
			ExpiresAt   string   `json:"expires_at"`
			Permissions []string `json:"permissions"`
		} `json:"session"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !authResponse.Success {
		return nil, fmt.Errorf("authentication failed: %s", authResponse.Message)
	}

	// Parse expiration time
	expiresAt, err := time.Parse(time.RFC3339, authResponse.Session.ExpiresAt)
	if err != nil {
		s.logger.Printf("Failed to parse expiration time, using default 1 hour")
		expiresAt = time.Now().Add(time.Hour)
	}

	session := &models.ShipperSession{
		ID:          authResponse.Session.ID,
		UserID:      authResponse.Session.UserID,
		Token:       authResponse.Token, // This should be encrypted in production
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Permissions: authResponse.Session.Permissions,
	}

	// Store session
	s.sessions[session.ID] = session

	// Record rate limit usage
	s.rateLimiter.RecordUsage()

	s.logger.Printf("Authenticated user %s with session %s", session.UserID, session.ID)
	return session, nil
}

// ExecuteCommand sends a command for execution through SM APOS Shipper
func (s *SMAposShipperTransport) ExecuteCommand(session *models.ShipperSession, command string, timeout int) (*models.CommandResult, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	// Check if session is valid
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Check rate limit
	if s.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Validate command (basic check)
	if command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	executePayload := map[string]interface{}{
		"command":     command,
		"timeout":     timeout,
		"environment": "bash", // Default shell
		"working_dir": "/tmp", // Default working directory
	}

	jsonData, err := json.Marshal(executePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execute payload: %w", err)
	}

	url := fmt.Sprintf("%s/commands/execute", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create execute request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("X-Session-ID", session.ID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read execute response: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		s.logger.Printf("Execute failed with status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("command execution failed with status %d", resp.StatusCode)
	}

	var executeResponse struct {
		Success       bool   `json:"success"`
		CommandID     string `json:"command_id"`
		Status        string `json:"status"`
		EstimatedTime int    `json:"estimated_time"`
		Message       string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(body, &executeResponse); err != nil {
		return nil, fmt.Errorf("failed to parse execute response: %w", err)
	}

	if !executeResponse.Success {
		return nil, fmt.Errorf("command execution rejected: %s", executeResponse.Message)
	}

	// Record rate limit usage
	s.rateLimiter.RecordUsage()

	result := &models.CommandResult{
		ID:            executeResponse.CommandID,
		CommandID:     executeResponse.CommandID,
		Status:        executeResponse.Status,
		ExitCode:      -1, // Not completed yet
		Stdout:        "",
		Stderr:        "",
		ExecutionTime: 0,
		TransportUsed: "sm_apos",
		Cost:          0.0,         // Will be updated when complete
		CompletedAt:   time.Time{}, // Not completed yet
	}

	s.logger.Printf("Command submitted with ID %s, status: %s", result.ID, result.Status)
	return result, nil
}

// GetCommandStatus checks the status of a running command
func (s *SMAposShipperTransport) GetCommandStatus(session *models.ShipperSession, commandID string) (*models.ShipperCommandStatus, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	// Check rate limit
	if s.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	url := fmt.Sprintf("%s/commands/%s/status", s.baseURL, commandID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("X-Session-ID", session.ID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("command not found")
		}
		return nil, fmt.Errorf("status request failed with status %d", resp.StatusCode)
	}

	var statusResponse struct {
		Success   bool   `json:"success"`
		CommandID string `json:"command_id"`
		Status    string `json:"status"`
		Progress  int    `json:"progress"`
		CreatedAt string `json:"created_at"`
		StartedAt string `json:"started_at,omitempty"`
		Message   string `json:"message,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	if !statusResponse.Success {
		return nil, fmt.Errorf("status request failed: %s", statusResponse.Message)
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, statusResponse.CreatedAt)
	var startedAt *time.Time
	if statusResponse.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, statusResponse.StartedAt); err == nil {
			startedAt = &t
		}
	}

	var completedAt *time.Time
	var status models.CommandStatus
	switch statusResponse.Status {
	case "queued":
		status = models.StatusQueued
	case "running", "executing":
		status = models.StatusExecuting
	case "completed", "success":
		status = models.StatusCompleted
		now := time.Now()
		completedAt = &now
	case "failed", "error":
		status = models.StatusFailed
		now := time.Now()
		completedAt = &now
	default:
		status = models.StatusQueued
	}

	// Record rate limit usage
	s.rateLimiter.RecordUsage()

	cmdStatus := &models.ShipperCommandStatus{
		CommandID:   statusResponse.CommandID,
		Status:      status,
		CreatedAt:   createdAt,
		QueuedAt:    &createdAt, // Assume queued at creation
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Progress:    statusResponse.Progress,
		Error:       statusResponse.Error,
	}

	return cmdStatus, nil
}

// CancelCommand cancels a running command
func (s *SMAposShipperTransport) CancelCommand(session *models.ShipperSession, commandID string) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	// Check rate limit
	if s.rateLimiter.IsRateLimited() {
		return fmt.Errorf("rate limit exceeded")
	}

	url := fmt.Sprintf("%s/commands/%s/cancel", s.baseURL, commandID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+session.Token)
	req.Header.Set("X-Session-ID", session.ID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cancel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel request failed with status %d", resp.StatusCode)
	}

	// Record rate limit usage
	s.rateLimiter.RecordUsage()

	s.logger.Printf("Command %s cancelled", commandID)
	return nil
}

// ListActiveSessions returns all active shipper sessions
func (s *SMAposShipperTransport) ListActiveSessions() ([]*models.ShipperSession, error) {
	var activeSessions []*models.ShipperSession

	now := time.Now()
	for _, session := range s.sessions {
		if now.Before(session.ExpiresAt) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// CloseSession terminates a shipper session
func (s *SMAposShipperTransport) CloseSession(sessionID string) error {
	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Call shipper logout endpoint if available
	url := fmt.Sprintf("%s/auth/logout", s.baseURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		s.logger.Printf("Failed to create logout request: %v", err)
	} else {
		req.Header.Set("Authorization", "Bearer "+session.Token)
		req.Header.Set("X-Session-ID", session.ID)

		// Don't wait for response, just fire and forget
		go func() {
			resp, err := s.httpClient.Do(req)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}()
	}

	// Remove from local storage
	delete(s.sessions, sessionID)

	s.logger.Printf("Session %s closed", sessionID)
	return nil
}
