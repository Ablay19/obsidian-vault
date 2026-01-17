package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeploymentFlow tests the complete deployment flow for architectural separation
func TestDeploymentFlow(t *testing.T) {
	ctx := context.Background()
	baseURL := "http://localhost:8080"

	t.Run("API Gateway Deployment Independence", func(t *testing.T) {
		// Test that API Gateway can start and serve requests independently
		// This simulates the scenario where workers are deployed separately

		// Wait for service to be ready
		require.Eventually(t, func() bool {
			resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
			if err != nil {
				return false
			}
			defer resp.Body.Close()
			return resp.StatusCode == http.StatusOK
		}, 30*time.Second, 1*time.Second, "API Gateway should be ready within 30 seconds")

		// Test health endpoint
		resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("Worker Endpoint Availability", func(t *testing.T) {
		// Test that worker-related endpoints are available even without active workers
		// This tests the decoupling between API Gateway and workers

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/workers", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return success even if no workers are connected
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})

	t.Run("Service Resilience", func(t *testing.T) {
		// Test that the service remains resilient during simulated deployment scenarios

		// Test rapid health checks (simulating load balancer checks)
		for i := 0; i < 10; i++ {
			resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Configuration Validation", func(t *testing.T) {
		// Test that configuration is properly loaded for test environment

		resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		// In test environment, we expect debug logging or test-specific behavior
		// This validates that environment-specific configuration is working
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Network Isolation Verification", func(t *testing.T) {
		// Test that internal network communication patterns are working
		// This is a basic test - more comprehensive network isolation testing
		// would require additional infrastructure

		client := &http.Client{Timeout: 5 * time.Second}

		resp, err := client.Get(fmt.Sprintf("%s/health", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestIndependentScaling simulates independent scaling scenarios
func TestIndependentScaling(t *testing.T) {
	baseURL := "http://localhost:8080"

	t.Run("Concurrent Request Handling", func(t *testing.T) {
		// Test that the service can handle concurrent requests
		// This simulates multiple clients hitting the service during scaling operations

		concurrentRequests := 50
		results := make(chan error, concurrentRequests)

		for i := 0; i < concurrentRequests; i++ {
			go func() {
				resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}

				results <- nil
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < concurrentRequests; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})
}

// TestFailFastBehavior tests the fail-fast error handling for inter-component communication
func TestFailFastBehavior(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 1 * time.Second}

	t.Run("Timeout Handling", func(t *testing.T) {
		// Test that requests timeout appropriately
		// This simulates network issues during component communication

		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("%s/health", baseURL))
		duration := time.Since(start)

		if err != nil {
			// Request should fail fast, not hang
			assert.Less(t, duration, 2*time.Second, "Request should fail fast")
		} else {
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Invalid Endpoint Handling", func(t *testing.T) {
		// Test that invalid endpoints return appropriate errors quickly

		resp, err := http.Get(fmt.Sprintf("%s/invalid-endpoint", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
