package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TracingConfig holds configuration for request tracing
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	HeaderName     string
	GenerateID     bool
}

// TracingMiddleware provides request tracing with correlation IDs
type TracingMiddleware struct {
	config TracingConfig
}

// TraceInfo holds tracing information for a request
type TraceInfo struct {
	CorrelationID  string    `json:"correlation_id"`
	RequestID      string    `json:"request_id"`
	ServiceName    string    `json:"service_name"`
	ServiceVersion string    `json:"service_version"`
	StartTime      time.Time `json:"start_time"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	UserAgent      string    `json:"user_agent,omitempty"`
	RemoteAddr     string    `json:"remote_addr,omitempty"`
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(config TracingConfig) *TracingMiddleware {
	if config.HeaderName == "" {
		config.HeaderName = "X-Correlation-ID"
	}
	if config.ServiceName == "" {
		config.ServiceName = "unknown-service"
	}

	return &TracingMiddleware{
		config: config,
	}
}

// Middleware returns the tracing middleware handler
func (tm *TracingMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate or extract correlation ID
			correlationID := r.Header.Get(tm.config.HeaderName)
			if correlationID == "" && tm.config.GenerateID {
				correlationID = tm.generateCorrelationID()
			}

			// Generate request ID
			requestID := tm.generateRequestID()

			// Create trace info
			traceInfo := &TraceInfo{
				CorrelationID:  correlationID,
				RequestID:      requestID,
				ServiceName:    tm.config.ServiceName,
				ServiceVersion: tm.config.ServiceVersion,
				StartTime:      time.Now(),
				Method:         r.Method,
				Path:           r.URL.Path,
				UserAgent:      r.Header.Get("User-Agent"),
				RemoteAddr:     tm.getClientIP(r),
			}

			// Add headers to response
			if correlationID != "" {
				w.Header().Set(tm.config.HeaderName, correlationID)
			}
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Service-Name", tm.config.ServiceName)
			if tm.config.ServiceVersion != "" {
				w.Header().Set("X-Service-Version", tm.config.ServiceVersion)
			}

			// Create context with trace info
			ctx := context.WithValue(r.Context(), traceInfoKey{}, traceInfo)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetTraceInfo extracts trace information from request context
func GetTraceInfo(ctx context.Context) *TraceInfo {
	if traceInfo, ok := ctx.Value(traceInfoKey{}).(*TraceInfo); ok {
		return traceInfo
	}
	return nil
}

// SetTraceInfo sets trace information in context
func SetTraceInfo(ctx context.Context, traceInfo *TraceInfo) context.Context {
	return context.WithValue(ctx, traceInfoKey{}, traceInfo)
}

// generateCorrelationID generates a unique correlation ID
func (tm *TracingMiddleware) generateCorrelationID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return tm.generateRequestID()
	}
	return hex.EncodeToString(bytes)
}

// generateRequestID generates a unique request ID
func (tm *TracingMiddleware) generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// getClientIP extracts the client IP address from the request
func (tm *TracingMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// traceInfoKey is the context key for trace information
type traceInfoKey struct{}

// Tracing utilities for propagating trace context

// InjectTraceHeaders injects trace headers into outgoing HTTP requests
func InjectTraceHeaders(ctx context.Context, req *http.Request) {
	if traceInfo := GetTraceInfo(ctx); traceInfo != nil {
		if traceInfo.CorrelationID != "" {
			req.Header.Set("X-Correlation-ID", traceInfo.CorrelationID)
		}
		req.Header.Set("X-Request-ID", traceInfo.RequestID)
		req.Header.Set("X-Service-Name", traceInfo.ServiceName)
		if traceInfo.ServiceVersion != "" {
			req.Header.Set("X-Service-Version", traceInfo.ServiceVersion)
		}
	}
}

// ExtractTraceHeaders extracts trace headers from incoming HTTP requests
func ExtractTraceHeaders(r *http.Request) *TraceInfo {
	correlationID := r.Header.Get("X-Correlation-ID")
	requestID := r.Header.Get("X-Request-ID")
	serviceName := r.Header.Get("X-Service-Name")
	serviceVersion := r.Header.Get("X-Service-Version")

	if correlationID == "" && requestID == "" {
		return nil
	}

	return &TraceInfo{
		CorrelationID:  correlationID,
		RequestID:      requestID,
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		StartTime:      time.Now(),
		Method:         r.Method,
		Path:           r.URL.Path,
		UserAgent:      r.Header.Get("User-Agent"),
	}
}

// CreateChildTrace creates a child trace from a parent trace
func CreateChildTrace(parent *TraceInfo) *TraceInfo {
	if parent == nil {
		return nil
	}

	child := *parent
	child.RequestID = generateTraceRequestID()
	child.StartTime = time.Now()

	return &child
}

// generateTraceRequestID generates a unique request ID (helper function)
func generateTraceRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Tracing helpers for logging

// GetCorrelationID returns the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if traceInfo := GetTraceInfo(ctx); traceInfo != nil {
		return traceInfo.CorrelationID
	}
	return ""
}

// GetRequestID returns the request ID from context
func GetRequestID(ctx context.Context) string {
	if traceInfo := GetTraceInfo(ctx); traceInfo != nil {
		return traceInfo.RequestID
	}
	return ""
}

// LogTraceInfo adds trace information to log fields
func LogTraceInfo(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})

	if traceInfo := GetTraceInfo(ctx); traceInfo != nil {
		if traceInfo.CorrelationID != "" {
			fields["correlation_id"] = traceInfo.CorrelationID
		}
		if traceInfo.RequestID != "" {
			fields["request_id"] = traceInfo.RequestID
		}
		fields["service"] = traceInfo.ServiceName
		fields["method"] = traceInfo.Method
		fields["path"] = traceInfo.Path
	}

	return fields
}
