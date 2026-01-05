package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionService manages user sessions using JWT tokens
type SessionService struct {
	secret []byte
	store  SessionStore
	config SessionConfig
	logger Logger
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Secret     string        `json:"-"` // Hidden from JSON
	Expiration time.Duration `json:"expiration"`
	Secure     bool          `json:"secure"`
	HTTPOnly   bool          `json:"http_only"`
	SameSite   http.SameSite `json:"same_site"`
	CookieName string        `json:"cookie_name"`
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	Store(ctx context.Context, session *Session) error
	Get(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
	Cleanup(ctx context.Context, expiredBefore time.Time) error
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	SessionID string `json:"sid"`
	UserID    string `json:"uid"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	jwt.RegisteredClaims
}

// NewSessionService creates a new session service
func NewSessionService(config SessionConfig, store SessionStore, logger Logger) (*SessionService, error) {
	if len(config.Secret) < 32 {
		return nil, fmt.Errorf("session secret must be at least 32 characters long")
	}

	if config.CookieName == "" {
		config.CookieName = "session"
	}

	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour
	}

	return &SessionService{
		secret: []byte(config.Secret),
		store:  store,
		config: config,
		logger: logger,
	}, nil
}

// CreateSession creates a new session for the user
func (ss *SessionService) CreateSession(ctx context.Context, userID, email, name string) (*Session, error) {
	sessionID := generateSessionID()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		Name:      name,
		ExpiresAt: time.Now().Add(ss.config.Expiration),
		CreatedAt: time.Now(),
	}

	if err := ss.store.Store(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	ss.logger.Info("Session created",
		"session_id", sessionID,
		"user_id", userID,
		"email", email)

	return session, nil
}

// ValidateSession validates a session token and returns the session
func (ss *SessionService) ValidateSession(ctx context.Context, tokenString string) (*Session, error) {
	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ss.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	// Retrieve session from store
	session, err := ss.store.Get(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	// Additional validation
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	if session.ID != claims.SessionID {
		return nil, fmt.Errorf("session ID mismatch")
	}

	return session, nil
}

// CreateSessionCookie creates a cookie containing the JWT session token
func (ss *SessionService) CreateSessionCookie(session *Session) (*http.Cookie, error) {
	token, err := ss.createJWT(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	return &http.Cookie{
		Name:     ss.config.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: ss.config.HTTPOnly,
		Secure:   ss.config.Secure,
		SameSite: ss.config.SameSite,
		Expires:  session.ExpiresAt,
	}, nil
}

// ExtractSessionFromRequest extracts session from HTTP request
func (ss *SessionService) ExtractSessionFromRequest(ctx context.Context, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(ss.config.CookieName)
	if err != nil {
		return nil, fmt.Errorf("session cookie not found: %w", err)
	}

	return ss.ValidateSession(ctx, cookie.Value)
}

// RevokeSession revokes a session
func (ss *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	if err := ss.store.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	ss.logger.Info("Session revoked", "session_id", sessionID)
	return nil
}

// CleanupExpiredSessions removes expired sessions from the store
func (ss *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	cutoff := time.Now()

	err := ss.store.Cleanup(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	ss.logger.Info("Expired sessions cleaned up", "cutoff_time", cutoff)
	return nil
}

// createJWT creates a JWT token for the session
func (ss *SessionService) createJWT(session *Session) (string, error) {
	claims := &JWTClaims{
		SessionID: session.ID,
		UserID:    session.UserID,
		Email:     session.Email,
		Name:      session.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(session.CreatedAt),
			NotBefore: jwt.NewNumericDate(session.CreatedAt),
			Issuer:    "obsidian-automation",
			Subject:   session.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ss.secret)
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random session ID: %v", err))
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
