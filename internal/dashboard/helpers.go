package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/telemetry" // Add telemetry import
	"regexp"
	"time"
)

func getSessionUser(ctx context.Context) *auth.UserSession {
	val := ctx.Value("session")
	if val == nil {
		return nil
	}
	return val.(*auth.UserSession)
}

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

func isTelegramLinked(email string) bool { // Added context
	if email == "" {
		return false
	}
	var telegramID sql.NullInt64
	err := database.Client.DB.QueryRow("SELECT telegram_id FROM users WHERE email = ?", email).Scan(&telegramID)
	if err != nil {
		telemetry.ZapLogger.Sugar().Debugw("isTelegramLinked check failed", "email", email, "error", err)
		return false
	}
	telemetry.ZapLogger.Sugar().Debugw("isTelegramLinked check successful", "email", email, "valid", telegramID.Valid, "id", telegramID.Int64)
	return telegramID.Valid
}
