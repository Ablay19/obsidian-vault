package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// NetworkStatus represents the current network connectivity state
type NetworkStatus struct {
	IsOnline     bool
	Latency      time.Duration
	LastChecked  time.Time
	Error        error
	Connectivity ConnectivityType
}

// ConnectivityType represents different types of network connectivity
type ConnectivityType string

const (
	ConnectivityOffline  ConnectivityType = "offline"
	ConnectivitySlow     ConnectivityType = "slow"
	ConnectivityMobile   ConnectivityType = "mobile"
	ConnectivityWiFi     ConnectivityType = "wifi"
	ConnectivityEthernet ConnectivityType = "ethernet"
	ConnectivityUnknown  ConnectivityType = "unknown"
)

// RetryConfig configures retry behavior for network operations
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Timeout       time.Duration
}

// DefaultRetryConfig returns sensible defaults for mobile environments
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		BaseDelay:     1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       10 * time.Second,
	}
}

// MobileRetryConfig returns optimized config for mobile networks
func MobileRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    5,
		BaseDelay:     2 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 1.5,
		Timeout:       15 * time.Second,
	}
}

// NetworkMonitor monitors network connectivity and provides resilience features
type NetworkMonitor struct {
	status         NetworkStatus
	mu             sync.RWMutex
	checkInterval  time.Duration
	testURLs       []string
	httpClient     *http.Client
	stopChan       chan struct{}
	statusCallback func(NetworkStatus)
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor() *NetworkMonitor {
	return &NetworkMonitor{
		status: NetworkStatus{
			IsOnline:     false,
			Connectivity: ConnectivityUnknown,
			LastChecked:  time.Now(),
		},
		checkInterval: 30 * time.Second,
		testURLs: []string{
			"https://www.google.com",
			"https://www.cloudflare.com",
			"https://1.1.1.1",
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Increased timeout for mobile
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins network monitoring in a background goroutine
func (nm *NetworkMonitor) Start() {
	go nm.monitorLoop()
}

// Stop stops network monitoring
func (nm *NetworkMonitor) Stop() {
	close(nm.stopChan)
}

// SetStatusCallback sets a callback for status changes
func (nm *NetworkMonitor) SetStatusCallback(callback func(NetworkStatus)) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.statusCallback = callback
}

// GetStatus returns the current network status
func (nm *NetworkMonitor) GetStatus() NetworkStatus {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.status
}

// CheckConnectivity performs an immediate connectivity check
func (nm *NetworkMonitor) CheckConnectivity() NetworkStatus {
	status := nm.performConnectivityCheck()

	nm.mu.Lock()
	oldStatus := nm.status
	nm.status = status
	nm.mu.Unlock()

	// Notify callback if status changed
	if nm.statusCallback != nil && (oldStatus.IsOnline != status.IsOnline || oldStatus.Connectivity != status.Connectivity) {
		go nm.statusCallback(status)
	}

	return status
}

// performConnectivityCheck does the actual connectivity testing
func (nm *NetworkMonitor) performConnectivityCheck() NetworkStatus {
	status := NetworkStatus{
		LastChecked: time.Now(),
	}

	// Try multiple connectivity tests for mobile environments
	dnsTests := []string{"google.com", "cloudflare.com", "1.1.1.1"}
	dnsSuccess := false

	for _, host := range dnsTests {
		if _, err := net.ResolveIPAddr("ip4", host); err == nil {
			dnsSuccess = true
			break
		}
	}

	// If DNS fails, try direct IP connectivity test
	if !dnsSuccess {
		// Test direct connection to Cloudflare DNS
		conn, err := net.DialTimeout("tcp", "1.1.1.1:53", 3*time.Second)
		if err != nil {
			status.IsOnline = false
			status.Error = fmt.Errorf("DNS and direct IP connectivity tests failed")
			status.Connectivity = ConnectivityOffline
			return status
		}
		conn.Close()
	}

	// Try HTTP connectivity to multiple endpoints
	var totalLatency time.Duration
	successCount := 0

	for _, url := range nm.testURLs {
		start := time.Now()
		resp, err := nm.httpClient.Get(url)
		latency := time.Since(start)

		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			totalLatency += latency
			successCount++
		}
	}

	if successCount == 0 {
		status.IsOnline = false
		status.Error = fmt.Errorf("all connectivity tests failed")
		status.Connectivity = ConnectivityOffline
		return status
	}

	status.IsOnline = true
	status.Latency = totalLatency / time.Duration(successCount)
	status.Connectivity = nm.detectConnectivityType(status.Latency)

	return status
}

// detectConnectivityType infers connection type from latency
func (nm *NetworkMonitor) detectConnectivityType(latency time.Duration) ConnectivityType {
	switch {
	case latency < 50*time.Millisecond:
		return ConnectivityEthernet
	case latency < 200*time.Millisecond:
		return ConnectivityWiFi
	case latency < 1000*time.Millisecond:
		return ConnectivityMobile
	default:
		return ConnectivitySlow
	}
}

// monitorLoop runs the continuous monitoring loop
func (nm *NetworkMonitor) monitorLoop() {
	ticker := time.NewTicker(nm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.CheckConnectivity()
		}
	}
}

// IsOnline returns true if network is currently online
func (nm *NetworkMonitor) IsOnline() bool {
	return nm.GetStatus().IsOnline
}

// WaitForConnectivity blocks until network connectivity is restored
func (nm *NetworkMonitor) WaitForConnectivity(timeout time.Duration) error {
	if nm.IsOnline() {
		return nil
	}

	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for network connectivity")
		case <-ticker.C:
			if nm.CheckConnectivity().IsOnline {
				return nil
			}
		}
	}
}

// RetryOperation retries a network operation with exponential backoff
func RetryOperation(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute operation
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(config.BaseDelay) * pow(config.BackoffFactor, attempt))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// pow calculates base^exponent for float64
func pow(base float64, exponent int) float64 {
	result := 1.0
	for i := 0; i < exponent; i++ {
		result *= base
	}
	return result
}

// IsRetryableError checks if an error is worth retrying
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network-related errors that are often transient
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"dns",
		"i/o timeout",
	}

	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if substring exists in string (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && toLower(s[:len(substr)]) == toLower(substr) ||
		(len(s) > len(substr) && containsIgnoreCase(s[1:], substr))
}

// toLower converts string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}
