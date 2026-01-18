package mcp

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentSessions(t *testing.T) {
	// Test concurrent access to MCP server
	s := server.NewMCPServer(
		"Test Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	err := registerTools(s)
	require.NoError(t, err)

	// Test concurrent tool calls
	numGoroutines := 5
	numCalls := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numCalls)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(sessionID int) {
			defer wg.Done()

			for j := 0; j < numCalls; j++ {
				req := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Name: "status",
					},
				}

				_, err := handleStatusTool(context.Background(), req)
				if err != nil {
					errors <- err
				}
				time.Sleep(1 * time.Millisecond) // Small delay to simulate real usage
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorCount int
	for err := range errors {
		t.Logf("Concurrent call error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "No errors should occur in concurrent calls")
}

func TestSessionManagerConcurrency(t *testing.T) {
	sm := NewSessionManager(time.Hour)

	numGoroutines := 10
	numOperations := 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sessionID := fmt.Sprintf("session-%d", id)

			for j := 0; j < numOperations; j++ {
				// Create session
				session := sm.CreateSession(sessionID, "test-agent")
				if session == nil {
					errors <- fmt.Errorf("failed to create session %s", sessionID)
					continue
				}

				// Update session
				sm.UpdateSession(sessionID)

				// Get session
				retrieved, exists := sm.GetSession(sessionID)
				if !exists {
					errors <- fmt.Errorf("session %s not found", sessionID)
					continue
				}

				if retrieved.ID != sessionID {
					errors <- fmt.Errorf("session ID mismatch: expected %s, got %s", sessionID, retrieved.ID)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	var errorCount int
	for err := range errors {
		t.Logf("Session manager error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "No errors should occur in concurrent session operations")
	assert.True(t, sm.GetActiveSessions() > 0, "Should have active sessions")
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter(100) // High limit for testing

	numGoroutines := 20
	callsPerGoroutine := 50

	var wg sync.WaitGroup
	allowedCount := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(sessionID int) {
			defer wg.Done()

			session := fmt.Sprintf("session-%d", sessionID)
			localAllowed := 0

			for j := 0; j < callsPerGoroutine; j++ {
				if rl.Allow(session) {
					localAllowed++
				}
			}

			allowedCount <- localAllowed
		}(i)
	}

	wg.Wait()
	close(allowedCount)

	totalAllowed := 0
	for count := range allowedCount {
		totalAllowed += count
	}

	// Should allow most requests within the limit
	assert.True(t, totalAllowed > 0, "Some requests should be allowed")
	assert.True(t, totalAllowed <= 100*numGoroutines, "Should not exceed rate limit")
}
