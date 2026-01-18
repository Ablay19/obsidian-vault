package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

// Colorize enhances CLI output with colors similar to nushell
type Colorize struct {
	// Fatih colors for basic logging
	success *color.Color
	error   *color.Color
	warning *color.Color
	info    *color.Color
	command *color.Color
	output  *color.Color
	header  *color.Color

	// Lipgloss colors for advanced styling
	primary lipgloss.Color
	border  lipgloss.Color
}

// NewColorize creates a new colorize instance
func NewColorize() *Colorize {
	return &Colorize{
		success: color.New(color.FgGreen, color.Bold),
		error:   color.New(color.FgRed, color.Bold),
		warning: color.New(color.FgYellow, color.Bold),
		info:    color.New(color.FgHiCyan, color.Bold), // Bright cyan
		command: color.New(color.FgBlue, color.Bold),
		output:  color.New(color.FgWhite),
		header:  color.New(color.FgMagenta, color.Bold),
	}
}

// Success prints a success message
func (c *Colorize) Success(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s %s %s\n", c.info.Sprintf("%s", timestamp), c.success.Sprintf("✅"), c.success.Sprintf("SUCCESS"), message)
}

// Error prints an error message
func (c *Colorize) Error(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s %s %s\n", c.info.Sprintf("%s", timestamp), c.error.Sprintf("❌"), c.error.Sprintf("ERROR"), message)
}

// Warning prints a warning message
func (c *Colorize) Warning(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s %s %s\n", c.info.Sprintf("%s", timestamp), c.warning.Sprintf("⚠️"), c.warning.Sprintf("WARN"), message)
}

// Info prints an info message
func (c *Colorize) Info(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s %s %s\n", c.info.Sprintf("%s", timestamp), c.info.Sprintf("ℹ️"), c.info.Sprintf("INFO"), message)
}

// Command prints a command message
func (c *Colorize) Command(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s %s %s %s\n", c.info.Sprintf("%s", timestamp), c.command.Sprintf("🔧"), c.command.Sprintf("CMD"), message)
}

// Header prints a header
func (c *Colorize) Header(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s\n", c.header.Sprintf(message))
}

// Print prints regular output
func (c *Colorize) Print(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s", c.output.Sprintf(message))
}

// Println prints regular output with newline
func (c *Colorize) Println(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s\n", c.output.Sprintf(message))
}

// Printf prints formatted output
func (c *Colorize) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s", c.output.Sprintf(message))
}

// FormatCommand formats a command with styling
func (c *Colorize) FormatCommand(cmd string) string {
	return c.command.Sprintf("🔧 %s", cmd)
}

// FormatPath formats a file path with styling
func (c *Colorize) FormatPath(path string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")).Render(path)
}

// FormatID formats an ID with styling
func (c *Colorize) FormatID(id string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#D19A66")).Render(id)
}

// FormatTimestamp formats a timestamp
func (c *Colorize) FormatTimestamp(t time.Time) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#5C6370")).Render(t.Format("15:04:05"))
}

// FormatStatus formats a status with appropriate color
func (c *Colorize) FormatStatus(status string) string {
	var color lipgloss.Color
	switch strings.ToLower(status) {
	case "success", "completed", "sent", "healthy":
		color = lipgloss.Color("#98C379") // Green
	case "error", "failed", "unhealthy":
		color = lipgloss.Color("#E06C75") // Red
	case "warning", "pending", "queued":
		color = lipgloss.Color("#E5C07B") // Yellow
	case "info", "running", "connected":
		color = lipgloss.Color("#56B6C2") // Cyan
	default:
		color = lipgloss.Color("#ABB2BF") // Default
	}
	return lipgloss.NewStyle().Foreground(color).Render(status)
}

// FormatTransport formats a transport name
func (c *Colorize) FormatTransport(transport string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")).Render(transport)
}

// Separator creates a styled separator line
func (c *Colorize) Separator() string {
	return lipgloss.NewStyle().Foreground(c.border).Render(strings.Repeat("─", 50))
}

// SectionHeader creates a section header
func (c *Colorize) SectionHeader(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(c.primary).
		MarginTop(1).
		MarginBottom(1).
		Render(fmt.Sprintf("┌─ %s ─┐", title))
}

// SectionFooter creates a section footer
func (c *Colorize) SectionFooter(title string) string {
	return lipgloss.NewStyle().
		Foreground(c.border).
		Render(fmt.Sprintf("└%s┘", strings.Repeat("─", len(title)+4)))
}

// SuccessBox creates a success-styled box
func (c *Colorize) SuccessBox(title string, content string) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#98C379")). // Green
		MarginBottom(1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#98C379")).
		Padding(1)

	return headerStyle.Render("✅ "+title) + "\n" + borderStyle.Render(content)
}

// ErrorBox creates an error-styled box
func (c *Colorize) ErrorBox(title string, content string) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E06C75")). // Red
		MarginBottom(1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E06C75")).
		Padding(1)

	return headerStyle.Render("❌ "+title) + "\n" + borderStyle.Render(content)
}

// InfoBox creates an info-styled box
func (c *Colorize) InfoBox(title string, content string) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")). // Pure bright cyan (better visibility)
		MarginBottom(1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1)

	return headerStyle.Render("ℹ️ "+title) + "\n" + borderStyle.Render(content)
}

// Table creates a simple styled table
func (c *Colorize) Table(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Create table
	var result strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(c.primary)
	for i, header := range headers {
		result.WriteString(headerStyle.Render(fmt.Sprintf("%-*s", colWidths[i], header)))
		if i < len(headers)-1 {
			result.WriteString(" │ ")
		}
	}
	result.WriteString("\n")

	// Separator
	sepStyle := lipgloss.NewStyle().Foreground(c.border)
	for i, width := range colWidths {
		result.WriteString(sepStyle.Render(strings.Repeat("─", width)))
		if i < len(colWidths)-1 {
			result.WriteString(sepStyle.Render("─┼─"))
		}
	}
	result.WriteString("\n")

	// Rows
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ABB2BF"))
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				result.WriteString(rowStyle.Render(fmt.Sprintf("%-*s", colWidths[i], cell)))
			}
			if i < len(row)-1 {
				result.WriteString(" │ ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// ProgressBar creates a styled progress bar
func (c *Colorize) ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percentage := float64(current) / float64(total)
	filled := int(float64(width) * percentage)

	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	var barColor lipgloss.Color
	if percentage < 0.3 {
		barColor = lipgloss.Color("#E06C75") // Red
	} else if percentage < 0.7 {
		barColor = lipgloss.Color("#E5C07B") // Yellow
	} else {
		barColor = lipgloss.Color("#98C379") // Green
	}

	return fmt.Sprintf("[%s] %d/%d (%.1f%%)",
		lipgloss.NewStyle().Foreground(barColor).Render(bar),
		current, total, percentage*100)
}

// ErrorBox creates an error-styled box
// GetColorize returns a global colorize instance
var globalColorize *Colorize

func init() {
	globalColorize = NewColorize()
}

// Success is a convenience function using the global colorize instance
func Success(format string, args ...interface{}) {
	globalColorize.Success(format, args...)
}

// Error is a convenience function using the global colorize instance
func Error(format string, args ...interface{}) {
	globalColorize.Error(format, args...)
}

// Warning is a convenience function using the global colorize instance
func Warning(format string, args ...interface{}) {
	globalColorize.Warning(format, args...)
}

// Info is a convenience function using the global colorize instance
func Info(format string, args ...interface{}) {
	globalColorize.Info(format, args...)
}

// Command is a convenience function using the global colorize instance
func Command(format string, args ...interface{}) {
	globalColorize.Command(format, args...)
}

// Header is a convenience function using the global colorize instance
func Header(format string, args ...interface{}) {
	globalColorize.Header(format, args...)
}

// Print is a convenience function using the global colorize instance
func Print(format string, args ...interface{}) {
	globalColorize.Print(format, args...)
}

// Println is a convenience function using the global colorize instance
func Println(format string, args ...interface{}) {
	globalColorize.Println(format, args...)
}

// InfoBox is a convenience function for styled info boxes
func InfoBox(title string, content string) string {
	return globalColorize.InfoBox(title, content)
}

// SuccessBox is a convenience function for styled success boxes
func SuccessBox(title string, content string) string {
	return globalColorize.SuccessBox(title, content)
}

// ErrorBox is a convenience function for styled error boxes
func ErrorBox(title string, content string) string {
	return globalColorize.ErrorBox(title, content)
}

// FormatCommand formats a command with styling
func FormatCommand(cmd string) string {
	return globalColorize.FormatCommand(cmd)
}

// FormatPath formats a file path with styling
func FormatPath(path string) string {
	return globalColorize.FormatPath(path)
}

// FormatID formats an ID with styling
func FormatID(id string) string {
	return globalColorize.FormatID(id)
}

// FormatTimestamp formats a timestamp
func FormatTimestamp(t time.Time) string {
	return globalColorize.FormatTimestamp(t)
}

// FormatStatus formats a status with appropriate color
func FormatStatus(status string) string {
	return globalColorize.FormatStatus(status)
}

// FormatTransport formats a transport name
func FormatTransport(transport string) string {
	return globalColorize.FormatTransport(transport)
}

// Separator creates a styled separator line
func Separator() string {
	return globalColorize.Separator()
}

// SectionHeader creates a section header
func SectionHeader(title string) string {
	return globalColorize.SectionHeader(title)
}

// SectionFooter creates a section footer
func SectionFooter(title string) string {
	return globalColorize.SectionFooter(title)
}

// Table creates a styled table
func Table(headers []string, rows [][]string) string {
	return globalColorize.Table(headers, rows)
}
