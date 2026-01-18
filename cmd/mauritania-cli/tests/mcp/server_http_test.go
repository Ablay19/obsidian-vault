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

func TestMCPHTTPTransport(t *testing.T) {
	// Test server initialization with HTTP transport
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

func TestHTTPTransportStartup(t *testing.T) {
	// This test would verify HTTP server startup
	// For now, just test the transport setup
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	err := registerTools(s)
	require.NoError(t, err)

	// Test HTTP server creation (without actually starting)
	httpServer := server.NewStreamableHTTPServer(s)
	require.NotNil(t, httpServer)
}

func TestRateLimiterWithSessions(t *testing.T) {
	rl := NewRateLimiter(5) // 5 requests per hour

	// Simulate multiple sessions
	sessions := []string{"session1", "session2", "session3"}

	for i := 0; i < 3; i++ {
		for _, session := range sessions {
			allowed := rl.Allow(session)
			if i < 2 {
				assert.True(t, allowed, "Request %d for %s should be allowed", i+1, session)
			}
		}
	}

	// Fourth request should be denied for each session
	for _, session := range sessions {
		assert.False(t, rl.Allow(session), "Fourth request for %s should be denied", session)
	}
}

func TestSessionManagerCleanup(t *testing.T) {
	sm := NewSessionManager(100 * time.Millisecond) // Very short timeout

	// Create session
	session := sm.CreateSession("test", "agent")
	require.NotNil(t, session)

	// Verify session exists
	retrieved, exists := sm.GetSession("test")
	assert.True(t, exists)
	assert.Equal(t, session, retrieved)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Session should be cleaned up
	_, exists = sm.GetSession("test")
	assert.False(t, exists, "Session should be cleaned up after timeout")
}
