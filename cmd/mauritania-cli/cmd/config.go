package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newConfigCmd creates the config command
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View and modify application configuration settings.`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigValidateCmd())

	return cmd
}

// newConfigShowCmd shows current configuration
func newConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current application configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := utils.NewConfigManager()
			if err := cm.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			config := cm.Get()
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(config)
			}

			// Pretty print configuration
			fmt.Println("📋 Current Configuration:")
			fmt.Println()

			fmt.Printf("Database:\n")
			fmt.Printf("  Type: %s\n", config.Database.Type)
			if config.Database.Type == "sqlite" {
				fmt.Printf("  Path: %s\n", config.Database.Path)
			}
			fmt.Println()

			fmt.Printf("Transports:\n")
			fmt.Printf("  Default: %s\n", config.Transports.Default)
			fmt.Printf("  Social Media:\n")
			fmt.Printf("    WhatsApp: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.whatsapp.api_key")))
			fmt.Printf("    Telegram: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.telegram.bot_token")))
			fmt.Printf("    Facebook: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.facebook.access_token")))
			fmt.Printf("  Shipper: %s\n", boolToConfigStatus(cm.IsSet("transports.shipper.api_key")))
			fmt.Println()

			fmt.Printf("Network:\n")
			fmt.Printf("  Timeout: %d seconds\n", config.Network.Timeout)
			fmt.Printf("  Retry Attempts: %d\n", config.Network.RetryAttempts)
			fmt.Printf("  Offline Mode: %t\n", config.Network.OfflineMode)
			fmt.Println()

			fmt.Printf("Logging:\n")
			fmt.Printf("  Level: %s\n", config.Logging.Level)
			fmt.Printf("  File: %s\n", config.Logging.File)
			fmt.Println()

			fmt.Printf("Authentication:\n")
			fmt.Printf("  Enabled: %t\n", config.Auth.Enabled)
			if config.Auth.Enabled {
				fmt.Printf("  Allowed Users: %d\n", len(config.Auth.AllowedUsers))
				fmt.Printf("  Require Approval: %t\n", config.Auth.RequireApproval)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "output configuration as JSON")

	return cmd
}

// newConfigInitCmd initializes a default configuration file
func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [file]",
		Short: "Initialize configuration file",
		Long:  `Create a default configuration file at the specified location or default path.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := "~/.mauritania-cli.toml"
			if len(args) > 0 {
				configPath = args[0]
			}

			// Expand home directory
			if configPath[:2] == "~/" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}
				configPath = homeDir + configPath[1:]
			}

			if err := utils.CreateDefaultConfig(configPath); err != nil {
				return fmt.Errorf("failed to create configuration file: %w", err)
			}

			fmt.Printf("✅ Configuration file created: %s\n", configPath)
			fmt.Println("You can now edit this file to configure your transports and settings.")
			fmt.Println("Use 'mauritania-cli config show' to view current configuration.")

			return nil
		},
	}

	return cmd
}

// newConfigValidateCmd validates the current configuration
func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  `Check the current configuration for errors and missing required settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := utils.NewConfigManager()
			if err := cm.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			errors := cm.ValidateConfig()

			if len(errors) == 0 {
				fmt.Println("✅ Configuration is valid")
				return nil
			}

			fmt.Println("❌ Configuration validation failed:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}

			fmt.Println()
			fmt.Println("Use 'mauritania-cli config init' to create a default configuration file,")
			fmt.Println("then edit it with your API keys and settings.")

			return fmt.Errorf("configuration validation failed with %d errors", len(errors))
		},
	}

	return cmd
}

// boolToConfigStatus converts boolean to configuration status string
func boolToConfigStatus(configured bool) string {
	if configured {
		return "configured"
	}
	return "not configured"
}
