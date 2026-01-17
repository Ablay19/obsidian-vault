package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newSendCmd creates the send command
func newSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [command]",
		Short: "Send a command via network transport",
		Long: `Send a command to be executed through the configured network transport
(social media, SM APOS Shipper, or NRT routing). The command will be
queued for execution when connectivity allows.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]
			transportFlag, _ := cmd.Flags().GetString("transport")
			priorityFlag, _ := cmd.Flags().GetString("priority")
			platformFlag, _ := cmd.Flags().GetString("platform")
			offlineMode, _ := cmd.Flags().GetBool("offline")

			// Initialize logger
			logger := utils.NewLogger("SEND")

			// Detect platform
			platform := utils.DetectPlatform()
			logger.Info("Platform detected", platform.Type)

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

			// Initialize network monitoring
			networkMonitor := utils.NewNetworkMonitor()
			networkMonitor.Start()
			defer networkMonitor.Stop()

			// Initialize offline queue
			offlineQueue := utils.NewOfflineQueue(db, networkMonitor, log.New(os.Stderr, "", 0)) // Use simple logger for now
			offlineQueue.Start()
			defer offlineQueue.Stop()

			// Parse priority
			priority := parsePriority(priorityFlag)

			// Create command
			cmdModel := models.Command{
				ID:          uuid.New().String(),
				SenderID:    getSenderID(),
				Platform:    platformFlag,
				Command:     command,
				Timestamp:   time.Now(),
				Priority:    priority,
				Status:      models.StatusQueued,
				TransportID: "",
				SessionID:   "",
			}

			// Select transport
			selectedTransport := selectTransport(transportFlag, platform)
			logger.Info("Transport selected", "transport", selectedTransport)

			// Check network connectivity if not in offline mode
			if !offlineMode {
				networkStatus := networkMonitor.CheckConnectivity()
				fmt.Printf("Network status: online=%v, type=%s, latency=%v, error=%v\n",
					networkStatus.IsOnline, networkStatus.Connectivity, networkStatus.Latency, networkStatus.Error)

				if !networkStatus.IsOnline {
					logger.Warn("Network offline, queuing command for later retry")
					fmt.Printf("⚠️  Network offline - command will be queued for retry when connectivity returns\n")
				} else {
					logger.Info("Network connectivity established", map[string]any{
						"type":    networkStatus.Connectivity,
						"latency": networkStatus.Latency,
					})
				}
			}

			// Save command to database
			if err := db.SaveCommand(cmdModel); err != nil {
				return fmt.Errorf("failed to save command: %w", err)
			}

			// Attempt immediate execution if online and not offline mode
			if !offlineMode && networkMonitor.IsOnline() {
				if err := executeCommandWithRetry(cmdModel, networkMonitor, offlineQueue, log.New(os.Stderr, "", 0)); err != nil {
					logger.Error("Immediate execution failed, command queued for retry", err)
					if addErr := offlineQueue.AddFailedCommand(cmdModel, err); addErr != nil {
						logger.Error("Failed to add to offline queue", addErr)
					}
				}
			}

			// Log command history
			status := "success"
			if !networkMonitor.IsOnline() && !offlineMode {
				status = "queued_offline"
			}

			history := models.CommandHistory{
				ID:               uuid.New().String(),
				Timestamp:        time.Now(),
				Command:          "send",
				Args:             []string{command},
				User:             os.Getenv("USER"),
				Environment:      string(platform.Type),
				Status:           status,
				Duration:         0,
				AffectedServices: []string{},
			}

			if err := db.SaveCommandHistory(history); err != nil {
				logger.Warn("Failed to save command history", err)
			}

			// Display result
			networkStatus := networkMonitor.GetStatus()
			fmt.Printf("Command queued for execution:\n")
			fmt.Printf("  ID: %s\n", cmdModel.ID)
			fmt.Printf("  Command: %s\n", command)
			fmt.Printf("  Transport: %s\n", selectedTransport)
			fmt.Printf("  Priority: %s\n", priority)
			fmt.Printf("  Status: %s\n", cmdModel.Status)
			fmt.Printf("  Network: %s", networkStatus.Connectivity)
			if networkStatus.IsOnline {
				fmt.Printf(" (online, %v latency)", networkStatus.Latency.Round(time.Millisecond))
			} else {
				fmt.Printf(" (offline)")
			}
			fmt.Printf("\n")

			if !networkMonitor.IsOnline() && !offlineMode {
				fmt.Printf("💡 Tip: Command will be retried automatically when network connectivity returns\n")
			}

			return nil
		},
	}

	cmd.Flags().StringP("transport", "t", "auto", "Transport method (social_media, sm_apos, nrt, auto)")
	cmd.Flags().StringP("priority", "p", "normal", "Command priority (low, normal, high, urgent)")
	cmd.Flags().StringP("platform", "P", "unknown", "Target platform (whatsapp, telegram, facebook)")
	cmd.Flags().Bool("offline", false, "Force offline mode (don't attempt immediate execution)")

	return cmd
}

// parsePriority converts string priority to enum
func parsePriority(priority string) models.CommandPriority {
	switch priority {
	case "low":
		return models.PriorityLow
	case "high":
		return models.PriorityHigh
	case "urgent":
		return models.PriorityUrgent
	default:
		return models.PriorityNormal
	}
}

// selectTransport chooses the appropriate transport method
func selectTransport(requested string, platform *utils.PlatformInfo) string {
	if requested != "auto" {
		return requested
	}

	// Auto-selection logic
	if platform.IsTermux {
		// On mobile, prefer social media transports
		return "social_media"
	}

	// On desktop, prefer direct or NRT
	if platform.HasDocker {
		return "direct"
	}

	return "nrt"
}

// executeCommandWithRetry attempts to execute a command with retry logic
func executeCommandWithRetry(cmd models.Command, networkMonitor *utils.NetworkMonitor, offlineQueue *utils.OfflineQueue, logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	retryConfig := utils.DefaultRetryConfig()
	if networkMonitor.GetStatus().Connectivity == utils.ConnectivityMobile {
		retryConfig = utils.MobileRetryConfig()
	}

	return utils.RetryOperation(ctx, retryConfig, func() error {
		// Check network connectivity before each attempt
		if !networkMonitor.IsOnline() {
			return fmt.Errorf("network offline")
		}

		// In a real implementation, this would use the actual transport
		// For now, simulate transport execution
		log.Printf("Executing command %s via transport", cmd.ID)

		// Simulate network operation
		time.Sleep(200 * time.Millisecond)

		// Simulate occasional failures for testing
		if time.Now().UnixNano()%10 < 2 { // 20% failure rate
			return fmt.Errorf("simulated network failure")
		}

		// Mark command as completed
		cmd.Status = models.StatusCompleted
		// Note: Database update would happen here in real implementation

		log.Printf("Command %s executed successfully", cmd.ID)
		return nil
	})
}

// getSenderID generates a unique sender identifier
func getSenderID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	return fmt.Sprintf("%s@%s", user, hostname)
}
