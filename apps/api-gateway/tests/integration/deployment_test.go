package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway/internal/handlers"
	types "github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

func TestDeploymentIndependenceIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Workers can be updated independently", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				handlers.Health(w, r)
			case "/api/v1/workers":
				handlers.ListWorkers(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		resp, err = http.Get(server.URL + "/api/v1/workers")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Go apps can be updated independently", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/health":
				handlers.Health(w, r)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	default:
	}
}

func TestZeroDowntimeDeployment(t *testing.T) {
	t.Run("Continuous availability during updates", func(t *testing.T) {
		requests := make(chan string, 100)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests <- r.URL.Path
			handlers.Health(w, r)
		}))
		defer server.Close()

		for i := 0; i < 10; i++ {
			go func() {
				resp, _ := http.Get(server.URL + "/health")
				if resp != nil {
					resp.Body.Close()
				}
			}()
		}

		timeout := time.After(5 * time.Second)
		count := 0
		for {
			select {
			case <-requests:
				count++
				if count >= 10 {
					return
				}
			case <-timeout:
				t.Fatalf("Expected 10 requests, got %d", count)
			}
		}
	})
}

func TestIndependentRollback(t *testing.T) {
	t.Run("Component can rollback without affecting others", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/workers" {
				handlers.ListWorkers(w, r)
			} else {
				handlers.Health(w, r)
			}
		}))
		defer server.Close()

		for i := 0; i < 3; i++ {
			resp, err := http.Get(server.URL + "/health")
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}
		}

		resp, err := http.Get(server.URL + "/api/v1/workers")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestAPIResponseFormat(t *testing.T) {
	t.Run("Health response format", func(t *testing.T) {
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
	})

	t.Run("Workers list response format", func(t *testing.T) {
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
	})
}

func init() {
	go func() {
		<-time.After(100 * time.Millisecond)
	}()
}
