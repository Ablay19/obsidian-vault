package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

// ErrorMetricsCollector collects and tracks error metrics
type ErrorMetricsCollector struct {
	metrics map[string]*ErrorMetrics
	mutex   sync.RWMutex
}

// ErrorMetrics holds metrics for a specific error type or service
type ErrorMetrics struct {
	Service         string
	ErrorType       string
	TotalCount      int64
	ErrorRate       float64
	LastErrorTime   time.Time
	FirstErrorTime  time.Time
	TimeWindow      time.Duration
	ErrorCounts     map[string]int64 // error_code -> count
	RecoveryCount   int64
	AverageDuration time.Duration
	mutex           sync.RWMutex
}

// NewErrorMetricsCollector creates a new error metrics collector
func NewErrorMetricsCollector() *ErrorMetricsCollector {
	return &ErrorMetricsCollector{
		metrics: make(map[string]*ErrorMetrics),
	}
}

// RecordError records an error occurrence
func (emc *ErrorMetricsCollector) RecordError(service, operation, errorType, errorCode string, duration time.Duration) {
	key := service + ":" + errorType

	emc.mutex.Lock()
	metrics, exists := emc.metrics[key]
	if !exists {
		metrics = &ErrorMetrics{
			Service:        service,
			ErrorType:      errorType,
			TimeWindow:     5 * time.Minute, // 5 minute rolling window
			ErrorCounts:    make(map[string]int64),
			FirstErrorTime: time.Now(),
		}
		emc.metrics[key] = metrics
	}
	emc.mutex.Unlock()

	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	atomic.AddInt64(&metrics.TotalCount, 1)
	metrics.LastErrorTime = time.Now()
	if metrics.FirstErrorTime.IsZero() {
		metrics.FirstErrorTime = time.Now()
	}

	// Update error code counts
	if errorCode != "" {
		metrics.ErrorCounts[errorCode]++
	}

	// Update average duration
	if metrics.AverageDuration == 0 {
		metrics.AverageDuration = duration
	} else {
		// Simple moving average
		metrics.AverageDuration = (metrics.AverageDuration + duration) / 2
	}

	// Calculate error rate (errors per minute in the time window)
	timeSinceFirst := time.Since(metrics.FirstErrorTime)
	if timeSinceFirst > 0 {
		minutes := timeSinceFirst.Minutes()
		metrics.ErrorRate = float64(metrics.TotalCount) / minutes
	}
}

// RecordRecovery records a successful recovery
func (emc *ErrorMetricsCollector) RecordRecovery(service, errorType string) {
	key := service + ":" + errorType

	emc.mutex.RLock()
	metrics, exists := emc.metrics[key]
	emc.mutex.RUnlock()

	if exists {
		atomic.AddInt64(&metrics.RecoveryCount, 1)
	}
}

// GetMetrics returns metrics for a specific service and error type
func (emc *ErrorMetricsCollector) GetMetrics(service, errorType string) *ErrorMetrics {
	key := service + ":" + errorType

	emc.mutex.RLock()
	metrics := emc.metrics[key]
	emc.mutex.RUnlock()

	return metrics
}

// GetAllMetrics returns all collected metrics
func (emc *ErrorMetricsCollector) GetAllMetrics() map[string]*ErrorMetrics {
	emc.mutex.RLock()
	defer emc.mutex.RUnlock()

	// Create a copy to avoid concurrent map writes
	result := make(map[string]*ErrorMetrics)
	for k, v := range emc.metrics {
		result[k] = v
	}
	return result
}

// GetServiceMetrics returns all metrics for a specific service
func (emc *ErrorMetricsCollector) GetServiceMetrics(service string) map[string]*ErrorMetrics {
	emc.mutex.RLock()
	defer emc.mutex.RUnlock()

	result := make(map[string]*ErrorMetrics)
	for k, v := range emc.metrics {
		if v.Service == service {
			result[k] = v
		}
	}
	return result
}

// GetHighErrorRateServices returns services with error rates above threshold
func (emc *ErrorMetricsCollector) GetHighErrorRateServices(threshold float64) []string {
	emc.mutex.RLock()
	defer emc.mutex.RUnlock()

	var highErrorServices []string
	serviceErrors := make(map[string]float64)

	for _, metrics := range emc.metrics {
		if metrics.ErrorRate > threshold {
			serviceErrors[metrics.Service] += metrics.ErrorRate
		}
	}

	for service, rate := range serviceErrors {
		if rate > threshold {
			highErrorServices = append(highErrorServices, service)
		}
	}

	return highErrorServices
}

// GetFailingServices returns services that have been failing for too long
func (emc *ErrorMetricsCollector) GetFailingServices(maxAge time.Duration) []string {
	emc.mutex.RLock()
	defer emc.mutex.RUnlock()

	var failingServices []string
	serviceLastErrors := make(map[string]time.Time)

	for _, metrics := range emc.metrics {
		if time.Since(metrics.LastErrorTime) < maxAge {
			if lastError, exists := serviceLastErrors[metrics.Service]; !exists || metrics.LastErrorTime.After(lastError) {
				serviceLastErrors[metrics.Service] = metrics.LastErrorTime
			}
		}
	}

	for service, lastError := range serviceLastErrors {
		if time.Since(lastError) < maxAge {
			failingServices = append(failingServices, service)
		}
	}

	return failingServices
}

// ResetMetrics resets all metrics (useful for testing or periodic cleanup)
func (emc *ErrorMetricsCollector) ResetMetrics() {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.metrics = make(map[string]*ErrorMetrics)
}

// CleanupOldMetrics removes metrics older than the specified duration
func (emc *ErrorMetricsCollector) CleanupOldMetrics(maxAge time.Duration) {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	for key, metrics := range emc.metrics {
		if time.Since(metrics.LastErrorTime) > maxAge {
			delete(emc.metrics, key)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to periodically clean up old metrics
func (emc *ErrorMetricsCollector) StartCleanupRoutine(interval, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			emc.CleanupOldMetrics(maxAge)
		}
	}()
}

// ErrorMetricsSnapshot provides a snapshot of current error metrics
type ErrorMetricsSnapshot struct {
	Service         string           `json:"service"`
	ErrorType       string           `json:"error_type"`
	TotalCount      int64            `json:"total_count"`
	ErrorRate       float64          `json:"error_rate"`
	LastErrorTime   time.Time        `json:"last_error_time"`
	RecoveryCount   int64            `json:"recovery_count"`
	AverageDuration time.Duration    `json:"average_duration"`
	ErrorCounts     map[string]int64 `json:"error_counts"`
	TimeWindow      time.Duration    `json:"time_window"`
}

// Snapshot returns a snapshot of current metrics
func (em *ErrorMetrics) Snapshot() *ErrorMetricsSnapshot {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	// Create a copy of error counts
	errorCounts := make(map[string]int64)
	for k, v := range em.ErrorCounts {
		errorCounts[k] = v
	}

	return &ErrorMetricsSnapshot{
		Service:         em.Service,
		ErrorType:       em.ErrorType,
		TotalCount:      atomic.LoadInt64(&em.TotalCount),
		ErrorRate:       em.ErrorRate,
		LastErrorTime:   em.LastErrorTime,
		RecoveryCount:   atomic.LoadInt64(&em.RecoveryCount),
		AverageDuration: em.AverageDuration,
		ErrorCounts:     errorCounts,
		TimeWindow:      em.TimeWindow,
	}
}

// GetSnapshot returns a snapshot of all metrics
func (emc *ErrorMetricsCollector) GetSnapshot() map[string]*ErrorMetricsSnapshot {
	allMetrics := emc.GetAllMetrics()
	snapshot := make(map[string]*ErrorMetricsSnapshot)

	for key, metrics := range allMetrics {
		snapshot[key] = metrics.Snapshot()
	}

	return snapshot
}
