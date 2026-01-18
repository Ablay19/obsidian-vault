package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerInitialization(t *testing.T) {
	// Test server creation
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	require.NotNil(t, s)

	// Test tool registration
	err := registerTools(s)
	require.NoError(t, err)
}

func TestToolHandlers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		toolName string
		handler  func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		args     map[string]interface{}
	}{
		{
			name:     "status tool",
			toolName: "status",
			handler:  handleStatusTool,
			args:     map[string]interface{}{},
		},
		{
			name:     "logs tool",
			toolName: "logs",
			handler:  handleLogsTool,
			args: map[string]interface{}{
				"level": "error",
				"hours": 24.0,
			},
		},
		{
			name:     "diagnostics tool",
			toolName: "diagnostics",
			handler:  handleDiagnosticsTool,
			args:     map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      tt.toolName,
					Arguments: tt.args,
				},
			}

			result, err := tt.handler(ctx, req)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Content)

			// Check content type
			content := result.Content[0]
			_, ok := content.(mcp.TextContent)
			assert.True(t, ok, "Content should be TextContent")
		})
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2) // 2 requests per hour

	// First two requests should be allowed
	assert.True(t, rl.Allow("test-session"))
	assert.True(t, rl.Allow("test-session"))

	// Third should be denied
	assert.False(t, rl.Allow("test-session"))
}

func TestSanitizer(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "password",
			input:    "password: secret123",
			expected: "password: [REDACTED]",
		},
		{
			name:     "api key",
			input:    "api_key=abc123",
			expected: "api_key: [REDACTED]",
		},
		{
			name:     "normal text",
			input:    "status: connected",
			expected: "status: connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionManager(t *testing.T) {
	sm := NewSessionManager(time.Hour)

	// Create session
	session := sm.CreateSession("test-id", "test-agent")
	require.NotNil(t, session)
	assert.Equal(t, "test-id", session.ID)

	// Get session
	retrieved, exists := sm.GetSession("test-id")
	assert.True(t, exists)
	assert.Equal(t, session, retrieved)

	// Check active sessions
	assert.Equal(t, 1, sm.GetActiveSessions())
}
