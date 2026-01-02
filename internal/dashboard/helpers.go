package dashboard

import (
	"fmt"
	"obsidian-automation/internal/status"
	"regexp"
	"time"
)

func getBotStatus(services []status.ServiceStatus) string {
	for _, s := range services {
		if s.Name == "Bot Core" {
			return s.Status
		}
	}
	return "Unknown"
}

func getUptime(services []status.ServiceStatus) string {
	startTime := status.GetStartTime()
	duration := time.Since(startTime)
	return fmt.Sprintf("%s", duration.Round(time.Second))
}

func getLastActivity(services []status.ServiceStatus) string {
	lastActivity := status.GetLastActivity()
	return fmt.Sprintf("%s ago", time.Since(lastActivity).Round(time.Second))
}

func getPID(services []status.ServiceStatus) string {
	for _, s := range services {
		if s.Name == "Bot Core" {
			// Example detail string: "Uptime: 1h2m3s, Last Activity: ..., PID: 12345"
			re := regexp.MustCompile(`PID: (\d+)`)
			match := re.FindStringSubmatch(s.Details)
			if len(match) > 1 {
				return match[1]
			}
		}
	}
	return "N/A"
}