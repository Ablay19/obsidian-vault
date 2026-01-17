package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	smapos_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/smapos"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// ShipperSessionManager manages SM APOS Shipper sessions
type ShipperSessionManager struct {
	db            *database.DB
	config        *utils.Config
	logger        *log.Logger
	transport     models.ShipperTransport
	sessions      map[string]*models.ShipperSession
	sessionMutex  sync.RWMutex
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewShipperSessionManager creates a new session manager
func NewShipperSessionManager(db *database.DB, config *utils.Config, logger *log.Logger) (*ShipperSessionManager, error) {
	// Initialize transport
	transport, err := smapos_transport.NewSMAposShipperTransport(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipper transport: %w", err)
	}

	manager := &ShipperSessionManager{
		db:          db,
		config:      config,
		logger:      logger,
		transport:   transport,
		sessions:    make(map[string]*models.ShipperSession),
		stopCleanup: make(chan struct{}),
	}

	// Load existing sessions from database
	if err := manager.loadSessions(); err != nil {
		logger.Printf("Warning: Failed to load existing sessions: %v", err)
	}

	// Start cleanup routine
	manager.startCleanupRoutine()

	return manager, nil
}

// AuthenticateUser authenticates a user with SM APOS Shipper
func (sm *ShipperSessionManager) AuthenticateUser(username, password string) (*models.ShipperSession, error) {
	sm.logger.Printf("Authenticating user: %s", username)

	credentials := map[string]string{
		"username": username,
		"password": password,
	}

	session, err := sm.transport.Authenticate(credentials)
	if err != nil {
		sm.logger.Printf("Authentication failed for user %s: %v", username, err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Store session
	sm.sessionMutex.Lock()
	sm.sessions[session.ID] = session
	sm.sessionMutex.Unlock()

	// Save to database
	if err := sm.saveSession(session); err != nil {
		sm.logger.Printf("Warning: Failed to save session to database: %v", err)
	}

	sm.logger.Printf("User %s authenticated successfully with session %s", username, session.ID)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *ShipperSessionManager) GetSession(sessionID string) (*models.ShipperSession, error) {
	sm.sessionMutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.sessionMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is still valid
	if time.Now().After(session.ExpiresAt) {
		sm.logger.Printf("Session %s has expired", sessionID)
		sm.RemoveSession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// GetUserSession gets the active session for a user
func (sm *ShipperSessionManager) GetUserSession(userID string) (*models.ShipperSession, error) {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	for _, session := range sm.sessions {
		if session.UserID == userID && time.Now().Before(session.ExpiresAt) {
			return session, nil
		}
	}

	return nil, fmt.Errorf("no active session found for user %s", userID)
}

// RefreshSession refreshes an expiring session
func (sm *ShipperSessionManager) RefreshSession(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Check if session needs refresh (expires within 5 minutes)
	if time.Until(session.ExpiresAt) > 5*time.Minute {
		return nil // No refresh needed
	}

	sm.logger.Printf("Refreshing session %s", sessionID)

	// For now, we don't have a refresh endpoint, so we need to re-authenticate
	// In a real implementation, you'd call a refresh endpoint
	return fmt.Errorf("session refresh not implemented - please re-authenticate")
}

// RemoveSession removes a session
func (sm *ShipperSessionManager) RemoveSession(sessionID string) error {
	sm.sessionMutex.Lock()
	defer sm.sessionMutex.Unlock()

	_, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Close session with the transport
	if err := sm.transport.CloseSession(sessionID); err != nil {
		sm.logger.Printf("Warning: Failed to close session with transport: %v", err)
	}

	// Remove from memory
	delete(sm.sessions, sessionID)

	// Remove from database
	if err := sm.removeSessionFromDB(sessionID); err != nil {
		sm.logger.Printf("Warning: Failed to remove session from database: %v", err)
	}

	sm.logger.Printf("Session %s removed", sessionID)
	return nil
}

// ListActiveSessions returns all active sessions
func (sm *ShipperSessionManager) ListActiveSessions() ([]*models.ShipperSession, error) {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	var activeSessions []*models.ShipperSession
	now := time.Now()

	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// ValidateSessionPermissions checks if a session has required permissions
func (sm *ShipperSessionManager) ValidateSessionPermissions(sessionID string, requiredPermissions []string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	for _, required := range requiredPermissions {
		found := false
		for _, permission := range session.Permissions {
			if permission == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("session lacks required permission: %s", required)
		}
	}

	return nil
}

// GetSessionStats returns session statistics
func (sm *ShipperSessionManager) GetSessionStats() map[string]interface{} {
	sm.sessionMutex.RLock()
	defer sm.sessionMutex.RUnlock()

	now := time.Now()
	total := len(sm.sessions)
	active := 0
	expiringSoon := 0

	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			active++
			if time.Until(session.ExpiresAt) < 10*time.Minute {
				expiringSoon++
			}
		}
	}

	return map[string]interface{}{
		"total_sessions":   total,
		"active_sessions":  active,
		"expired_sessions": total - active,
		"expiring_soon":    expiringSoon,
	}
}

// loadSessions loads sessions from database on startup
func (sm *ShipperSessionManager) loadSessions() error {
	// In a real implementation, you'd load sessions from database
	// For now, sessions are only stored in memory
	return nil
}

// saveSession saves a session to database
func (sm *ShipperSessionManager) saveSession(session *models.ShipperSession) error {
	// In a real implementation, you'd encrypt the token and save to database
	// For now, sessions are only stored in memory
	return nil
}

// removeSessionFromDB removes a session from database
func (sm *ShipperSessionManager) removeSessionFromDB(sessionID string) error {
	// In a real implementation, you'd remove from database
	return nil
}

// startCleanupRoutine starts a background routine to clean up expired sessions
func (sm *ShipperSessionManager) startCleanupRoutine() {
	sm.cleanupTicker = time.NewTicker(5 * time.Minute) // Clean up every 5 minutes

	go func() {
		for {
			select {
			case <-sm.cleanupTicker.C:
				sm.cleanupExpiredSessions()
			case <-sm.stopCleanup:
				sm.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanupExpiredSessions removes expired sessions
func (sm *ShipperSessionManager) cleanupExpiredSessions() {
	sm.sessionMutex.Lock()
	defer sm.sessionMutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
			expiredCount++

			// Remove from database
			if err := sm.removeSessionFromDB(id); err != nil {
				sm.logger.Printf("Warning: Failed to remove expired session from DB: %v", err)
			}
		}
	}

	if expiredCount > 0 {
		sm.logger.Printf("Cleaned up %d expired sessions", expiredCount)
	}
}

// Stop stops the session manager and cleans up resources
func (sm *ShipperSessionManager) Stop() {
	sm.logger.Printf("Stopping session manager")

	// Stop cleanup routine
	close(sm.stopCleanup)

	// Close all active sessions
	sm.sessionMutex.Lock()
	for id := range sm.sessions {
		if err := sm.transport.CloseSession(id); err != nil {
			sm.logger.Printf("Warning: Failed to close session %s: %v", id, err)
		}
	}
	sm.sessionMutex.Unlock()

	sm.logger.Printf("Session manager stopped")
}
