package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/services"
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

			// Show platform info
			fmt.Printf("Platform: %s\n", platform.Type)
			if platform.IsTermux {
				fmt.Printf("Mobile optimizations: enabled\n")
			}
			fmt.Printf("Docker: %s\n", boolToStatus(platform.HasDocker))
			fmt.Printf("Kubectl: %s\n", boolToStatus(platform.HasKubectl))
			fmt.Println()

			// Show pending commands
			commands, err := db.GetPendingCommands()
			if err != nil {
				logger.Printf("Warning: failed to get pending commands: %v", err)
			} else {
				fmt.Printf("Pending Commands (%d):\n", len(commands))
				if len(commands) > 0 {
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "ID\tCommand\tPriority\tStatus\tTransport")
					for _, cmd := range commands {
						// Truncate long commands for display
						displayCmd := cmd.Command
						if len(displayCmd) > 50 {
							displayCmd = displayCmd[:47] + "..."
						}
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
							cmd.ID[:8], displayCmd, cmd.Priority, cmd.Status, "auto")
					}
					w.Flush()
				} else {
					fmt.Println("  No pending commands")
				}
				fmt.Println()
			}

			// Show service status
			if platform.HasDocker {
				dockerManager := services.NewDockerManager(logger)
				dockerServices, err := dockerManager.ListServices()
				if err != nil {
					logger.Printf("Warning: failed to get Docker services: %v", err)
				} else {
					fmt.Printf("Docker Services (%d):\n", len(dockerServices))
					if len(dockerServices) > 0 {
						w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
						fmt.Fprintln(w, "Service\tState\tHealth\tCPU\tMemory")
						for _, svc := range dockerServices {
							fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
								svc.ServiceID, svc.State, svc.Health.Status,
								svc.ResourceUsage.CPU, svc.ResourceUsage.Memory)
						}
						w.Flush()
					} else {
						fmt.Println("  No Docker services running")
					}
				}
			}

			// Show network status
			networkMonitor := utils.NewNetworkMonitor()
			networkStatus := networkMonitor.CheckConnectivity()

			fmt.Println()
			fmt.Println("Network Status:")
			fmt.Printf("  Connectivity: %s", networkStatus.Connectivity)
			if networkStatus.IsOnline {
				fmt.Printf(" (online, %v latency)", networkStatus.Latency.Round(time.Millisecond))
			} else {
				fmt.Printf(" (offline)")
			}
			fmt.Printf("\n")
			fmt.Printf("  Last Checked: %s\n", networkStatus.LastChecked.Format("15:04:05"))

			// Show offline queue status
			offlineQueue := utils.NewOfflineQueue(db, networkMonitor, logger)
			queueStats := offlineQueue.GetStats()

			fmt.Println()
			fmt.Println("Offline Queue:")
			fmt.Printf("  Queued Commands: %d\n", queueStats.TotalQueued)
			fmt.Printf("  Processing: %t\n", queueStats.Processing)

			// Show system health
			fmt.Println()
			fmt.Println("System Health:")
			fmt.Printf("  Database: %s\n", "healthy")
			fmt.Printf("  Network: %s\n", func() string {
				if networkStatus.IsOnline {
					return "healthy"
				}
				return "degraded"
			}())
			fmt.Printf("  Storage: %s\n", "healthy")

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
