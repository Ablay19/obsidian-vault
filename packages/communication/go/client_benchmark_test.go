package communication

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkHttpClient_Do benchmarks the HTTP client Do method
func BenchmarkHttpClient_Do(b *testing.B) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "data": {"test": "data"}, "message": "success"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.DiscardHandler) // Discard logger for benchmarks
	client := NewHttpClient(server.URL, logger)

	config := RequestConfig{
		Method:   "GET",
		Endpoint: "/",
		Timeout:  5 * time.Second,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Do(config)
		if err != nil {
			b.Fatalf("HTTP request failed: %v", err)
		}
	}
}

// BenchmarkHttpClient_Get benchmarks the HTTP client Get method
func BenchmarkHttpClient_Get(b *testing.B) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "data": {"test": "data"}, "message": "success"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.DiscardHandler) // Discard logger for benchmarks
	client := NewHttpClient(server.URL, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Get("/", nil)
		if err != nil {
			b.Fatalf("HTTP GET failed: %v", err)
		}
	}
}

// BenchmarkHttpClient_Post benchmarks the HTTP client Post method
func BenchmarkHttpClient_Post(b *testing.B) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "data": {"created": true}, "message": "created"}`))
	}))
	defer server.Close()

	logger := slog.New(slog.DiscardHandler) // Discard logger for benchmarks
	client := NewHttpClient(server.URL, logger)

	body := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.Post("/", body, nil)
		if err != nil {
			b.Fatalf("HTTP POST failed: %v", err)
		}
	}
}
