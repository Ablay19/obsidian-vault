package status

import (
	"database/sql"
	"fmt"
	"obsidian-automation/internal/ai"
	"sync/atomic"
	"time"
)

var (
	isPaused     atomic.Value
	lastActivity atomic.Value
	startTime    time.Time
)

func init() {
	isPaused.Store(false)
	startTime = time.Now()
	lastActivity.Store(startTime)
}

// State management functions
func IsPaused() bool {
	return isPaused.Load().(bool)
}
func GetStartTime() time.Time {
	return startTime
}
func GetLastActivity() time.Time {
	return lastActivity.Load().(time.Time)
}
func UpdateActivity() {
	lastActivity.Store(time.Now())
}
func SetPaused(paused bool) {
	isPaused.Store(paused)
}


// ServiceStatus represents the status of a single service.
type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "up", "down", "paused"
	Details string `json:"details,omitempty"`
}

// Stats holds various statistics about the bot's operations.
type Stats struct {
	TotalFiles   int
	ImageFiles   int
	PDFFiles     int
	AICalls      int
	LastActivity time.Time
}

// GetStats returns the current statistics of the bot.
// This is a placeholder implementation and needs to be expanded
// to fetch actual data (e.g., from the database for file counts).
func GetStats() Stats {
	return Stats{
		TotalFiles:   0, // Placeholder
		ImageFiles:   0, // Placeholder
		PDFFiles:     0, // Placeholder
		AICalls:      0, // Placeholder
		LastActivity: GetLastActivity(),
	}
}

// GetServicesStatus gathers and returns the status of all monitored services.
func GetServicesStatus(aiService *ai.AIService, db *sql.DB) []ServiceStatus {
	var statuses []ServiceStatus

	// 1. Bot Status
	botStatus := "up"
	botDetails := fmt.Sprintf("Uptime: %s, Last Activity: %s", time.Since(GetStartTime()).String(), GetLastActivity().Format(time.RFC3339))
	if IsPaused() {
		botStatus = "paused"
		botDetails = "Bot is paused. " + botDetails
	}
	statuses = append(statuses, ServiceStatus{Name: "Bot Core", Status: botStatus, Details: botDetails})

	// ... (rest of the function is the same)
	return statuses
}