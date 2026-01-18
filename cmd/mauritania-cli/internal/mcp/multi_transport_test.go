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

func TestMultiTransportIntegration(t *testing.T) {
	// Test that the server can handle different transport configurations
	s := server.NewMCPServer(
		"Multi-Transport Test Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	err := registerTools(s)
	require.NoError(t, err)

	// Test stdio transport creation (without actually starting)
	// This tests the transport setup logic
	assert.NotNil(t, s)

	// Test HTTP transport creation
	httpServer := server.NewStreamableHTTPServer(s)
	assert.NotNil(t, httpServer)
}

func TestSessionIsolation(t *testing.T) {
	sm := NewSessionManager(time.Hour)

	// Create multiple sessions
	session1 := sm.CreateSession("ai-1", "Claude")
	session2 := sm.CreateSession("ai-2", "ChatGPT")
	session3 := sm.CreateSession("ai-3", "Gemini")

	require.NotNil(t, session1)
	require.NotNil(t, session2)
	require.NotNil(t, session3)

	// Verify sessions are isolated
	retrieved1, exists1 := sm.GetSession("ai-1")
	retrieved2, exists2 := sm.GetSession("ai-2")
	retrieved3, exists3 := sm.GetSession("ai-3")

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.True(t, exists3)

	assert.Equal(t, "ai-1", retrieved1.ID)
	assert.Equal(t, "ai-2", retrieved2.ID)
	assert.Equal(t, "ai-3", retrieved3.ID)

	assert.Equal(t, "Claude", retrieved1.UserAgent)
	assert.Equal(t, "ChatGPT", retrieved2.UserAgent)
	assert.Equal(t, "Gemini", retrieved3.UserAgent)

	// Test session updates
	time.Sleep(10 * time.Millisecond)
	sm.UpdateSession("ai-1")
	retrieved1Updated, _ := sm.GetSession("ai-1")
	assert.True(t, retrieved1Updated.LastSeen.After(retrieved1.LastSeen))
}

func TestRateLimitingMultiSession(t *testing.T) {
	rl := NewRateLimiter(3) // 3 requests per hour per session

	sessions := []string{"claude", "chatgpt", "gemini", "copilot"}

	// Each session should be able to make requests independently
	for _, session := range sessions {
		assert.True(t, rl.Allow(session), "First request for %s should be allowed", session)
		assert.True(t, rl.Allow(session), "Second request for %s should be allowed", session)
	}

	// Sixth request for first session should be denied
	assert.False(t, rl.Allow("claude"), "Fourth request for claude should be denied")

	// But other sessions should still be allowed
	assert.True(t, rl.Allow("chatgpt"), "ChatGPT should still be allowed")
	assert.True(t, rl.Allow("gemini"), "Gemini should still be allowed")
}

func TestToolExecutionAcrossSessions(t *testing.T) {
	ctx := context.Background()

	// Simulate multiple AIs calling the same tool
	tools := []string{"status", "logs", "diagnostics"}

	for _, toolName := range tools {
		t.Run("tool_"+toolName, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: toolName,
				},
			}

			var result *mcp.CallToolResult
			var err error

			switch toolName {
			case "status":
				result, err = handleStatusTool(ctx, req)
			case "logs":
				result, err = handleLogsTool(ctx, req)
			case "diagnostics":
				result, err = handleDiagnosticsTool(ctx, req)
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Content)
		})
	}
}
