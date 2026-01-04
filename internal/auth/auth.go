package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/database"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	oauthConfig *oauth2.Config
	secret      []byte
}

type UserSession struct {
	GoogleID string `json:"google_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Exp      int64  `json:"exp"`
}

func NewAuthService(cfg config.Config) *AuthService {
	zap.S().Info("Initializing Auth Service", "redirect_url", cfg.Auth.GoogleRedirectURL)
	return &AuthService{
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.Auth.GoogleClientID,
			ClientSecret: cfg.Auth.GoogleClientSecret,
			RedirectURL:  cfg.Auth.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/cloud-platform.read-only",
			},
			Endpoint: google.Endpoint,
		},
		secret: []byte(cfg.Auth.SessionSecret),
	}
}

func (s *AuthService) GetLoginURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state)
}

func (s *AuthService) HandleCallback(ctx context.Context, code string) (*UserSession, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var profile struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Update or create user in DB with tokens
	_, err = database.DB.Exec(`
		INSERT INTO users (id, first_name, google_id, email, profile_picture, access_token, refresh_token, token_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(google_id) DO UPDATE SET
			first_name = excluded.first_name,
			email = excluded.email,
			profile_picture = excluded.profile_picture,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_expiry = excluded.token_expiry
	`, profile.Sub[0:10], profile.Name, profile.Sub, profile.Email, profile.Picture, token.AccessToken, token.RefreshToken, token.Expiry)
	
	if err != nil {
		// Log error but continue
	}

	return &UserSession{
		GoogleID: profile.Sub,
		Email:    profile.Email,
		Name:     profile.Name,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (s *AuthService) CreateSessionCookie(session *UserSession) (*http.Cookie, error) {
	payload, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(payload)
	signature := s.sign(encoded)

	return &http.Cookie{
		Name:     "session",
		Value:    encoded + "." + signature,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(session.Exp, 0),
	}, nil
}

func (s *AuthService) VerifySession(r *http.Request) (*UserSession, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cookie format")
	}

	payload := parts[0]
	signature := parts[1]

	if !s.verify(payload, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var session UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if time.Now().Unix() > session.Exp {
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (s *AuthService) sign(data string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (s *AuthService) verify(data, signature string) bool {
	expected := s.sign(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *AuthService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow login routes
		if strings.HasPrefix(r.URL.Path, "/auth/") || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for dev bypass
		if os.Getenv("ENVIRONMENT_MODE") == "dev" {
			// If we are in dev mode and have no session, we'll allow a "dev-session"
			_, err := s.VerifySession(r)
			if err != nil {
				zap.S().Info("Dev mode detected: Bypassing real OAuth")
				devSession, _ := s.CreateDevSession()
				ctx := context.WithValue(r.Context(), "session", devSession)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		session, err := s.VerifySession(r)
		if err != nil {
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}

		// Add session to context if needed
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *AuthService) CreateDevSession() (*UserSession, error) {
	return &UserSession{
		GoogleID: "dev-user-id",
		Email:    "dev@example.com",
		Name:     "Development User",
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (s *AuthService) GetTokenForUser(ctx context.Context, googleID string) (*oauth2.Token, error) {
	var accessToken, refreshToken string
	var expiry time.Time

	err := database.DB.QueryRow(`
		SELECT access_token, refresh_token, token_expiry 
		FROM users 
		WHERE google_id = ?
	`, googleID).Scan(&accessToken, &refreshToken, &expiry)

	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}

	// Use TokenSource to handle auto-refresh
	ts := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}

	// If token was refreshed, update DB
	if newToken.AccessToken != accessToken {
		_, _ = database.DB.Exec(`
			UPDATE users 
			SET access_token = ?, refresh_token = ?, token_expiry = ? 
			WHERE google_id = ?
		`, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry, googleID)
	}

	return newToken, nil
}

func (s *AuthService) GetClientForUser(ctx context.Context, googleID string) (*http.Client, error) {
	token, err := s.GetTokenForUser(ctx, googleID)
	if err != nil {
		return nil, err
	}
	return s.oauthConfig.Client(ctx, token), nil
}

