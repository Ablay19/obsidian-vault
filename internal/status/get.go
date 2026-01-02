package status

import (
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/state" // Import the state package
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
func GetServicesStatus(aiService *ai.AIService, rcm *state.RuntimeConfigManager) []ServiceStatus {
	var statuses []ServiceStatus

	// 1. Bot Status
	botStatus := "up"
	botDetails := fmt.Sprintf("Uptime: %s, Last Activity: %s", time.Since(GetStartTime()).String(), GetLastActivity().Format(time.RFC3339))
	if IsPaused() {
		botStatus = "paused"
		botDetails = "Bot is paused. " + botDetails
	}
	statuses = append(statuses, ServiceStatus{Name: "Bot Core", Status: botStatus, Details: botDetails})

	// 2. AI Service Status
	if aiService != nil {
		activeProvider := aiService.GetActiveProviderName()
		providerCount := len(aiService.GetAvailableProviders())
		aiStatus := "up"
		aiDetails := fmt.Sprintf("Active Provider: %s, Available Providers: %d", activeProvider, providerCount)

		// Check RCM for global AI enabled status
		rcmConfig := rcm.GetConfig()
		if !rcmConfig.AIEnabled {
			aiStatus = "disabled"
			aiDetails = "AI globally disabled by dashboard. " + aiDetails
		}

		statuses = append(statuses, ServiceStatus{Name: "AI Service", Status: aiStatus, Details: aiDetails})
	} else {
		statuses = append(statuses, ServiceStatus{Name: "AI Service", Status: "down", Details: "Not initialized"})
	}

	// 3. Database Status (using rcm's db)
	dbStatus := "up"
	dbDetails := "Connection OK"
	if err := rcm.GetDB().Ping(); err != nil { // Access db from rcm
		dbStatus = "down"
		dbDetails = fmt.Sprintf("Connection failed: %v", err)
	}
	statuses = append(statuses, ServiceStatus{Name: "Database", Status: dbStatus, Details: dbDetails})

	// Add other services as needed
	return statuses
}