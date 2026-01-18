package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

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
