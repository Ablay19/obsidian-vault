package status

import (
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/state" // Import the state package
	"os"
	"runtime"
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
	TotalFiles       int
	ImageFiles       int
	PDFFiles         int
	AICalls          int
	LastActivity     time.Time
	SelectedProvider string
	EstimatedCost    float64
	ActualLatency    time.Duration
	AccuracyResult   float64
}

// GetStats returns the current statistics of the bot.
func GetStats(rcm *state.RuntimeConfigManager) Stats {
	stats := Stats{
		LastActivity: GetLastActivity(),
	}

	if rcm == nil || rcm.GetDB() == nil {
		return stats
	}

	db := rcm.GetDB()

	// Total Files
	_ = db.QueryRow("SELECT COUNT(*) FROM processed_files").Scan(&stats.TotalFiles)

	// AI Calls
	_ = db.QueryRow("SELECT COUNT(*) FROM chat_history WHERE direction = 'out'").Scan(&stats.AICalls)

	// The new fields will be populated by the calling code.
	return stats
}

// GetServicesStatus gathers and returns the status of all monitored services.
func GetServicesStatus(aiService ai.AIServiceInterface, rcm *state.RuntimeConfigManager) []ServiceStatus {
	var statuses []ServiceStatus

	// 1. Bot Status
	botStatus := "up"
	botDetails := fmt.Sprintf("Uptime: %s, Last Activity: %s, PID: %d, OS: %s/%s, Go: %s",
		time.Since(GetStartTime()).Round(time.Second).String(),
		GetLastActivity().Format("15:04:05"),
		os.Getpid(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
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
		rcmConfig := rcm.GetConfig(true)
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
