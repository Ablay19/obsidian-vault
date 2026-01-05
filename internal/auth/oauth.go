package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthService handles the OAuth2 flow with Google
type OAuthService struct {
	config     *oauth2.Config
	httpClient *http.Client
	logger     Logger
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// UserInfo represents user information from Google
type UserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

// TokenPair represents OAuth tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(config OAuthConfig, logger Logger) (*OAuthService, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &OAuthService{
		config: oauthConfig,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}, nil
}

// GetAuthorizationURL generates the authorization URL for Google OAuth
func (os *OAuthService) GetAuthorizationURL(state string) string {
	return os.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForTokens exchanges authorization code for tokens and user info
func (os *OAuthService) ExchangeCodeForTokens(ctx context.Context, code, state string) (*TokenPair, *UserInfo, error) {
	// Exchange authorization code for tokens
	token, err := os.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Get user info from Google
	userInfo, err := os.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// TODO: Implement ID token verification when OIDC library is available
	if idToken := token.Extra("id_token"); idToken != nil {
		os.logger.Warn("ID token verification not implemented", "token_present", true)
		// Continue even if ID token verification is not implemented
	}

	tokenPair := &TokenPair{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
	}

	return tokenPair, userInfo, nil
}

// RefreshTokens refreshes access token using refresh token
func (os *OAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour), // Force refresh
	}

	tokenSource := os.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	tokenPair := &TokenPair{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		Expiry:       newToken.Expiry,
		TokenType:    newToken.TokenType,
	}

	return tokenPair, nil
}

// getUserInfo retrieves user information from Google's API
func (os *OAuthService) getUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := os.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// verifyIDToken verifies the ID token and checks subject claim
// TODO: Implement when OIDC library is available
func (os *OAuthService) verifyIDToken(ctx context.Context, idToken, expectedSubject string) error {
	os.logger.Warn("ID token verification not implemented")
	return nil
}

// CreateAuthenticatedClient creates an HTTP client with OAuth credentials
func (os *OAuthService) CreateAuthenticatedClient(ctx context.Context, token *TokenPair) *http.Client {
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
	}

	return os.config.Client(ctx, oauthToken)
}
