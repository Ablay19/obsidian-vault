package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var isPaused atomic.Value
var lastActivity atomic.Value
var startTime time.Time

func init() {
	isPaused.Store(false)
	startTime = time.Now()
	lastActivity.Store(startTime)
}

func UpdateActivity() {
	lastActivity.Store(time.Now())
}

func StartHealthServer() {
	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		isPaused.Store(true)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is paused."))
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		isPaused.Store(false)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is resumed."))
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := "running"
		if isPaused.Load().(bool) {
			status = "paused"
		}

		uptime := time.Since(startTime)
		pid := os.Getpid()
		goVersion := runtime.Version()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Bot-Status", status)
		w.Header().Set("X-Bot-PID", fmt.Sprintf("%d", pid))
		w.Header().Set("X-Bot-Go-Version", goVersion)
		w.Header().Set("X-Bot-Uptime", uptime.String())

		data := map[string]interface{}{
			"status":        status,
			"last_activity": lastActivity.Load(),
			"pid":           pid,
			"go_version":    goVersion,
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"uptime":        uptime.String(),
		}
		json.NewEncoder(w).Encode(data)
	})

	go http.ListenAndServe(":8080", nil)
}