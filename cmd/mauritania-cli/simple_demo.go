//go:build demo

package main

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

// Simple demo without lipgloss dependencies
func main() {
	fmt.Println("🎨 Mauritania CLI - Styled Output Demo")
	fmt.Println("=====================================")

	// Color functions
	success := color.New(color.FgGreen, color.Bold)
	error := color.New(color.FgRed, color.Bold)
	warning := color.New(color.FgYellow, color.Bold)
	info := color.New(color.FgCyan)
	command := color.New(color.FgBlue, color.Bold)

	// Demo messages
	timestamp := time.Now().Format("15:04:05")

	fmt.Printf("%s %s %s %s\n", timestamp, success.Sprintf("✅"), success.Sprintf("SUCCESS"), "Configuration loaded successfully")
	fmt.Printf("%s %s %s %s\n", timestamp, error.Sprintf("❌"), error.Sprintf("ERROR"), "Connection failed: network timeout")
	fmt.Printf("%s %s %s %s\n", timestamp, warning.Sprintf("⚠️"), warning.Sprintf("WARN"), "Transport may be unreliable")
	fmt.Printf("%s %s %s %s\n", timestamp, info.Sprintf("ℹ️"), info.Sprintf("INFO"), "Command queued for execution")
	fmt.Printf("%s %s %s %s\n", timestamp, command.Sprintf("🔧"), command.Sprintf("CMD"), "Executing: ls -la")

	fmt.Println("\n🎉 Colorized output is working perfectly!")
}
