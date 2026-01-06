package middleware

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"obsidian-automation/internal/logger"
)

// GoogleCloudLoggingMiddleware integrates Google Cloud logging
func GoogleCloudLoggingMiddleware() gin.HandlerFunc {
	// Check if Google Cloud logging is enabled
	googleCloudEnabled := os.Getenv("ENABLE_GOOGLE_LOGGING") == "true"
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	var googleLogger *logger.GoogleLogger
	if googleCloudEnabled && projectID != "" {
		googleLogger = logger.NewGoogleLogger(projectID)
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log to Google Cloud if enabled
		if googleLogger != nil {
			duration := time.Since(startTime)
			googleLogger.LogRequest(c, duration)
		}
	}
}

// UserActionLoggingMiddleware logs user actions to Google Cloud
func UserActionLoggingMiddleware() gin.HandlerFunc {
	googleCloudEnabled := os.Getenv("ENABLE_GOOGLE_LOGGING") == "true"
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	var googleLogger *logger.GoogleLogger
	if googleCloudEnabled && projectID != "" {
		googleLogger = logger.NewGoogleLogger(projectID)
	}

	return func(c *gin.Context) {
		// Log specific user actions
		if googleLogger != nil {
			action := c.Request.Method + " " + c.Request.URL.Path
			googleLogger.LogUserAction(action, c)
		}
		c.Next()
	}
}
