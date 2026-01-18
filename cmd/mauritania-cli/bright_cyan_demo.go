package main

import (
	"fmt"
	"time"
	"github.com/fatih/color"
)

func main() {
	fmt.Println("🎨 Mauritania CLI - Updated Bright Cyan Demo")
	fmt.Println("===========================================")
	
	// Color functions with improved bright cyan
	success := color.New(color.FgGreen, color.Bold)
	error := color.New(color.FgRed, color.Bold)
	warning := color.New(color.FgYellow, color.Bold)
	info := color.New(color.FgHiCyan, color.Bold) // Bright cyan
	command := color.New(color.FgBlue, color.Bold)
	
	// Demo messages
	timestamp := time.Now().Format("15:04:05")
	
	fmt.Printf("%s %s %s %s\n", timestamp, success.Sprintf("✅"), success.Sprintf("SUCCESS"), "Configuration loaded successfully")
	fmt.Printf("%s %s %s %s\n", timestamp, error.Sprintf("❌"), error.Sprintf("ERROR"), "Connection failed: network timeout")
	fmt.Printf("%s %s %s %s\n", timestamp, warning.Sprintf("⚠️"), warning.Sprintf("WARN"), "Transport may be unreliable")
	fmt.Printf("%s %s %s %s\n", timestamp, info.Sprintf("ℹ️"), info.Sprintf("INFO"), "Command queued for execution")
	fmt.Printf("%s %s %s %s\n", timestamp, command.Sprintf("🔧"), command.Sprintf("CMD"), "Executing: ls -la")
	
	fmt.Println("\n🎉 Bright cyan color updated successfully!")
	fmt.Println("Info messages now use bright cyan (#00FFFF) for better visibility!")
}
