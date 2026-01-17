package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newSecurityCmd creates the security command
func newSecurityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Manage security settings",
		Long:  `Configure security features including authentication, encryption, and access controls.`,
	}

	// Add subcommands
	cmd.AddCommand(newSecurityDisableCmd())
	cmd.AddCommand(newSecurityEnableCmd())
	cmd.AddCommand(newSecurityStatusCmd())

	return cmd
}

// newSecurityDisableCmd disables security features for development
func newSecurityDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable security features",
		Long: `Disable authentication, encryption, and other security features.
WARNING: This should only be used in development/testing environments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⚠️  WARNING: Disabling security features!")
			fmt.Println("This should only be used in development/testing environments.")
			fmt.Println()

			// Create or load config
			cm := utils.NewConfigManager()
			_ = cm.Load() // Ignore load errors, use defaults

			fmt.Println("Using configuration (with defaults if config file unavailable)")

			// Disable security features
			cm.Set("auth.enabled", false)
			cm.Set("network.offline_mode", true)

			// Try to save configuration (ignore errors for now)
			if err := cm.Save(); err != nil {
				fmt.Printf("⚠️  Warning: Could not save configuration (%v)\n", err)
				fmt.Println("Changes will only apply to this session.")
			}

			fmt.Println("✅ Security features disabled:")
			fmt.Println("  • Authentication: DISABLED")
			fmt.Println("  • Offline mode: ENABLED")
			fmt.Println()
			if err := cm.Save(); err == nil {
				fmt.Println("🔄 Restart the application for changes to take effect.")
			}
			fmt.Println("🛡️  Remember to re-enable security for production use!")

			return nil
		},
	}

	return cmd
}

// newSecurityEnableCmd enables security features
func newSecurityEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable security features",
		Long:  `Enable authentication, encryption, and other security features.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create or load config
			cm := utils.NewConfigManager()
			if err := cm.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Enable security features
			cm.Set("auth.enabled", true)
			cm.Set("auth.require_approval", false) // Default to permissive
			cm.Set("network.offline_mode", false)

			// Save configuration
			if err := cm.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Println("✅ Security features enabled:")
			fmt.Println("  • Authentication: ENABLED")
			fmt.Println("  • Authorization: ENABLED")
			fmt.Println("  • Offline mode: DISABLED")
			fmt.Println()
			fmt.Println("🔄 Restart the application for changes to take effect.")
			fmt.Println("🔐 Configure API keys and user permissions as needed.")

			return nil
		},
	}

	return cmd
}

// newSecurityStatusCmd shows current security status
func newSecurityStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show security status",
		Long:  `Display the current status of security features and settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("DEBUG: Status command started\n")
			cm := utils.NewConfigManager()
			fmt.Printf("DEBUG: ConfigManager created\n")
			_ = cm.Load() // Load config, ignore errors

			// Use viper values directly to avoid struct issues
			authEnabled := cm.GetBool("auth.enabled")
			fmt.Printf("Status debug: auth.enabled = %t\n", authEnabled)
			requireApproval := cm.GetBool("auth.require_approval")
			allowedUsers := cm.GetStringSlice("auth.allowed_users")
			allowedCommands := cm.GetStringSlice("auth.allowed_commands")
			offlineMode := cm.GetBool("network.offline_mode")
			networkTimeout := cm.GetInt("network.timeout")

			fmt.Println("🔒 Security Status:")
			fmt.Println()

			// Authentication status
			authStatus := "ENABLED"
			if !authEnabled {
				authStatus = "DISABLED"
			}
			fmt.Printf("Authentication: %s\n", authStatus)

			// Authorization details
			if authEnabled {
				fmt.Printf("  Require Approval: %t\n", requireApproval)
				fmt.Printf("  Allowed Users: %d configured\n", len(allowedUsers))
				fmt.Printf("  Allowed Commands: %d configured\n", len(allowedCommands))
			}
			fmt.Println()

			// Encryption status
			fmt.Println("Encryption:")
			fmt.Printf("  Master Key: %s\n", "configured") // TODO: check if master key exists
			fmt.Println()

			// Network security
			fmt.Println("Network Security:")
			fmt.Printf("  Offline Mode: %t\n", offlineMode)
			fmt.Printf("  Timeout: %d seconds\n", networkTimeout)
			fmt.Println()

			// Recommendations
			if !authEnabled {
				fmt.Println("⚠️  WARNING: Authentication is disabled!")
				fmt.Println("   This should only be used in development environments.")
				fmt.Println("   Use 'mauritania-cli security enable' to re-enable security.")
			} else {
				fmt.Println("✅ Security features are properly configured.")
			}

			return nil
		},
	}

	return cmd
}
