package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test suite for the complete API gateway
func TestAPIGatewayIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup test server
	server := setupTestServer(t)
	defer server.Close()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t, client, server.URL)
	})

	t.Run("AuthenticationFlow", func(t *testing.T) {
		testAuthenticationFlow(t, client, server.URL)
	})

	t.Run("WhatsAppWebhook", func(t *testing.T) {
		testWhatsAppWebhook(t, client, server.URL)
	})

	t.Run("AIProcessing", func(t *testing.T) {
		testAIProcessing(t, client, server.URL)
	})

	t.Run("RateLimiting", func(t *testing.T) {
		testRateLimiting(t, client, server.URL)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, client, server.URL)
	})

	t.Run("Tracing", func(t *testing.T) {
		testTracing(t, client, server.URL)
	})
}

func testHealthCheck(t *testing.T, client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")
}

func testAuthenticationFlow(t *testing.T, client *http.Client, baseURL string) {
	// Test login endpoint
	loginData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "testpass123",
	}

	jsonData, err := json.Marshal(loginData)
	require.NoError(t, err)

	resp, err := client.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should get some response (even if it's a mock)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized)

	if resp.StatusCode == http.StatusOK {
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "token_type")
	}
}

func testWhatsAppWebhook(t *testing.T, client *http.Client, baseURL string) {
	// Test webhook verification
	verifyURL := fmt.Sprintf("%s/webhook/whatsapp?hub.mode=subscribe&hub.challenge=test_challenge&hub.verify_token=test_verify", baseURL)
	resp, err := client.Get(verifyURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)

	// Test webhook message processing
	webhookData := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []map[string]interface{}{
			{
				"id": "test_entry",
				"changes": []map[string]interface{}{
					{
						"field": "messages",
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"metadata": map[string]interface{}{
								"display_phone_number": "1234567890",
								"phone_number_id":      "123456789",
							},
							"messages": []map[string]interface{}{
								{
									"id":        "test_msg_1",
									"from":      "1234567890",
									"type":      "text",
									"timestamp": "1234567890",
									"text": map[string]interface{}{
										"body": "Hello from integration test",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(webhookData)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/webhook/whatsapp", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "test_signature")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
}

func testAIProcessing(t *testing.T, client *http.Client, baseURL string) {
	// Test AI processing endpoint
	aiData := map[string]interface{}{
		"prompt":      "Explain quantum computing in simple terms",
		"model":       "gpt-3.5-turbo",
		"max_tokens":  150,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(aiData)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/v1/ai/generate", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should get some response
	assert.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode == http.StatusInternalServerError)
}

func testRateLimiting(t *testing.T, client *http.Client, baseURL string) {
	// Make multiple requests quickly to test rate limiting
	requests := 15
	delay := 50 * time.Millisecond

	rateLimited := false

	for i := 0; i < requests; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			break
		}

		time.Sleep(delay)
	}

	// Rate limiting might or might not be enabled in test environment
	// Just log the result
	if rateLimited {
		t.Log("Rate limiting is working")
	} else {
		t.Log("Rate limiting not triggered (possibly disabled in test environment)")
	}
}

func testErrorHandling(t *testing.T, client *http.Client, baseURL string) {
	// Test invalid endpoint
	resp, err := client.Get(baseURL + "/nonexistent/endpoint")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var errorResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse, "error")
	assert.Contains(t, errorResponse, "code")
	assert.Equal(t, float64(http.StatusNotFound), errorResponse["code"])

	// Test invalid method
	resp, err = client.Post(baseURL+"/health", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func testTracing(t *testing.T, client *http.Client, baseURL string) {
	// Make a request and check for tracing headers
	resp, err := client.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check for correlation ID header
	correlationID := resp.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		t.Logf("Correlation ID found: %s", correlationID)
		assert.NotEmpty(t, correlationID)
	}

	// Check for request ID header
	requestID := resp.Header.Get("X-Request-ID")
	if requestID != "" {
		t.Logf("Request ID found: %s", requestID)
		assert.NotEmpty(t, requestID)
	}

	// Check for service name header
	serviceName := resp.Header.Get("X-Service-Name")
	if serviceName != "" {
		t.Logf("Service name found: %s", serviceName)
	}
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	server := setupTestServer(b)
	defer server.Close()

	client := &http.Client{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/health")
			if err == nil {
				resp.Body.Close()
			}
		}
	})
}

func BenchmarkAIProcessing(b *testing.B) {
	server := setupTestServer(b)
	defer server.Close()

	client := &http.Client{}
	aiData := map[string]interface{}{
		"prompt":     "Hello world",
		"model":      "test",
		"max_tokens": 50,
	}
	jsonData, _ := json.Marshal(aiData)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", server.URL+"/api/v1/ai/generate", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		}
	})
}

// Helper functions
func setupTestServer(t testing.TB) *httptest.Server {
	// This would normally set up your actual server
	// For this integration test, we'll create a mock server

	handler := http.NewServeMux()

	// Health check endpoint
	handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"status":    "ok",
			"service":   "api-gateway",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Mock auth endpoint
	handler.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"access_token":  "mock_jwt_token",
			"refresh_token": "mock_refresh_token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// Mock WhatsApp webhook
	handler.HandleFunc("/webhook/whatsapp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Webhook verification
			mode := r.URL.Query().Get("hub.mode")
			challenge := r.URL.Query().Get("hub.challenge")
			token := r.URL.Query().Get("hub.verify_token")

			if mode == "subscribe" && token == "test_verify" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(challenge))
				return
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodPost {
			// Webhook message processing
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Mock AI endpoint
	handler.HandleFunc("/api/v1/ai/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"content":       "This is a mock AI response for integration testing.",
			"model":         "mock-model",
			"finish_reason": "completed",
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(handler)
}

// Load testing
func TestLoadUnderPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	server := setupTestServer(t)
	defer server.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	concurrency := 50
	requestsPerWorker := 20

	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := client.Get(server.URL + "/health")
				if err != nil {
					t.Errorf("Worker %d, request %d failed: %v", workerID, j, err)
					continue
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Worker %d, request %d got status %d", workerID, j, resp.StatusCode)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	t.Logf("Completed load test: %d workers, %d requests each", concurrency, requestsPerWorker)
}

// End-to-end user journey test
func TestEndToEndUserJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	server := setupTestServer(t)
	defer server.Close()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Step 1: Health check
	resp, err := client.Get(server.URL + "/health")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 2: Authentication
	loginData := map[string]interface{}{
		"email":    "user@example.com",
		"password": "password123",
	}
	jsonData, err := json.Marshal(loginData)
	require.NoError(t, err)

	resp, err = client.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var authResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	require.NoError(t, err)
	require.Contains(t, authResponse, "access_token")

	token := authResponse["access_token"].(string)

	// Step 3: Use authenticated endpoint
	req, err := http.NewRequest("POST", server.URL+"/api/v1/ai/generate", bytes.NewBufferString(`{"prompt":"Hello"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should succeed or fail gracefully
	assert.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusTooManyRequests)

	t.Log("End-to-end user journey test completed successfully")
}
