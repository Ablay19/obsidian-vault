package services

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// TransportSelector selects the optimal transport for command execution
type TransportSelector struct {
	db             *database.DB
	config         *utils.Config
	logger         *log.Logger
	transportStats map[string]*TransportStats
}

// TransportStats tracks performance statistics for each transport
type TransportStats struct {
	Platform            string
	TotalRequests       int
	SuccessCount        int
	FailureCount        int
	AverageLatency      time.Duration
	TotalCost           float64
	LastUsed            time.Time
	ConsecutiveFailures int
	IsHealthy           bool
}

// SelectionCriteria defines criteria for transport selection
type SelectionCriteria struct {
	Priority         string // "cost", "speed", "reliability", "availability"
	MaxCost          float64
	MaxLatency       time.Duration
	RequiredPlatform string   // Force specific platform if needed
	ExcludePlatforms []string // Platforms to avoid
}

// NewTransportSelector creates a new transport selector
func NewTransportSelector(db *database.DB, config *utils.Config, logger *log.Logger) *TransportSelector {
	ts := &TransportSelector{
		db:             db,
		config:         config,
		logger:         logger,
		transportStats: make(map[string]*TransportStats),
	}

	// Initialize stats for known platforms
	ts.initializeTransportStats()

	return ts
}

// initializeTransportStats sets up initial statistics for known transports
func (ts *TransportSelector) initializeTransportStats() {
	platforms := []string{"whatsapp", "telegram", "facebook"}

	for _, platform := range platforms {
		ts.transportStats[platform] = &TransportStats{
			Platform:            platform,
			TotalRequests:       0,
			SuccessCount:        0,
			FailureCount:        0,
			AverageLatency:      5 * time.Second, // Default assumption
			TotalCost:           0.0,
			LastUsed:            time.Now().Add(-24 * time.Hour), // Not used recently
			ConsecutiveFailures: 0,
			IsHealthy:           true,
		}
	}
}

// SelectTransport selects the optimal transport based on criteria
func (ts *TransportSelector) SelectTransport(criteria SelectionCriteria) (string, error) {
	ts.logger.Printf("Selecting transport with criteria: %+v", criteria)

	// Get available transports
	availableTransports := ts.getAvailableTransports()

	if len(availableTransports) == 0 {
		return "", fmt.Errorf("no transports available")
	}

	// Filter by criteria
	candidates := ts.filterTransports(availableTransports, criteria)

	if len(candidates) == 0 {
		return "", fmt.Errorf("no transports meet the specified criteria")
	}

	// Score and rank candidates
	scoredCandidates := ts.scoreTransports(candidates, criteria)

	// Select the best candidate
	if len(scoredCandidates) == 0 {
		return "", fmt.Errorf("no suitable transports found")
	}

	selected := scoredCandidates[0].Platform
	ts.logger.Printf("Selected transport: %s (score: %.2f)", selected, scoredCandidates[0].Score)

	return selected, nil
}

// getAvailableTransports returns currently available transport platforms
func (ts *TransportSelector) getAvailableTransports() []string {
	available := []string{}

	// Check which transports are configured
	if ts.config.Transports.SocialMedia.WhatsApp.APIKey != "" {
		available = append(available, "whatsapp")
	}
	if ts.config.Transports.SocialMedia.Telegram.BotToken != "" {
		available = append(available, "telegram")
	}
	if ts.config.Transports.SocialMedia.Facebook.AccessToken != "" {
		available = append(available, "facebook")
	}

	ts.logger.Printf("Available transports: %v", available)
	return available
}

// filterTransports filters transports based on criteria
func (ts *TransportSelector) filterTransports(transports []string, criteria SelectionCriteria) []string {
	filtered := []string{}

	for _, transport := range transports {
		// Check required platform
		if criteria.RequiredPlatform != "" && transport != criteria.RequiredPlatform {
			continue
		}

		// Check excluded platforms
		excluded := false
		for _, excludedPlatform := range criteria.ExcludePlatforms {
			if transport == excludedPlatform {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check if transport is healthy
		if stats, exists := ts.transportStats[transport]; exists {
			if !stats.IsHealthy && stats.ConsecutiveFailures > 2 {
				ts.logger.Printf("Skipping unhealthy transport: %s (%d consecutive failures)", transport, stats.ConsecutiveFailures)
				continue
			}
		}

		filtered = append(filtered, transport)
	}

	return filtered
}

// TransportScore represents a scored transport option
type TransportScore struct {
	Platform string
	Score    float64
	Reason   string
}

// scoreTransports scores and ranks transport candidates
func (ts *TransportSelector) scoreTransports(transports []string, criteria SelectionCriteria) []TransportScore {
	scores := []TransportScore{}

	for _, transport := range transports {
		score := ts.calculateTransportScore(transport, criteria)
		scores = append(scores, score)
	}

	// Sort by score (higher is better)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
}

// calculateTransportScore calculates a score for a transport based on criteria
func (ts *TransportSelector) calculateTransportScore(transport string, criteria SelectionCriteria) TransportScore {
	stats := ts.transportStats[transport]
	score := 0.0
	reasons := []string{}

	// Base score from success rate
	if stats.TotalRequests > 0 {
		successRate := float64(stats.SuccessCount) / float64(stats.TotalRequests)
		score += successRate * 30 // 0-30 points for reliability
		reasons = append(reasons, fmt.Sprintf("success rate: %.1f%%", successRate*100))
	} else {
		score += 15 // Neutral score for new transports
		reasons = append(reasons, "new transport")
	}

	// Cost factor
	cost := ts.getTransportCost(transport)
	if criteria.MaxCost > 0 && cost > criteria.MaxCost {
		score -= 50 // Heavy penalty for exceeding cost limit
		reasons = append(reasons, fmt.Sprintf("cost $%.4f exceeds limit $%.4f", cost, criteria.MaxCost))
	} else {
		// Lower cost is better
		costScore := math.Max(0, 20-cost) // 0-20 points, lower cost = higher score
		score += costScore
		reasons = append(reasons, fmt.Sprintf("cost: $%.4f", cost))
	}

	// Latency factor
	latency := stats.AverageLatency
	if criteria.MaxLatency > 0 && latency > criteria.MaxLatency {
		score -= 30 // Penalty for exceeding latency limit
		reasons = append(reasons, fmt.Sprintf("latency %v exceeds limit %v", latency, criteria.MaxLatency))
	} else {
		// Lower latency is better
		latencyScore := math.Max(0, 20-latency.Seconds()) // 0-20 points
		score += latencyScore
		reasons = append(reasons, fmt.Sprintf("latency: %v", latency))
	}

	// Recency bonus (recently used transports might be more reliable)
	timeSinceLastUse := time.Since(stats.LastUsed)
	if timeSinceLastUse < time.Hour {
		score += 10 // Bonus for recently used
		reasons = append(reasons, "recently used")
	}

	// Health bonus
	if stats.IsHealthy {
		score += 5
		reasons = append(reasons, "healthy")
	} else {
		score -= 10 // Penalty for unhealthy
		reasons = append(reasons, "unhealthy")
	}

	return TransportScore{
		Platform: transport,
		Score:    score,
		Reason:   fmt.Sprintf("%v", reasons),
	}
}

// getTransportCost returns the cost per request for a transport
func (ts *TransportSelector) getTransportCost(transport string) float64 {
	// Cost estimates (in USD per request)
	costs := map[string]float64{
		"whatsapp": 0.005, // ~$0.005 per message
		"telegram": 0.0,   // Free for bots
		"facebook": 0.001, // ~$0.001 per message
	}

	if cost, exists := costs[transport]; exists {
		return cost
	}
	return 0.01 // Default cost
}

// UpdateTransportStats updates statistics for a transport after use
func (ts *TransportSelector) UpdateTransportStats(platform string, success bool, latency time.Duration, cost float64) {
	stats, exists := ts.transportStats[platform]
	if !exists {
		ts.initializeTransportStats()
		stats = ts.transportStats[platform]
	}

	stats.TotalRequests++
	stats.LastUsed = time.Now()

	if success {
		stats.SuccessCount++
		stats.ConsecutiveFailures = 0
		stats.IsHealthy = true
	} else {
		stats.FailureCount++
		stats.ConsecutiveFailures++
		if stats.ConsecutiveFailures > 3 {
			stats.IsHealthy = false
		}
	}

	// Update average latency
	if stats.TotalRequests == 1 {
		stats.AverageLatency = latency
	} else {
		// Exponential moving average
		alpha := 0.1
		stats.AverageLatency = time.Duration(float64(stats.AverageLatency)*(1-alpha) + float64(latency)*alpha)
	}

	stats.TotalCost += cost

	ts.logger.Printf("Updated stats for %s: success=%v, latency=%v, total_requests=%d",
		platform, success, latency, stats.TotalRequests)
}

// GetTransportStats returns statistics for all transports
func (ts *TransportSelector) GetTransportStats() map[string]*TransportStats {
	// Return a copy to prevent external modification
	stats := make(map[string]*TransportStats)
	for k, v := range ts.transportStats {
		stats[k] = v
	}
	return stats
}

// OptimizeRoute finds the optimal route for a command based on current conditions
func (ts *TransportSelector) OptimizeRoute(command string, senderID string) (string, error) {
	// Analyze command characteristics
	criteria := SelectionCriteria{
		Priority: "cost", // Default to cost optimization
	}

	// For urgent commands, prioritize speed
	if strings.Contains(strings.ToLower(command), "urgent") {
		criteria.Priority = "speed"
	}

	// For large outputs, prefer reliable transports
	if strings.Contains(command, "cat") || strings.Contains(command, "tail") || strings.Contains(command, "log") {
		criteria.Priority = "reliability"
	}

	return ts.SelectTransport(criteria)
}

// GetRecommendedTransport returns the recommended transport for general use
func (ts *TransportSelector) GetRecommendedTransport() string {
	criteria := SelectionCriteria{
		Priority: "cost",
	}

	selected, err := ts.SelectTransport(criteria)
	if err != nil {
		ts.logger.Printf("Failed to select recommended transport: %v", err)
		// Return first available as fallback
		available := ts.getAvailableTransports()
		if len(available) > 0 {
			return available[0]
		}
		return "whatsapp" // Ultimate fallback
	}

	return selected
}

// ResetTransportStats resets statistics for a transport (useful for testing)
func (ts *TransportSelector) ResetTransportStats(platform string) {
	if stats, exists := ts.transportStats[platform]; exists {
		stats.TotalRequests = 0
		stats.SuccessCount = 0
		stats.FailureCount = 0
		stats.ConsecutiveFailures = 0
		stats.IsHealthy = true
		stats.TotalCost = 0.0
		stats.LastUsed = time.Now().Add(-24 * time.Hour)
	}
}

// IsTransportAvailable checks if a specific transport is available
func (ts *TransportSelector) IsTransportAvailable(platform string) bool {
	available := ts.getAvailableTransports()
	for _, p := range available {
		if p == platform {
			return true
		}
	}
	return false
}
