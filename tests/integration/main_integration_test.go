package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"obsidian-automation/tests/integration"
)

// TestMainIntegration runs all integration tests
func TestMainIntegration(t *testing.T) {
	// Skip integration tests in CI unless explicitly requested
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration tests - SKIP_INTEGRATION is set")
		return
	}

	suite.Run(t, "Integration Test Suite")
}
