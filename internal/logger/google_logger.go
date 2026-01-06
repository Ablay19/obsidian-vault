package logger

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GoogleLogger provides structured logging
type GoogleLogger struct {
	projectID string
	logger    *zap.Logger
}

// NewGoogleLogger creates a new Google logger
func NewGoogleLogger(projectID string) *GoogleLogger {
	zapLogger, _ := zap.NewProduction()

	return &GoogleLogger{
		projectID: projectID,
		logger:    zapLogger,
	}
}

// LogRequest logs HTTP request with structured data
func (g *GoogleLogger) LogRequest(c *gin.Context, duration time.Duration) {
	logData := map[string]interface{}{
		"timestamp":   time.Now(),
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"query":       c.Request.URL.RawQuery,
		"user_agent":  c.Request.UserAgent(),
		"client_ip":   c.ClientIP(),
		"duration_ms": float64(duration.Milliseconds()),
		"status_code": c.Writer.Status(),
		"user_id":     getUserID(c),
		"project_id":  g.projectID,
	}

	g.logger.Info("HTTP Request", zap.Any("details", logData))
}

// LogUserAction logs user-specific actions
func (g *GoogleLogger) LogUserAction(action string, c *gin.Context) {
	logData := map[string]interface{}{
		"timestamp":  time.Now(),
		"action":     action,
		"user_id":    getUserID(c),
		"project_id": g.projectID,
	}

	g.logger.Info("User Action", zap.Any("details", logData))
}

// LogError logs application errors with context
func (g *GoogleLogger) LogError(err error, c *gin.Context) {
	logData := map[string]interface{}{
		"timestamp":   time.Now(),
		"error":       err.Error(),
		"stack_trace": getStackTrace(),
		"user_id":     getUserID(c),
		"project_id":  g.projectID,
	}

	g.logger.Error("Application Error", zap.Any("details", logData))
}

// LogSystemEvent logs system events
func (g *GoogleLogger) LogSystemEvent(event string, level string, details map[string]interface{}) {
	logData := map[string]interface{}{
		"timestamp":  time.Now(),
		"event":      event,
		"level":      level,
		"details":    details,
		"project_id": g.projectID,
	}

	severity := "INFO"
	if level == "ERROR" {
		severity = "ERROR"
	} else if level == "WARN" {
		severity = "WARN"
	}

	g.logger.Info("System Event", zap.String("severity", severity), zap.Any("details", logData))
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
