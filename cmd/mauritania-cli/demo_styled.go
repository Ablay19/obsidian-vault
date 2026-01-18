//go:build demo

package main

import (
	"fmt"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/ui"
)

func main() {
	fmt.Println("🎨 Mauritania CLI - Styled Output Demo")
	fmt.Println("=====================================")
	fmt.Println()

	// Demo all the styling features
	ui.Header("📋 Current Configuration")

	// Database section
	dbContent := "Type: sqlite\nPath: ./data/mauritania-cli.db"
	ui.Println(ui.InfoBox("Database", dbContent))

	// Transports section
	transportContent := "Default: social_media\n\nSocial Media:\n  WhatsApp: ❌ not configured\n  Telegram: ❌ not configured\nShipper: not configured"
	ui.Println(ui.InfoBox("Transports", transportContent))

	// Network section
	networkContent := "Timeout: 30 seconds\nRetry Attempts: 3\nOffline Mode: false"
	ui.Println(ui.InfoBox("Network", networkContent))

	fmt.Println()
	ui.Header("📊 System Status")

	// Platform info
	platformContent := "Type: android\nMobile optimizations: enabled\nDocker: ✅ available"
	ui.Println(ui.InfoBox("Platform", platformContent))

	// Network status
	networkStatusContent := "Connectivity: mobile\nStatus: ✅ Connected\nLatency: 150ms\nLast Checked: " + ui.FormatTimestamp(time.Now())
	ui.Println(ui.InfoBox("Network Status", networkStatusContent))

	// System health
	healthContent := "Database: " + ui.FormatStatus("healthy") + "\nNetwork: " + ui.FormatStatus("healthy") + "\nStorage: " + ui.FormatStatus("healthy")
	ui.Println(ui.InfoBox("System Health", healthContent))

	fmt.Println()
	ui.Header("📤 Command Examples")

	// Command result
	resultContent := "ID: " + ui.FormatID("abc123def") + "\nCommand: " + ui.FormatCommand("ls -la") + "\nTransport: " + ui.FormatTransport("whatsapp") + "\nStatus: " + ui.FormatStatus("queued")
	ui.Println(ui.SuccessBox("Command Queued", resultContent))

	fmt.Println()
	ui.Info("Command will be retried automatically when network connectivity returns")

	fmt.Println()
	ui.Header("📝 Log Messages")

	// Demo log messages
	ui.Success("Configuration loaded successfully")
	ui.Error("Connection failed: network timeout")
	ui.Warning("Transport may be unreliable")
	ui.Info("Command queued for execution")
	ui.Command("Executing: ls -la")

	fmt.Println()
	ui.Header("📊 Table Display")

	// Demo table
	headers := []string{"ID", "Status", "Time"}
	rows := [][]string{
		{ui.FormatID("abc123"), ui.FormatStatus("success"), "15:04"},
		{ui.FormatID("def456"), ui.FormatStatus("queued"), "15:03"},
	}
	ui.Println(ui.Table(headers, rows))

	fmt.Println()
	fmt.Println("🎉 " + ui.FormatStatus("SUCCESS") + " Lipgloss styling is working perfectly!")
	fmt.Println("The Mauritania CLI now has beautiful, consistent styling throughout! 🚀✨")
}
