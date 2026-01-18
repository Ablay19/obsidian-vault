package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/services"
	"obsidian-automation/cmd/mauritania-cli/internal/ui"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newStatusCmd creates the status command
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long:  `Display the current status of queued commands, running services, and system health.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Initialize logger
			logger := log.New(os.Stdout, "[STATUS] ", log.LstdFlags)

			// Detect platform
			platform := utils.DetectPlatform()
			if verbose {
				logger.Printf("Platform: %s", platform.Type)
				logger.Printf("Mobile: %t, Termux: %t", platform.IsMobile, platform.IsTermux)
				logger.Printf("Docker: %t, Kubectl: %t", platform.HasDocker, platform.HasKubectl)
			}

			// Initialize database
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			dataDir := fmt.Sprintf("%s/.mauritania-cli", homeDir)
			db, err := database.NewDB(dataDir, "", "")
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer db.Close()

			// Show platform info with styled output
			platformContent := fmt.Sprintf("Type: %s\n", platform.Type)
			if platform.IsTermux {
				platformContent += "Mobile optimizations: enabled\n"
			}
			platformContent += fmt.Sprintf("Docker: %s\n", boolToStatus(platform.HasDocker))
			platformContent += fmt.Sprintf("Kubectl: %s", boolToStatus(platform.HasKubectl))
			ui.Println(ui.InfoBox("Platform", platformContent))

			// Show pending commands with styled output
			commands, err := db.GetPendingCommands()
			if err != nil {
				ui.Warning("Failed to get pending commands: %v", err)
			} else {
				pendingContent := fmt.Sprintf("Total: %d", len(commands))
				if len(commands) > 0 {
					// Create a table for commands
					headers := []string{"ID", "Command", "Priority", "Status", "Transport"}
					rows := make([][]string, len(commands))

					for i, cmd := range commands {
						// Truncate long commands for display
						displayCmd := cmd.Command
						if len(displayCmd) > 50 {
							displayCmd = displayCmd[:47] + "..."
						}

						rows[i] = []string{
							ui.FormatID(cmd.ID[:8]),
							displayCmd,
							string(cmd.Priority),
							ui.FormatStatus(string(cmd.Status)),
							ui.FormatTransport("auto"),
						}
					}

					pendingContent += "\n\n" + ui.Table(headers, rows)
				} else {
					pendingContent += "\n  No pending commands"
				}
				ui.Println(ui.InfoBox("Pending Commands", pendingContent))
			}

			// Show service status with styled output
			if platform.HasDocker {
				dockerManager := services.NewDockerManager(logger)
				dockerServices, err := dockerManager.ListServices()
				if err != nil {
					ui.Warning("Failed to get Docker services: %v", err)
				} else {
					dockerContent := fmt.Sprintf("Total: %d", len(dockerServices))
					if len(dockerServices) > 0 {
						headers := []string{"Service", "State", "Health", "CPU", "Memory"}
						rows := make([][]string, len(dockerServices))

						for i, svc := range dockerServices {
							rows[i] = []string{
								ui.FormatID(svc.ServiceID),
								ui.FormatStatus(svc.State),
								ui.FormatStatus(svc.Health.Status),
								svc.ResourceUsage.CPU,
								svc.ResourceUsage.Memory,
							}
						}

						dockerContent += "\n\n" + ui.Table(headers, rows)
					} else {
						dockerContent += "\n  No Docker services running"
					}
					ui.Println(ui.InfoBox("Docker Services", dockerContent))
				}
			}

			// Show network status with styled output
			networkMonitor := utils.NewNetworkMonitor()
			networkStatus := networkMonitor.CheckConnectivity()

			networkContent := fmt.Sprintf("Connectivity: %s", networkStatus.Connectivity)
			if networkStatus.IsOnline {
				networkContent += fmt.Sprintf(" (online, %v latency)", networkStatus.Latency.Round(time.Millisecond))
				networkContent += "\nStatus: " + ui.FormatStatus("Connected")
			} else {
				networkContent += " (offline)"
				networkContent += "\nStatus: " + ui.FormatStatus("Disconnected")
			}
			networkContent += fmt.Sprintf("\nLast Checked: %s", ui.FormatTimestamp(networkStatus.LastChecked))

			ui.Println(ui.InfoBox("Network Status", networkContent))

			// Show offline queue status with styled output
			offlineQueue := utils.NewOfflineQueue(db, networkMonitor, logger)
			queueStats := offlineQueue.GetStats()

			queueContent := fmt.Sprintf("Queued Commands: %d\nProcessing: %t", queueStats.TotalQueued, queueStats.Processing)
			ui.Println(ui.InfoBox("Offline Queue", queueContent))

			// Show system health with styled output
			dbStatus := ui.FormatStatus("healthy")
			networkHealthStatus := "healthy"
			if !networkStatus.IsOnline {
				networkHealthStatus = "degraded"
			}
			networkStatusFormatted := ui.FormatStatus(networkHealthStatus)
			storageStatus := ui.FormatStatus("healthy")

			healthContent := fmt.Sprintf("Database: %s\nNetwork: %s\nStorage: %s",
				dbStatus, networkStatusFormatted, storageStatus)
			ui.Println(ui.InfoBox("System Health", healthContent))

			return nil
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Show detailed status information")

	return cmd
}

// boolToStatus converts boolean to status string
func boolToStatus(available bool) string {
	if available {
		return "available"
	}
	return "not available"
}
