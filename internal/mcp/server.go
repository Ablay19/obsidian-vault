package mcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StartServer initializes and starts the MCP server
func StartServer(transport, host, port string) error {
	// Initialize caching and monitoring
	// In a real implementation, this would set up actual caching

	// Create MCP server
	s := server.NewMCPServer(
		"Mauritania CLI Diagnostics",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	if err := registerTools(s); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Start server based on transport
	switch transport {
	case "stdio":
		return startStdioServer(s)
	case "http":
		return startHTTPServer(s, host, port)
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}
}

// registerTools adds all diagnostic tools to the server
func registerTools(s *server.MCPServer) error {
	// Status tool
	s.AddTool(mcp.NewTool("status",
		mcp.WithDescription("Check connectivity status of all transports"),
	), handleStatusTool)

	// Logs tool
	s.AddTool(mcp.NewTool("logs",
		mcp.WithDescription("Retrieve recent log entries with filtering"),
		mcp.WithString("level",
			mcp.Description("Log level filter (debug, info, warn, error)"),
			mcp.DefaultString("error"),
		),
		mcp.WithNumber("hours",
			mcp.Description("Hours back to search"),
			mcp.DefaultNumber(24),
		),
	), handleLogsTool)

	// Diagnostics tool
	s.AddTool(mcp.NewTool("diagnostics",
		mcp.WithDescription("Run comprehensive system diagnostics"),
	), handleDiagnosticsTool)

	// Config tool
	s.AddTool(mcp.NewTool("config",
		mcp.WithDescription("View sanitized configuration status"),
		mcp.WithString("transport",
			mcp.Description("Specific transport to check"),
		),
	), handleConfigTool)

	// Test connection tool
	s.AddTool(mcp.NewTool("test_connection",
		mcp.WithDescription("Test connectivity to a specific transport"),
		mcp.WithString("transport",
			mcp.Description("Transport to test (whatsapp, telegram, facebook, shipper)"),
			mcp.Enum("whatsapp", "telegram", "facebook", "shipper"),
		),
	), handleTestConnectionTool)

	// Error metrics tool
	s.AddTool(mcp.NewTool("error_metrics",
		mcp.WithDescription("Get error statistics and trends"),
		mcp.WithString("window",
			mcp.Description("Time window (hour, day, week)"),
			mcp.DefaultString("day"),
			mcp.Enum("hour", "day", "week"),
		),
	), handleErrorMetricsTool)

	return nil
}

// startStdioServer starts the server with stdio transport
func startStdioServer(s *server.MCPServer) error {
	return server.ServeStdio(s)
}

// startHTTPServer starts the server with HTTP transport
func startHTTPServer(s *server.MCPServer, host, port string) error {
	httpServer := server.NewStreamableHTTPServer(s)
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Starting MCP HTTP server on %s", addr)
	return httpServer.Start(addr)
}

// Tool handlers

// handleStatusTool provides transport connectivity status
func handleStatusTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// In a real implementation, this would check actual transport statuses
	// For now, return mock status information

	status := map[string]interface{}{
		"whatsapp": map[string]interface{}{
			"connected":    true,
			"last_message": "2 hours ago",
			"uptime":       "99.5%",
		},
		"telegram": map[string]interface{}{
			"connected":  false,
			"last_error": "API rate limit exceeded",
			"uptime":     "85.2%",
		},
		"facebook": map[string]interface{}{
			"connected":    true,
			"last_message": "30 minutes ago",
			"uptime":       "97.8%",
		},
		"shipper": map[string]interface{}{
			"connected":    true,
			"last_request": "15 minutes ago",
			"uptime":       "100%",
		},
	}

	// Format status as readable text
	var statusText string
	for transport, info := range status {
		connInfo := info.(map[string]interface{})
		statusText += fmt.Sprintf("%s: ", transport)
		if connected, ok := connInfo["connected"].(bool); ok && connected {
			statusText += "✓ Connected"
			if last, ok := connInfo["last_message"]; ok {
				statusText += fmt.Sprintf(" (last: %s)", last)
			}
			if uptime, ok := connInfo["uptime"]; ok {
				statusText += fmt.Sprintf(" | uptime: %s", uptime)
			}
		} else {
			statusText += "✗ Disconnected"
			if lastErr, ok := connInfo["last_error"]; ok {
				statusText += fmt.Sprintf(" (%s)", lastErr)
			}
		}
		statusText += "\n"
	}

	return mcp.NewToolResultText(statusText), nil
}

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

// handleDiagnosticsTool runs comprehensive system diagnostics
func handleDiagnosticsTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// In a real implementation, this would run actual diagnostic checks
	// For now, return mock diagnostic results

	diagnostics := []map[string]interface{}{
		{
			"check":   "Configuration Validation",
			"status":  "PASS",
			"details": "All transport configurations are valid",
		},
		{
			"check":   "Database Connectivity",
			"status":  "PASS",
			"details": "Database connections healthy",
		},
		{
			"check":   "Transport Services",
			"status":  "WARN",
			"details": "Telegram transport experiencing rate limiting",
		},
		{
			"check":   "Security Scan",
			"status":  "PASS",
			"details": "No security vulnerabilities detected",
		},
		{
			"check":   "Performance Metrics",
			"status":  "PASS",
			"details": "Response times within acceptable limits",
		},
		{
			"check":   "Rate Limiting",
			"status":  "INFO",
			"details": "Rate limiting active (normal operation)",
		},
	}

	// Format diagnostics
	var diagText string
	passed := 0
	total := len(diagnostics)

	for _, check := range diagnostics {
		status := check["status"].(string)
		if status == "PASS" {
			passed++
		}

		emoji := map[string]string{
			"PASS":  "✓",
			"WARN":  "⚠",
			"INFO":  "ℹ",
			"ERROR": "✗",
		}[status]

		diagText += fmt.Sprintf("%s %s: %s\n   %s\n\n",
			emoji,
			check["check"],
			status,
			check["details"])
	}

	summary := fmt.Sprintf("Diagnostics completed: %d/%d checks passed\n\n", passed, total)
	diagText = summary + diagText

	return mcp.NewToolResultText(diagText), nil
}

// handleConfigTool provides sanitized configuration information
func handleConfigTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get optional transport parameter
	transport := ""
	if args := req.Params.Arguments; args != nil {
		if argsMap, ok := args.(map[string]interface{}); ok {
			if t, ok := argsMap["transport"].(string); ok {
				transport = t
			}
		}
	}

	// In a real implementation, this would read actual config
	// For now, return sanitized mock config
	config := map[string]interface{}{
		"general": map[string]interface{}{
			"log_level": "info",
			"timeout":   30,
		},
		"transports": map[string]interface{}{
			"whatsapp": map[string]interface{}{
				"enabled": true,
				"timeout": 30,
			},
			"telegram": map[string]interface{}{
				"enabled": true,
				"timeout": 25,
			},
			"facebook": map[string]interface{}{
				"enabled": false,
				"reason":  "API key not configured",
			},
			"shipper": map[string]interface{}{
				"enabled": true,
				"timeout": 60,
			},
		},
	}

	// Sanitize config using our sanitizer
	sanitizer := NewSanitizer()
	sanitizedConfig := sanitizer.SanitizeMap(config)

	// Format for specific transport or all
	var configText string
	if transport != "" {
		if transportConfig, ok := sanitizedConfig["transports"].(map[string]interface{})[transport]; ok {
			configText = fmt.Sprintf("Configuration for %s transport:\n", transport)
			if tc, ok := transportConfig.(map[string]interface{}); ok {
				for k, v := range tc {
					configText += fmt.Sprintf("  %s: %v\n", k, v)
				}
			}
		} else {
			configText = fmt.Sprintf("Transport '%s' not found in configuration", transport)
		}
	} else {
		configText = "Current sanitized configuration:\n"
		for section, values := range sanitizedConfig {
			configText += fmt.Sprintf("\n%s:\n", section)
			if sectionMap, ok := values.(map[string]interface{}); ok {
				for k, v := range sectionMap {
					configText += fmt.Sprintf("  %s: %v\n", k, v)
				}
			}
		}
	}

	return mcp.NewToolResultText(configText), nil
}

// handleTestConnectionTool tests connectivity to a specific transport
func handleTestConnectionTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get required transport parameter
	transport := ""
	if args := req.Params.Arguments; args != nil {
		if argsMap, ok := args.(map[string]interface{}); ok {
			if t, ok := argsMap["transport"].(string); ok {
				transport = t
			}
		}
	}

	if transport == "" {
		return mcp.NewToolResultText("Error: transport parameter is required"), nil
	}

	// In a real implementation, this would test actual connectivity
	// For now, return mock test results
	mockResults := map[string]map[string]interface{}{
		"whatsapp": {
			"success": true,
			"latency": "250ms",
			"status":  "Connected",
		},
		"telegram": {
			"success": false,
			"error":   "Rate limit exceeded",
			"status":  "Disconnected",
		},
		"facebook": {
			"success": false,
			"error":   "API key not configured",
			"status":  "Not configured",
		},
		"shipper": {
			"success": true,
			"latency": "180ms",
			"status":  "Connected",
		},
	}

	var resultText string
	if result, ok := mockResults[transport]; ok {
		resultText = fmt.Sprintf("Connection test for %s:\n", transport)
		if success, ok := result["success"].(bool); ok && success {
			resultText += fmt.Sprintf("✓ %s\n", result["status"])
			if latency, ok := result["latency"]; ok {
				resultText += fmt.Sprintf("  Latency: %s\n", latency)
			}
		} else {
			resultText += fmt.Sprintf("✗ %s\n", result["status"])
			if errMsg, ok := result["error"]; ok {
				resultText += fmt.Sprintf("  Error: %s\n", errMsg)
			}
		}
	} else {
		resultText = fmt.Sprintf("Unknown transport: %s", transport)
	}

	return mcp.NewToolResultText(resultText), nil
}

// handleErrorMetricsTool provides error statistics and trends
func handleErrorMetricsTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get optional window parameter
	window := "day"
	if args := req.Params.Arguments; args != nil {
		if argsMap, ok := args.(map[string]interface{}); ok {
			if w, ok := argsMap["window"].(string); ok {
				window = w
			}
		}
	}

	// In a real implementation, this would calculate actual metrics
	// For now, return mock error metrics
	mockMetrics := map[string]map[string]interface{}{
		"hour": {
			"total_errors": 5,
			"error_rate":   0.2,
			"time_range":   "Last hour",
			"top_errors": []map[string]interface{}{
				{"type": "timeout", "count": 3},
				{"type": "rate_limit", "count": 2},
			},
		},
		"day": {
			"total_errors": 23,
			"error_rate":   1.0,
			"time_range":   "Last 24 hours",
			"top_errors": []map[string]interface{}{
				{"type": "timeout", "count": 12},
				{"type": "rate_limit", "count": 8},
				{"type": "auth_failure", "count": 3},
			},
		},
		"week": {
			"total_errors": 156,
			"error_rate":   2.2,
			"time_range":   "Last 7 days",
			"top_errors": []map[string]interface{}{
				{"type": "timeout", "count": 89},
				{"type": "rate_limit", "count": 45},
				{"type": "network_error", "count": 22},
			},
		},
	}

	var metricsText string
	if metrics, ok := mockMetrics[window]; ok {
		metricsText = fmt.Sprintf("Error Metrics (%s):\n\n", metrics["time_range"])
		metricsText += fmt.Sprintf("Total Errors: %v\n", metrics["total_errors"])
		metricsText += fmt.Sprintf("Error Rate: %.1f per hour\n\n", metrics["error_rate"])
		metricsText += "Top Error Types:\n"

		if topErrors, ok := metrics["top_errors"].([]map[string]interface{}); ok {
			for _, err := range topErrors {
				metricsText += fmt.Sprintf("  - %s: %v occurrences\n",
					err["type"], err["count"])
			}
		}
	} else {
		metricsText = fmt.Sprintf("Unknown time window: %s", window)
	}

	return mcp.NewToolResultText(metricsText), nil
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
