package utils

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// StandardizedErrorLogger provides consistent error logging across services
type StandardizedErrorLogger struct {
	slogLogger *slog.Logger
	service    string
}

// ErrorLogEntry represents a standardized error log entry
type ErrorLogEntry struct {
	Service    string                 `json:"service"`
	Operation  string                 `json:"operation"`
	UserID     string                 `json:"user_id,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	ErrorType  string                 `json:"error_type"`
	Message    string                 `json:"message"`
	Severity   string                 `json:"severity"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// NewStandardizedErrorLogger creates a new standardized error logger
func NewStandardizedErrorLogger(service string, slogLogger *slog.Logger) *StandardizedErrorLogger {
	return &StandardizedErrorLogger{
		slogLogger: slogLogger,
		service:    service,
	}
}

// LogError logs an error with standardized context
func (l *StandardizedErrorLogger) LogError(err error, operation string, severity string, metadata map[string]interface{}) {
	entry := l.createErrorEntry(err, operation, severity, metadata)

	if l.slogLogger == nil {
		return
	}

	// Log with slog (structured logging)
	args := []any{
		"service", entry.Service,
		"operation", entry.Operation,
		"error_type", entry.ErrorType,
		"severity", entry.Severity,
		"timestamp", entry.Timestamp,
	}

	if entry.UserID != "" {
		args = append(args, "user_id", entry.UserID)
	}
	if entry.Resource != "" {
		args = append(args, "resource", entry.Resource)
	}
	if entry.Duration > 0 {
		args = append(args, "duration", entry.Duration)
	}
	if entry.StackTrace != "" {
		args = append(args, "stack_trace", entry.StackTrace)
	}

	// Add metadata
	for k, v := range entry.Metadata {
		args = append(args, k, v)
	}

	switch severity {
	case "critical":
		l.slogLogger.Error(entry.Message, args...)
	case "error":
		l.slogLogger.Error(entry.Message, args...)
	case "warning":
		l.slogLogger.Warn(entry.Message, args...)
	case "info":
		l.slogLogger.Info(entry.Message, args...)
	default:
		l.slogLogger.Info(entry.Message, args...)
	}
}

// LogValidationError logs validation errors
func (l *StandardizedErrorLogger) LogValidationError(operation, field, reason string, userID string) {
	l.LogError(
		&ValidationError{Field: field, Reason: reason},
		operation,
		"warning",
		map[string]interface{}{
			"user_id": userID,
			"field":   field,
			"reason":  reason,
		},
	)
}

// LogNetworkError logs network-related errors
func (l *StandardizedErrorLogger) LogNetworkError(err error, operation, endpoint string, statusCode int) {
	metadata := map[string]interface{}{
		"endpoint":    endpoint,
		"status_code": statusCode,
	}

	l.LogError(err, operation, "error", metadata)
}

// LogDatabaseError logs database errors
func (l *StandardizedErrorLogger) LogDatabaseError(err error, operation, table string, query string) {
	metadata := map[string]interface{}{
		"table": table,
		"query": query,
	}

	l.LogError(err, operation, "error", metadata)
}

// LogAuthenticationError logs authentication failures
func (l *StandardizedErrorLogger) LogAuthenticationError(operation, reason string, userID string, ip string) {
	metadata := map[string]interface{}{
		"reason":  reason,
		"user_id": userID,
		"ip":      ip,
	}

	l.LogError(
		&AuthenticationError{Reason: reason},
		operation,
		"warning",
		metadata,
	)
}

// LogAuthorizationError logs authorization failures
func (l *StandardizedErrorLogger) LogAuthorizationError(operation, resource, action string, userID string) {
	metadata := map[string]interface{}{
		"resource": resource,
		"action":   action,
		"user_id":  userID,
	}

	l.LogError(
		&AuthorizationError{Resource: resource, Action: action},
		operation,
		"warning",
		metadata,
	)
}

// LogRateLimitError logs rate limiting events
func (l *StandardizedErrorLogger) LogRateLimitError(operation, limitType string, userID string, resetTime int64) {
	metadata := map[string]interface{}{
		"limit_type": limitType,
		"user_id":    userID,
		"reset_time": resetTime,
	}

	l.LogError(
		&RateLimitError{LimitType: limitType, ResetTime: resetTime},
		operation,
		"warning",
		metadata,
	)
}

// LogServiceUnavailableError logs service unavailable errors
func (l *StandardizedErrorLogger) LogServiceUnavailableError(operation, serviceName string, retryAfter int) {
	metadata := map[string]interface{}{
		"service_name": serviceName,
		"retry_after":  retryAfter,
	}

	l.LogError(
		&ServiceUnavailableError{ServiceName: serviceName, RetryAfter: retryAfter},
		operation,
		"error",
		metadata,
	)
}

func (l *StandardizedErrorLogger) createErrorEntry(err error, operation, severity string, metadata map[string]interface{}) *ErrorLogEntry {
	entry := &ErrorLogEntry{
		Service:   l.service,
		Operation: operation,
		Severity:  severity,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	if err != nil {
		entry.Message = err.Error()

		// Extract context from error if available
		if ctx := GetErrorContext(err); ctx != nil {
			if ctx.UserID != "" {
				entry.UserID = ctx.UserID
			}
			if ctx.Resource != "" {
				entry.Resource = ctx.Resource
			}
			if ctx.Metadata != nil {
				for k, v := range ctx.Metadata {
					entry.Metadata[k] = v
				}
			}
		}

		// Categorize error type
		entry.ErrorType = l.categorizeError(err)
	}

	return entry
}

func (l *StandardizedErrorLogger) categorizeError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// Network errors
	if containsAny(errStr, "connection", "timeout", "network", "dial", "tcp") {
		return "network"
	}

	// Database errors
	if containsAny(errStr, "database", "sql", "query", "transaction") {
		return "database"
	}

	// Validation errors
	if containsAny(errStr, "validation", "invalid", "required", "format") {
		return "validation"
	}

	// Authentication errors
	if containsAny(errStr, "auth", "login", "password", "token", "unauthorized") {
		return "authentication"
	}

	// Authorization errors
	if containsAny(errStr, "permission", "forbidden", "access", "role") {
		return "authorization"
	}

	// External service errors
	if containsAny(errStr, "external", "third-party", "api", "service") {
		return "external_service"
	}

	// Rate limiting errors
	if containsAny(errStr, "rate limit", "throttle", "quota") {
		return "rate_limit"
	}

	return "application"
}

func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(strings.ToLower(s), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// Error types for different categories
type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Reason)
}

type AuthenticationError struct {
	Reason string
}

func (e AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Reason)
}

type AuthorizationError struct {
	Resource string
	Action   string
}

func (e AuthorizationError) Error() string {
	return fmt.Sprintf("authorization failed: %s access to %s", e.Action, e.Resource)
}

type RateLimitError struct {
	LimitType string
	ResetTime int64
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s", e.LimitType)
}

type ServiceUnavailableError struct {
	ServiceName string
	RetryAfter  int
}

func (e ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service %s unavailable, retry after %d seconds", e.ServiceName, e.RetryAfter)
}
