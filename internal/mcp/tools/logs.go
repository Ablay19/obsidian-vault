package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleLogsTool retrieves recent log entries with filtering
func handleLogsTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get parameters
	level := "error"
	hours := 24.0

	if args := req.Params.Arguments; args != nil {
		if argsMap, ok := args.(map[string]interface{}); ok {
			if l, ok := argsMap["level"].(string); ok {
				level = l
			}
			if h, ok := argsMap["hours"].(float64); ok {
				hours = h
			}
		}
	}

	// Calculate time range
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	// In a real implementation, this would query actual logs
	// For now, return mock log entries
	mockLogs := []map[string]interface{}{
		{
			"timestamp": "2026-01-18T14:30:00Z",
			"level":     "error",
			"message":   "WhatsApp connection timeout after 30 seconds",
			"component": "whatsapp-transport",
		},
		{
			"timestamp": "2026-01-18T14:45:00Z",
			"level":     "error",
			"message":   "Telegram API rate limit exceeded (1000 requests/hour)",
			"component": "telegram-transport",
		},
		{
			"timestamp": "2026-01-18T15:00:00Z",
			"level":     "warn",
			"message":   "Configuration validation failed for shipper transport",
			"component": "config-validator",
		},
		{
			"timestamp": "2026-01-18T15:15:00Z",
			"level":     "info",
			"message":   "MCP server started successfully on port 8080",
			"component": "mcp-server",
		},
	}

	// Filter logs
	var filteredLogs []map[string]interface{}
	for _, log := range mockLogs {
		if logLevel, ok := log["level"].(string); ok {
			if shouldIncludeLog(logLevel, level) {
				if logTime, err := time.Parse(time.RFC3339, log["timestamp"].(string)); err == nil && logTime.After(since) {
					filteredLogs = append(filteredLogs, log)
				}
			}
		}
	}

	// Format logs
	var logsText string
	if len(filteredLogs) == 0 {
		logsText = fmt.Sprintf("No %s logs found in the last %.0f hours", level, hours)
	} else {
		logsText = fmt.Sprintf("Found %d %s log entries in the last %.0f hours:\n\n", len(filteredLogs), level, hours)
		for _, log := range filteredLogs {
			logsText += fmt.Sprintf("[%s] %s: %s\n",
				log["timestamp"],
				log["level"],
				log["message"])
		}
	}

	return mcp.NewToolResultText(logsText), nil
}

// shouldIncludeLog determines if a log level should be included based on filter
func shouldIncludeLog(logLevel, filterLevel string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	logPriority, logExists := levels[logLevel]
	filterPriority, filterExists := levels[filterLevel]

	if !logExists || !filterExists {
		return false
	}

	return logPriority >= filterPriority
}
