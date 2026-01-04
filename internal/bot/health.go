package bot

import (
	"fmt"
	"net/http"
	"obsidian-automation/internal/status" // Import the new status package
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// StartHealthServer registers the health and control endpoints on the provided router.
func StartHealthServer(router *gin.Engine) {
	router.GET("/pause", func(c *gin.Context) {
		status.SetPaused(true)
		c.String(http.StatusOK, "Bot is paused.")
	})

	router.GET("/resume", func(c *gin.Context) {
		status.SetPaused(false)
		c.String(http.StatusOK, "Bot is resumed.")
	})

	router.GET("/status", func(c *gin.Context) {
		botStatus := "running"
		if status.IsPaused() {
			botStatus = "paused"
		}

		uptime := time.Since(status.GetStartTime())
		pid := os.Getpid()
		goVersion := runtime.Version()

		c.Header("X-Bot-Status", botStatus)
		c.Header("X-Bot-PID", fmt.Sprintf("%d", pid))
		c.Header("X-Bot-Go-Version", goVersion)
		c.Header("X-Bot-Uptime", uptime.String())

		c.JSON(http.StatusOK, gin.H{
			"status":        botStatus,
			"last_activity": status.GetLastActivity(),
			"pid":           pid,
			"go_version":    goVersion,
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"uptime":        uptime.String(),
		})
	})

	router.GET("/info", func(c *gin.Context) {
		if aiService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service not available"})
			return
		}

		infos := aiService.GetProvidersInfo()
		c.JSON(http.StatusOK, infos)
	})
}
