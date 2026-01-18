package mcp

import (
	"sync"
	"time"
)

// Session represents an MCP client session
type Session struct {
	ID        string
	CreatedAt time.Time
	LastSeen  time.Time
	UserAgent string
}

// SessionManager manages MCP client sessions with enhanced concurrency support
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	timeout     time.Duration
	maxSessions int
}

// NewSessionManager creates a new session manager with concurrency optimizations
func NewSessionManager(timeout time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:    make(map[string]*Session),
		timeout:     timeout,
		maxSessions: 50, // Reasonable limit for concurrent AI sessions
	}
	go sm.cleanup()
	return sm
}

// CreateSession creates a new session with concurrency limits
func (sm *SessionManager) CreateSession(id, userAgent string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check session limit
	if len(sm.sessions) >= sm.maxSessions {
		// Remove oldest session if at limit
		var oldestID string
		var oldestTime time.Time
		first := true

		for sid, session := range sm.sessions {
			if first || session.LastSeen.Before(oldestTime) {
				oldestID = sid
				oldestTime = session.LastSeen
				first = false
			}
		}

		if oldestID != "" {
			delete(sm.sessions, oldestID)
		}
	}

	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		UserAgent: userAgent,
	}
	sm.sessions[id] = session
	return session
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	if exists {
		session.LastSeen = time.Now()
	}
	return session, exists
}

// UpdateSession updates session activity
func (sm *SessionManager) UpdateSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[id]; exists {
		session.LastSeen = time.Now()
	}
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if now.Sub(session.CreatedAt) > sm.timeout {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// GetActiveSessions returns count of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Cleanup removes expired sessions (for testing)
func (sm *SessionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	now := time.Now()
	for id, session := range sm.sessions {
		if now.Sub(session.CreatedAt) > sm.timeout {
			delete(sm.sessions, id)
		}
	}
}
