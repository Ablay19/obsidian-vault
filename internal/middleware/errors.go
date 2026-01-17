package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ErrorHandler struct {
	logger           *zap.Logger
	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
	Code    int               `json:"code"`
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Time    time.Time         `json:"time"`
}

type HTTPError struct {
	Code    int
	Message string
	Details map[string]string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		notFound: func(w http.ResponseWriter, r *http.Request) {
			err := HTTPError{
				Code:    http.StatusNotFound,
				Message: "Resource not found",
			}

			response := ErrorResponse{
				Error:  err.Message,
				Code:   err.Code,
				Path:   r.URL.Path,
				Method: r.Method,
				Time:   time.Now(),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode(response)

			logger.Warn("404 Not Found",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
		},
		methodNotAllowed: func(w http.ResponseWriter, r *http.Request) {
			err := HTTPError{
				Code:    http.StatusMethodNotAllowed,
				Message: "Method not allowed",
			}

			response := ErrorResponse{
				Error:  err.Message,
				Code:   err.Code,
				Path:   r.URL.Path,
				Method: r.Method,
				Time:   time.Now(),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode(response)

			logger.Warn("405 Method Not Allowed",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
		},
	}
}

func (h *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr HTTPError
	if he, ok := err.(HTTPError); ok {
		httpErr = he
	} else {
		httpErr = HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Internal Server Error",
		}
	}

	response := ErrorResponse{
		Error:   httpErr.Message,
		Details: httpErr.Details,
		Code:    httpErr.Code,
		Path:    r.URL.Path,
		Method:  r.Method,
		Time:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Code)
	json.NewEncoder(w).Encode(response)

	h.logger.Error("HTTP error",
		zap.Int("status", httpErr.Code),
		zap.String("error", httpErr.Message),
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("stack", getStackTrace()),
	)
}

func (h *ErrorHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.notFound(w, r)
}

func (h *ErrorHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.methodNotAllowed(w, r)
}

func (h *ErrorHandler) RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err, ok := recovered.(error)
					if !ok {
						err = fmt.Errorf("%v", recovered)
					}
					h.logger.Error("Panic recovered",
						zap.Error(err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.String("stack", getStackTrace()),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func (h *ErrorHandler) RequestLoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)
			h.logger.Info("HTTP request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", duration),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}

type requestIDContextKey struct{}

func (h *ErrorHandler) RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return strings.TrimSpace(string(buf[:n]))
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond()%10000)
}
