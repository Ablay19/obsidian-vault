//go:build ignore

package integration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Skip integration tests in CI unless explicitly requested
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		os.Exit(0)
	}

	// Run the tests
	os.Exit(m.Run())
}
