package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Nushell-inspired color palette
const (
	ColorReset         = "\033[0m"
	ColorBlack         = "\033[30m"
	ColorRed           = "\033[31m"
	ColorGreen         = "\033[32m"
	ColorYellow        = "\033[33m"
	ColorBlue          = "\033[34m"
	ColorMagenta       = "\033[35m"
	ColorCyan          = "\033[36m"
	ColorWhite         = "\033[37m"
	ColorGray          = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// Nushell-specific colors
	NushellBlue   = "\033[38;5;27m"  // Deep blue like Nushell
	NushellGreen  = "\033[38;5;28m"  // Forest green
	NushellPurple = "\033[38;5;93m"  // Purple
	NushellOrange = "\033[38;5;208m" // Orange
	NushellPink   = "\033[38;5;205m" // Pink
)

// NushellHandler provides colored, structured logging similar to Nushell
type NushellHandler struct {
	slog.Handler
	useColors bool
}

// NewNushellHandler creates a handler for structured logging with Nushell-like colors
func NewNushellHandler(opts *slog.HandlerOptions) *NushellHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Set default level to INFO if not specified
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}

	return &NushellHandler{
		Handler:   slog.NewTextHandler(os.Stdout, opts),
		useColors: shouldUseColors(),
	}
}

// shouldUseColors determines if colors should be used based on terminal capabilities
func shouldUseColors() bool {
	// Check if running in a terminal
	if !isTerminal() {
		return false
	}

	// Check NO_COLOR environment variable (standard)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		return false
	}

	return true
}

// isTerminal checks if output is going to a terminal
func isTerminal() bool {
	// Simple check for now - could be enhanced
	return true // Assume terminal for CLI usage
}

// Handle formats log records with Nushell-like colors (no icons)
func (h *NushellHandler) Handle(ctx context.Context, r slog.Record) error {
	var color string

	switch r.Level {
	case slog.LevelDebug:
		color = ColorBrightCyan
	case slog.LevelInfo:
		color = NushellBlue
	case slog.LevelWarn:
		color = ColorBrightYellow
	case slog.LevelError:
		color = ColorBrightRed
	default:
		color = ColorGray
	}

	// Get component from attributes and collect other attributes
	component := "CLI"
	var attrs []string

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" {
			component = a.Value.String()
		} else {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value))
		}
		return true
	})

	// Format message with Nushell-style colors (no icons)
	if h.useColors {
		// Nushell-style: level component message
		levelStr := strings.ToUpper(r.Level.String())
		fmt.Fprintf(os.Stdout, "%s%s %s %s%s",
			color, levelStr, component, r.Message, ColorReset)

		// Add attributes if any
		if len(attrs) > 0 {
			fmt.Fprintf(os.Stdout, " %s\n", strings.Join(attrs, " "))
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	} else {
		fmt.Fprintf(os.Stdout, "%s %s %s",
			strings.ToUpper(r.Level.String()), component, r.Message)

		if len(attrs) > 0 {
			fmt.Fprintf(os.Stdout, " %s\n", strings.Join(attrs, " "))
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	}

	return nil
}

// NushellJSONHandler provides JSON logging in Nushell style
type NushellJSONHandler struct {
	slog.Handler
}

// NewNushellJSONHandler creates a JSON handler for structured logging
func NewNushellJSONHandler(opts *slog.HandlerOptions) *NushellJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &NushellJSONHandler{
		Handler: slog.NewJSONHandler(os.Stdout, opts),
	}
}

// Logger provides a convenient interface for structured logging
type Logger struct {
	textLogger *slog.Logger
	jsonLogger *slog.Logger
	useJSON    bool
}

// NewLogger creates a new logger with Nushell-like formatting
func NewLogger(component string) *Logger {
	textHandler := NewNushellHandler(&slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	jsonHandler := NewNushellJSONHandler(&slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &Logger{
		textLogger: slog.New(textHandler).With("component", component),
		jsonLogger: slog.New(jsonHandler).With("component", component),
		useJSON:    false,
	}
}

// SetJSON enables JSON output mode
func (l *Logger) SetJSON(json bool) {
	l.useJSON = json
}

// currentLogger returns the appropriate logger based on mode
func (l *Logger) currentLogger() *slog.Logger {
	if l.useJSON {
		return l.jsonLogger
	}
	return l.textLogger
}

// Success logs a success message with green color
func (l *Logger) Success(message string, args ...any) {
	// Use slog's Info level (no icon prefix)
	l.currentLogger().Info(message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, args ...any) {
	if err != nil {
		args = append(args, "error", err)
	}
	l.currentLogger().Error(message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...any) {
	l.currentLogger().Warn(message, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...any) {
	l.currentLogger().Debug(message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...any) {
	l.currentLogger().Info(message, args...)
}

// Banner displays a prominent banner message
func (l *Logger) Banner(message string) {
	width := 60
	padding := (width - len(message) - 4) / 2 // 4 for borders and spaces
	if padding < 0 {
		padding = 0
	}

	border := strings.Repeat("═", width)
	paddingStr := strings.Repeat(" ", padding)

	fmt.Fprintf(os.Stdout, "%s%s%s\n", ColorCyan, border, ColorReset)
	fmt.Fprintf(os.Stdout, "%s║%s %s %s║%s\n",
		ColorCyan, paddingStr, message, paddingStr, ColorReset)
	fmt.Fprintf(os.Stdout, "%s%s%s\n", ColorCyan, border, ColorReset)
}

// Section displays a section header
func (l *Logger) Section(title string) {
	fmt.Fprintf(os.Stdout, "\n%s═══ %s ═══%s\n", ColorMagenta, title, ColorReset)
}

// Table displays data in a formatted table (similar to Nushell)
func (l *Logger) Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		l.Info("No data to display")
		return
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

	// Create format string
	var format strings.Builder
	for i, width := range colWidths {
		if i > 0 {
			format.WriteString(" │ ")
		}
		format.WriteString(fmt.Sprintf("%%-%ds", width))
	}
	formatStr := format.String()

	// Print header
	fmt.Fprintf(os.Stdout, "%s", ColorCyan)
	fmt.Fprintf(os.Stdout, formatStr+"\n", stringSliceToInterface(headers)...)

	// Print separator
	var separator strings.Builder
	for i, width := range colWidths {
		if i > 0 {
			separator.WriteString("─┼─")
		}
		separator.WriteString(strings.Repeat("─", width))
	}
	fmt.Fprintf(os.Stdout, "%s%s%s\n", ColorGray, separator.String(), ColorReset)

	// Print rows
	for _, row := range rows {
		fmt.Fprintf(os.Stdout, formatStr+"\n", stringSliceToInterface(row)...)
	}

	fmt.Fprintf(os.Stdout, "%s", ColorReset)
}

// Progress shows a progress indicator for long-running operations
func (l *Logger) Progress(current, total int, message string) {
	width := 30
	filled := int(float64(current) / float64(total) * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := float64(current) / float64(total) * 100

	fmt.Fprintf(os.Stdout, "\r%s %s%s %.1f%% (%d/%d)%s",
		ColorCyan+"["+ColorReset, ColorGreen, bar, percentage, current, total, ColorReset)

	if current == total {
		fmt.Fprintf(os.Stdout, " %sComplete!%s\n", ColorGreen, ColorReset)
	}
}

// CommandOutput displays command execution results in a structured way
func (l *Logger) CommandOutput(cmd string, success bool, output string, duration time.Duration) {
	var status string
	if success {
		status = ColorGreen + "SUCCESS" + ColorReset
	} else {
		status = ColorRed + "FAILED" + ColorReset
	}

	fmt.Fprintf(os.Stdout, "%s | %s | %v\n",
		status, cmd, duration.Round(time.Millisecond))

	if output != "" {
		// Indent output
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			fmt.Fprintf(os.Stdout, "%s    %s%s\n", ColorGray, line, ColorReset)
		}
	}
}

// stringSliceToInterface converts []string to []interface{}
func stringSliceToInterface(slice []string) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}
