package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// ShipperResultParser parses and formats results from SM APOS Shipper
type ShipperResultParser struct {
	logger *log.Logger
}

// NewShipperResultParser creates a new shipper result parser
func NewShipperResultParser(logger *log.Logger) *ShipperResultParser {
	return &ShipperResultParser{
		logger: logger,
	}
}

// ParseExecutionResult parses a raw execution result from the shipper
func (srp *ShipperResultParser) ParseExecutionResult(rawResult string) (*models.CommandResult, error) {
	var result models.CommandResult

	// Try to parse as JSON first (structured response)
	if strings.HasPrefix(strings.TrimSpace(rawResult), "{") {
		if err := json.Unmarshal([]byte(rawResult), &result); err == nil {
			srp.logger.Printf("Parsed structured result for command %s", result.ID)
			return &result, nil
		}
	}

	// Parse as plain text output
	result = srp.parsePlainTextResult(rawResult)
	return &result, nil
}

// parsePlainTextResult parses a plain text command result
func (srp *ShipperResultParser) parsePlainTextResult(output string) models.CommandResult {
	result := models.CommandResult{
		Status:        "success", // Assume success unless we find error indicators
		ExecutionTime: 1000,      // Default 1 second
		TransportUsed: "sm_apos",
		CompletedAt:   time.Now(),
	}

	lines := strings.Split(output, "\n")

	// Look for exit code indicators
	exitCodeRegex := regexp.MustCompile(`Exit code: (\d+)`)
	if matches := exitCodeRegex.FindStringSubmatch(output); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			result.ExitCode = code
			if code != 0 {
				result.Status = "failure"
			}
		}
	}

	// Look for execution time
	timeRegex := regexp.MustCompile(`Execution time: (\d+)ms`)
	if matches := timeRegex.FindStringSubmatch(output); len(matches) > 1 {
		if duration, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			result.ExecutionTime = duration
		}
	}

	// Separate stdout and stderr
	var stdout, stderr strings.Builder
	inStderr := false

	for _, line := range lines {
		// Check for stderr markers
		if strings.Contains(line, "STDERR") || strings.Contains(line, "Error:") {
			inStderr = true
		}

		if inStderr {
			if stderr.Len() > 0 {
				stderr.WriteString("\n")
			}
			stderr.WriteString(line)
		} else {
			if stdout.Len() > 0 {
				stdout.WriteString("\n")
			}
			stdout.WriteString(line)
		}
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Clean up the output
	result.Stdout = srp.cleanOutput(result.Stdout)
	result.Stderr = srp.cleanOutput(result.Stderr)

	return result
}

// FormatResultForDisplay formats a command result for user display
func (srp *ShipperResultParser) FormatResultForDisplay(result *models.CommandResult) string {
	var output strings.Builder

	// Add header
	output.WriteString("```\n")
	output.WriteString("$ command executed via SM APOS Shipper\n")
	output.WriteString("```\n\n")

	// Add metadata
	output.WriteString(fmt.Sprintf("⏱️ **Execution Time:** %d ms\n", result.ExecutionTime))
	output.WriteString(fmt.Sprintf("📊 **Exit Code:** %d\n", result.ExitCode))
	output.WriteString(fmt.Sprintf("🚛 **Transport:** %s\n", result.TransportUsed))

	if result.Cost > 0 {
		output.WriteString(fmt.Sprintf("💰 **Cost:** $%.4f\n", result.Cost))
	}

	output.WriteString("\n")

	// Add output sections
	if result.Status == "success" {
		output.WriteString("✅ **Output:**\n")
		output.WriteString("```\n")
		if result.Stdout != "" {
			// Limit output size for display
			displayOutput := result.Stdout
			if len(displayOutput) > 2000 {
				displayOutput = displayOutput[:2000] + "\n... (output truncated)"
			}
			output.WriteString(displayOutput)
		} else {
			output.WriteString("(no output)")
		}
		output.WriteString("\n```\n")
	} else {
		output.WriteString("❌ **Result:**\n")
		output.WriteString("```\n")
		output.WriteString(result.Status)
		if result.Stderr != "" {
			output.WriteString("\n\nSTDERR:\n")
			errorOutput := result.Stderr
			if len(errorOutput) > 2000 {
				errorOutput = errorOutput[:2000] + "\n... (error output truncated)"
			}
			output.WriteString(errorOutput)
		}
		output.WriteString("\n```\n")
	}

	return output.String()
}

// ValidateResult validates that a parsed result is complete and correct
func (srp *ShipperResultParser) ValidateResult(result *models.CommandResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	if result.ID == "" {
		return fmt.Errorf("result must have an ID")
	}

	if result.CommandID == "" {
		return fmt.Errorf("result must have a command ID")
	}

	// Validate status
	validStatuses := []string{"success", "failure", "partial", "timeout"}
	valid := false
	for _, status := range validStatuses {
		if result.Status == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid result status: %s", result.Status)
	}

	// Validate exit code
	if result.ExitCode < 0 || result.ExitCode > 255 {
		srp.logger.Printf("Warning: Unusual exit code: %d", result.ExitCode)
	}

	// Validate execution time
	if result.ExecutionTime < 0 {
		return fmt.Errorf("execution time cannot be negative")
	}
	if result.ExecutionTime > 300000 { // 5 minutes in milliseconds
		srp.logger.Printf("Warning: Very long execution time: %d ms", result.ExecutionTime)
	}

	return nil
}

// ExtractMetrics extracts performance metrics from command results
func (srp *ShipperResultParser) ExtractMetrics(result *models.CommandResult) map[string]interface{} {
	metrics := map[string]interface{}{
		"exit_code":      result.ExitCode,
		"execution_time": result.ExecutionTime,
		"output_size":    len(result.Stdout),
		"error_size":     len(result.Stderr),
		"transport":      result.TransportUsed,
		"cost":           result.Cost,
		"success":        result.Status == "success",
	}

	// Extract additional metrics from output
	if strings.Contains(result.Stdout, "lines") {
		// Could be output from wc -l or similar
		lineRegex := regexp.MustCompile(`(\d+)\s+lines?`)
		if matches := lineRegex.FindStringSubmatch(result.Stdout); len(matches) > 1 {
			if lines, err := strconv.Atoi(matches[1]); err == nil {
				metrics["output_lines"] = lines
			}
		}
	}

	return metrics
}

// MergePartialResults merges multiple partial results into a complete result
func (srp *ShipperResultParser) MergePartialResults(results []*models.CommandResult) *models.CommandResult {
	if len(results) == 0 {
		return nil
	}

	if len(results) == 1 {
		return results[0]
	}

	// Use the first result as base
	merged := *results[0]

	// Merge outputs
	var stdout, stderr strings.Builder

	for _, result := range results {
		if result.Stdout != "" {
			if stdout.Len() > 0 {
				stdout.WriteString("\n")
			}
			stdout.WriteString(result.Stdout)
		}
		if result.Stderr != "" {
			if stderr.Len() > 0 {
				stderr.WriteString("\n")
			}
			stderr.WriteString(result.Stderr)
		}

		// Sum execution times
		merged.ExecutionTime += result.ExecutionTime

		// Use worst exit code
		if result.ExitCode > merged.ExitCode {
			merged.ExitCode = result.ExitCode
		}

		// Sum costs
		merged.Cost += result.Cost
	}

	merged.Stdout = stdout.String()
	merged.Stderr = stderr.String()

	// Update status based on merged results
	if merged.ExitCode != 0 {
		merged.Status = "failure"
	} else {
		merged.Status = "success"
	}

	return &merged
}

// ParseShipperError parses shipper-specific error messages
func (srp *ShipperResultParser) ParseShipperError(errorOutput string) (errorType string, message string) {
	// Common shipper error patterns
	if strings.Contains(errorOutput, "timeout") {
		return "timeout", "Command execution timed out"
	}

	if strings.Contains(errorOutput, "permission denied") {
		return "permission", "Insufficient permissions to execute command"
	}

	if strings.Contains(errorOutput, "command not found") {
		return "not_found", "Command not found in system"
	}

	if strings.Contains(errorOutput, "network") {
		return "network", "Network connectivity issue"
	}

	if strings.Contains(errorOutput, "quota") {
		return "quota", "Resource quota exceeded"
	}

	return "unknown", errorOutput
}

// FormatErrorForDisplay formats an error for user-friendly display
func (srp *ShipperResultParser) FormatErrorForDisplay(errorType, message string) string {
	emoji := "❌"
	suggestion := ""

	switch errorType {
	case "timeout":
		emoji = "⏰"
		suggestion = "Try breaking the command into smaller steps or increasing the timeout."
	case "permission":
		emoji = "🔒"
		suggestion = "Check that you have the necessary permissions or contact your administrator."
	case "not_found":
		emoji = "🔍"
		suggestion = "Verify that the command is installed and available in the system PATH."
	case "network":
		emoji = "📡"
		suggestion = "Check your network connection and try again."
	case "quota":
		emoji = "📊"
		suggestion = "You've reached your usage limit. Try again later or upgrade your plan."
	}

	return fmt.Sprintf("%s **%s Error**\n\n%s\n\n💡 *Suggestion:* %s",
		emoji, strings.Title(errorType), message, suggestion)
}

// cleanOutput cleans up command output by removing shipper-specific artifacts
func (srp *ShipperResultParser) cleanOutput(output string) string {
	if output == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Remove shipper metadata lines
		if strings.HasPrefix(line, "[SHIPPER]") ||
			strings.HasPrefix(line, "[EXECUTOR]") ||
			strings.HasPrefix(line, "[METADATA]") {
			continue
		}

		// Remove empty lines at start/end
		if strings.TrimSpace(line) != "" || len(cleanLines) > 0 {
			cleanLines = append(cleanLines, line)
		}
	}

	// Remove trailing empty lines
	for len(cleanLines) > 0 && strings.TrimSpace(cleanLines[len(cleanLines)-1]) == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}

	return strings.Join(cleanLines, "\n")
}

// EstimateOutputSize estimates the size of command output for quota management
func (srp *ShipperResultParser) EstimateOutputSize(command string) int {
	// Rough estimation based on command type
	baseSize := 100 // bytes

	if strings.Contains(command, "ls") {
		baseSize = 1000
	} else if strings.Contains(command, "cat") || strings.Contains(command, "grep") {
		baseSize = 10000
	} else if strings.Contains(command, "find") {
		baseSize = 50000
	} else if strings.Contains(command, "ps") || strings.Contains(command, "top") {
		baseSize = 5000
	}

	return baseSize
}
