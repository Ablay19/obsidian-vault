package auth

import "time"

// Session represents an authenticated user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// SessionConfig holds session management configuration
type SessionConfig struct {
	Secret     string        `json:"secret"`
	Expiration time.Duration `json:"expiration"`
	Secure     bool          `json:"secure"`
	HTTPOnly   bool          `json:"http_only"`
	CookieName string        `json:"cookie_name"`
}
