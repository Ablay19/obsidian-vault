package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CSRFConfig struct {
	Secret        string
	CookieName    string
	HeaderName    string
	FormFieldName string
	TokenLength   int
	MaxAge        time.Duration
	Secure        bool
	HTTPOnly      bool
	SameSite      http.SameSite
	SkipPaths     []string
	Logger        Logger
}

type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

type CSRFStore interface {
	Store(ctx context.Context, token string, expiresAt time.Time) error
	Get(ctx context.Context, token string) (bool, error)
	Delete(ctx context.Context, token string) error
	Cleanup(ctx context.Context, before time.Time) error
}

type InMemoryCSRFStore struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
}

func NewInMemoryCSRFStore() *InMemoryCSRFStore {
	return &InMemoryCSRFStore{
		tokens: make(map[string]time.Time),
	}
}

func (s *InMemoryCSRFStore) Store(ctx context.Context, token string, expiresAt time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.tokens[token] = expiresAt
	return nil
}

func (s *InMemoryCSRFStore) Get(ctx context.Context, token string) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	expiresAt, exists := s.tokens[token]
	if !exists {
		return false, nil
	}
	if time.Now().After(expiresAt) {
		return false, nil
	}
	return true, nil
}

func (s *InMemoryCSRFStore) Delete(ctx context.Context, token string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.tokens, token)
	return nil
}

func (s *InMemoryCSRFStore) Cleanup(ctx context.Context, before time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for token, expiresAt := range s.tokens {
		if expiresAt.Before(before) {
			delete(s.tokens, token)
		}
	}
	return nil
}

type CSRFService struct {
	config CSRFConfig
	store  CSRFStore
	logger Logger
}

type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

func NewCSRFService(config CSRFConfig, store CSRFStore, logger Logger) (*CSRFService, error) {
	if config.Secret == "" {
		return nil, fmt.Errorf("CSRF secret is required")
	}
	if len(config.Secret) < 32 {
		return nil, fmt.Errorf("CSRF secret must be at least 32 characters")
	}
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "csrf_token"
	}
	if config.MaxAge == 0 {
		config.MaxAge = 1 * time.Hour
	}

	return &CSRFService{
		config: config,
		store:  store,
		logger: logger,
	}, nil
}

func (s *CSRFService) GenerateToken(ctx context.Context) (*CSRFToken, error) {
	tokenBytes := make([]byte, s.config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(s.config.MaxAge)

	if err := s.store.Store(ctx, token, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to store CSRF token: %w", err)
	}

	return &CSRFToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *CSRFService) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("CSRF token is empty")
	}

	valid, err := s.store.Get(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to validate CSRF token: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid or expired CSRF token")
	}

	return nil
}

func (s *CSRFService) CreateTokenCookie(token *CSRFToken) *http.Cookie {
	return &http.Cookie{
		Name:     s.config.CookieName,
		Value:    token.Token,
		Path:     "/",
		HttpOnly: s.config.HTTPOnly,
		Secure:   s.config.Secure,
		SameSite: s.config.SameSite,
		Expires:  token.ExpiresAt,
	}
}

func (s *CSRFService) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if s.shouldSkip(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			token, err := s.extractToken(r)
			if err != nil {
				s.logger.Warn("CSRF token extraction failed", "error", err.Error())
				http.Error(w, "CSRF token missing or invalid", http.StatusForbidden)
				return
			}

			if err := s.ValidateToken(r.Context(), token); err != nil {
				s.logger.Warn("CSRF token validation failed", "error", err.Error())
				http.Error(w, "CSRF token invalid or expired", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (s *CSRFService) shouldSkip(path string) bool {
	for _, skipPath := range s.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func (s *CSRFService) extractToken(r *http.Request) (string, error) {
	headerToken := r.Header.Get(s.config.HeaderName)
	if headerToken != "" {
		return headerToken, nil
	}

	formToken := r.FormValue(s.config.FormFieldName)
	if formToken != "" {
		return formToken, nil
	}

	cookie, err := r.Cookie(s.config.CookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no CSRF token found in header, form, or cookie")
}
