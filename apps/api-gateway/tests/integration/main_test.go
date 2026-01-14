package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway/internal/handlers"
	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}

	if response["service"] != "api-gateway" {
		t.Errorf("Expected service 'api-gateway', got '%v'", response["service"])
	}
}

func TestListWorkersEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	w := httptest.NewRecorder()

	handlers.ListWorkers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response types.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}
}

func TestColoredLogger(t *testing.T) {
	logger := types.NewColoredLogger("test-component")

	var capturedOutput string
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger.Info("Test message", "key", "value")

	w.Close()
	os.Stdout = originalStdout
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	capturedOutput = string(buf[:n])

	if capturedOutput == "" {
		t.Error("Expected colored output, got empty string")
	}

	if !containsAny(capturedOutput, "\x1b[38;5;") {
		t.Error("Expected ANSI color codes in output")
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) && contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDeploymentIndependence(t *testing.T) {
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.Health(w, r)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Errorf("Request failed: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	}()

	go func() {
		defer func() { done <- true }()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.ListWorkers(w, r)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/api/v1/workers")
		if err != nil {
			t.Errorf("Request failed: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	}()

	timeout := time.After(10 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}
}

func init() {
	fmt.Println("Running integration tests...")
}
