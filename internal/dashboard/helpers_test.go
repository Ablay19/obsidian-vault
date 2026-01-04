package dashboard

import (
	"context"
	"testing"

	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/status"
)

func TestGetBotStatus(t *testing.T) {
	tests := []struct {
		name     string
		services []status.ServiceStatus
		want     string
	}{
		{
			name: "Bot Core found",
			services: []status.ServiceStatus{
				{Name: "Database", Status: "up"},
				{Name: "Bot Core", Status: "paused"},
			},
			want: "paused",
		},
		{
			name: "Bot Core not found",
			services: []status.ServiceStatus{
				{Name: "Database", Status: "up"},
			},
			want: "Unknown",
		},
		{
			name:     "Empty services",
			services: []status.ServiceStatus{},
			want:     "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBotStatus(tt.services); got != tt.want {
				t.Errorf("getBotStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPID(t *testing.T) {
	tests := []struct {
		name     string
		services []status.ServiceStatus
		want     string
	}{
		{
			name: "PID found",
			services: []status.ServiceStatus{
				{Name: "Bot Core", Details: "Uptime: 1h, PID: 12345, OS: linux"},
			},
			want: "12345",
		},
		{
			name: "PID not found in details",
			services: []status.ServiceStatus{
				{Name: "Bot Core", Details: "Uptime: 1h, OS: linux"},
			},
			want: "N/A",
		},
		{
			name: "Bot Core not found",
			services: []status.ServiceStatus{
				{Name: "Database", Details: "Connection OK"},
			},
			want: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPID(tt.services); got != tt.want {
				t.Errorf("getPID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSessionUser(t *testing.T) {
	user := &auth.UserSession{
		GoogleID: "123",
		Email:    "test@example.com",
		Name:     "Test User",
	}

	ctxWithUser := context.WithValue(context.Background(), "session", user)
	ctxNoUser := context.Background()

	t.Run("User present", func(t *testing.T) {
		got := getSessionUser(ctxWithUser)
		if got == nil {
			t.Fatal("expected user, got nil")
		}
		if got.Email != user.Email {
			t.Errorf("expected email %s, got %s", user.Email, got.Email)
		}
	})

	t.Run("User missing", func(t *testing.T) {
		got := getSessionUser(ctxNoUser)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}
