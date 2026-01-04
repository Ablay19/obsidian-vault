package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"obsidian-automation/internal/config"
)

func TestAuthService_SessionManagement(t *testing.T) {
	// Setup
	cfg := config.Config{}
	cfg.Auth.SessionSecret = "test-secret-key-12345"
	cfg.Auth.GoogleClientID = "test-client-id"
	cfg.Auth.GoogleClientSecret = "test-client-secret"
	cfg.Auth.GoogleRedirectURL = "http://localhost:8080/callback"

	authService := NewAuthService(cfg)

	// Test 1: Create Session Cookie
	session := &UserSession{
		GoogleID: "12345",
		Email:    "test@example.com",
		Name:     "Test User",
		Exp:      time.Now().Add(1 * time.Hour).Unix(),
	}

	cookie, err := authService.CreateSessionCookie(session)
	if err != nil {
		t.Fatalf("Failed to create session cookie: %v", err)
	}

	if cookie.Name != "session" {
		t.Errorf("Expected cookie name 'session', got %s", cookie.Name)
	}

	// Test 2: Verify Session
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	verifiedSession, err := authService.VerifySession(req)
	if err != nil {
		t.Fatalf("Failed to verify session: %v", err)
	}

	if verifiedSession.GoogleID != session.GoogleID {
		t.Errorf("Expected GoogleID %s, got %s", session.GoogleID, verifiedSession.GoogleID)
	}

	// Test 3: Tampered Cookie
	reqTampered, _ := http.NewRequest("GET", "/", nil)
	tamperedCookie := &http.Cookie{
		Name:  "session",
		Value: cookie.Value + "tampered",
	}
	reqTampered.AddCookie(tamperedCookie)

	_, err = authService.VerifySession(reqTampered)
	if err == nil {
		t.Error("Expected error for tampered cookie, got nil")
	}
}

func TestAuthService_Middleware(t *testing.T) {
	cfg := config.Config{}
	cfg.Auth.SessionSecret = "test-secret"
	authService := NewAuthService(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := authService.Middleware(nextHandler)

	// Test 1: Redirect when no session (ensure NOT in dev mode)
	originalMode := os.Getenv("ENVIRONMENT_MODE")
	os.Unsetenv("ENVIRONMENT_MODE")
	defer os.Setenv("ENVIRONMENT_MODE", originalMode)

	req, _ := http.NewRequest("GET", "/dashboard", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Expected redirect 302, got %d", rr.Code)
	}

	// Test 2: Dev Mode Bypass
	os.Setenv("ENVIRONMENT_MODE", "dev")

	reqDev, _ := http.NewRequest("GET", "/dashboard", nil)
	rrDev := httptest.NewRecorder()

	middleware.ServeHTTP(rrDev, reqDev)

	if rrDev.Code != http.StatusOK {
		t.Errorf("Expected status 200 in dev mode, got %d", rrDev.Code)
	}
}

func TestAuthService_GetLoginURL(t *testing.T) {
	cfg := config.Config{}
	cfg.Auth.GoogleClientID = "test-client-id"
	cfg.Auth.GoogleRedirectURL = "http://localhost:8080/callback"

	authService := NewAuthService(cfg)

	url := authService.GetLoginURL("test-state")
	if url == "" {
		t.Error("Expected login URL, got empty string")
	}
}
