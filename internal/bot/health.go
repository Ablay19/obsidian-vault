package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"obsidian-automation/internal/status" // Import the new status package
	"os"
	"runtime"
	"time"
)

// StartHealthServer registers the health and control endpoints on the provided router.
func StartHealthServer(router *http.ServeMux) {
	router.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		status.SetPaused(true)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is paused."))
	})

	router.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		status.SetPaused(false)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is resumed."))
	})

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		botStatus := "running"
		if status.IsPaused() {
			botStatus = "paused"
		}

		uptime := time.Since(status.GetStartTime())
		pid := os.Getpid()
		goVersion := runtime.Version()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Bot-Status", botStatus)
		w.Header().Set("X-Bot-PID", fmt.Sprintf("%d", pid))
		w.Header().Set("X-Bot-Go-Version", goVersion)
		w.Header().Set("X-Bot-Uptime", uptime.String())

		data := map[string]interface{}{
			"status":        botStatus,
			"last_activity": status.GetLastActivity(),
			"pid":           pid,
			"go_version":    goVersion,
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"uptime":        uptime.String(),
		}
		json.NewEncoder(w).Encode(data)
	})

	router.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if aiService == nil {
			http.Error(w, "AI service not available", http.StatusInternalServerError)
			return
		}

		infos := aiService.GetProvidersInfo()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(infos)
	})
}