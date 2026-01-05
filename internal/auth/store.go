package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemorySessionStore implements SessionStore interface using in-memory storage
type InMemorySessionStore struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() SessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

// Store stores a session in memory
func (ims *InMemorySessionStore) Store(ctx context.Context, session *Session) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	ims.sessions[session.ID] = session
	return nil
}

// Get retrieves a session from memory
func (ims *InMemorySessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	session, exists := ims.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// Delete removes a session from memory
func (ims *InMemorySessionStore) Delete(ctx context.Context, sessionID string) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	delete(ims.sessions, sessionID)
	return nil
}

// Cleanup removes expired sessions from memory
func (ims *InMemorySessionStore) Cleanup(ctx context.Context, expiredBefore time.Time) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	for id, session := range ims.sessions {
		if session.ExpiresAt.Before(expiredBefore) {
			delete(ims.sessions, id)
		}
	}

	return nil
}

// Count returns the number of active sessions (helper method for testing)
func (ims *InMemorySessionStore) Count() int {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()

	return len(ims.sessions)
}

// Clear removes all sessions (helper method for testing)
func (ims *InMemorySessionStore) Clear() {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()

	ims.sessions = make(map[string]*Session)
}
