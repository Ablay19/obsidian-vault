package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Colorize enhances CLI output with colors similar to nushell
type Colorize struct {
	success *color.Color
	error   *color.Color
	warning *color.Color
	info    *color.Color
	command *color.Color
	output  *color.Color
	header  *color.Color
}

// NewColorize creates a new colorize instance
func NewColorize() *Colorize {
	return &Colorize{
		success: color.New(color.FgGreen, color.Bold),
		error:   color.New(color.FgRed, color.Bold),
		warning: color.New(color.FgYellow, color.Bold),
		info:    color.New(color.FgCyan),
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
