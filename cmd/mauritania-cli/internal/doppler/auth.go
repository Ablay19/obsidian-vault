package doppler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Authenticator handles Doppler authentication
type Authenticator struct {
	token string
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator() *Authenticator {
	return &Authenticator{}
}

// Login performs Doppler CLI login
func (a *Authenticator) Login(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "doppler", "login")
	cmd.Stdin = strings.NewReader("") // Non-interactive

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("doppler login failed: %w, output: %s", err, string(output))
	}

	return nil
}

// IsAuthenticated checks if user is logged in
func (a *Authenticator) IsAuthenticated(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "doppler", "me")
	return cmd.Run() == nil
}

// SetToken sets the authentication token
func (a *Authenticator) SetToken(token string) {
	a.token = token
}

// GetToken retrieves the current token
func (a *Authenticator) GetToken() string {
	if a.token != "" {
		return a.token
	}
	return os.Getenv("DOPPLER_TOKEN")
}
