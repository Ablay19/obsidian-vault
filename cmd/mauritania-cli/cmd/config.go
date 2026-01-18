package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/ui"
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
	cmd.AddCommand(newConfigSetupCmd())

	return cmd
}

// newConfigSetupCmd creates the setup command
func newConfigSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Long:  `Launch an interactive setup wizard to configure WhatsApp and Telegram transports.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWizard()
		},
	}

	return cmd
}

func runSetupWizard() error {
	fmt.Println("🤖 Mauritania CLI - Transport Setup Wizard")
	fmt.Println("==========================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Check current configuration
	fmt.Println("📋 Current Configuration Status:")
	fmt.Println("• WhatsApp: Not configured")
	fmt.Println("• Telegram: Not configured")
	fmt.Println()

	// Setup WhatsApp
	fmt.Println("🔧 WhatsApp Business API Setup")
	fmt.Println("------------------------------")
	fmt.Print("Do you want to configure WhatsApp? (y/n): ")
	whatsappChoice, _ := reader.ReadString('\n')
	whatsappChoice = strings.TrimSpace(strings.ToLower(whatsappChoice))

	if whatsappChoice == "y" || whatsappChoice == "yes" {
		if err := setupWhatsApp(reader); err != nil {
			fmt.Printf("❌ WhatsApp setup failed: %v\n", err)
		} else {
			fmt.Println("✅ WhatsApp configured successfully!")
		}
		fmt.Println()
	}

	// Setup Telegram
	fmt.Println("🔧 Telegram Bot Setup")
	fmt.Println("---------------------")
	fmt.Print("Do you want to configure Telegram? (y/n): ")
	telegramChoice, _ := reader.ReadString('\n')
	telegramChoice = strings.TrimSpace(strings.ToLower(telegramChoice))

	if telegramChoice == "y" || telegramChoice == "yes" {
		if err := setupTelegram(reader); err != nil {
			fmt.Printf("❌ Telegram setup failed: %v\n", err)
		} else {
			fmt.Println("✅ Telegram configured successfully!")
		}
		fmt.Println()
	}

	fmt.Println("🎉 Setup complete!")
	fmt.Println("Run 'mauritania-cli config show' to verify your configuration.")
	fmt.Println("Run 'mauritania-cli status' to check transport connectivity.")

	return nil
}

func setupWhatsApp(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("WhatsApp Business API Configuration:")
	fmt.Println("-----------------------------------")
	fmt.Println("You'll need:")
	fmt.Println("1. WhatsApp Business Account")
	fmt.Println("2. Access Token from Facebook Developers")
	fmt.Println("3. Phone Number ID from WhatsApp Business API")
	fmt.Println()

	fmt.Print("Enter WhatsApp Access Token: ")
	accessToken, _ := reader.ReadString('\n')
	accessToken = strings.TrimSpace(accessToken)

	if accessToken == "" {
		return fmt.Errorf("access token is required")
	}

	fmt.Print("Enter Phone Number ID: ")
	phoneNumberID, _ := reader.ReadString('\n')
	phoneNumberID = strings.TrimSpace(phoneNumberID)

	if phoneNumberID == "" {
		return fmt.Errorf("phone number ID is required")
	}

	fmt.Print("Enter Webhook Secret (optional, press Enter to skip): ")
	webhookSecret, _ := reader.ReadString('\n')
	webhookSecret = strings.TrimSpace(webhookSecret)

	// Here you would save to config file
	// For now, just display what would be configured
	fmt.Println()
	fmt.Println("📝 Configuration to be saved:")
	fmt.Printf("  Access Token: %s...%s\n", accessToken[:10], accessToken[len(accessToken)-4:])
	fmt.Printf("  Phone Number ID: %s\n", phoneNumberID)
	if webhookSecret != "" {
		fmt.Printf("  Webhook Secret: %s...%s\n", webhookSecret[:4], webhookSecret[len(webhookSecret)-4:])
	}

	// TODO: Save to config file
	fmt.Println("⚠️  Config file saving not yet implemented in setup wizard")
	fmt.Println("Please manually edit ~/.mauritania-cli.toml with the values above")

	return nil
}

func setupTelegram(reader *bufio.Reader) error {
	fmt.Println()
	fmt.Println("Telegram Bot Configuration:")
	fmt.Println("--------------------------")
	fmt.Println("You'll need:")
	fmt.Println("1. Telegram Bot Token from @BotFather")
	fmt.Println("2. Chat ID (optional, for restricting to specific chats)")
	fmt.Println()

	fmt.Print("Enter Telegram Bot Token: ")
	botToken, _ := reader.ReadString('\n')
	botToken = strings.TrimSpace(botToken)

	if botToken == "" {
		return fmt.Errorf("bot token is required")
	}

	fmt.Print("Enter Allowed Chat ID (optional, press Enter to allow all): ")
	chatID, _ := reader.ReadString('\n')
	chatID = strings.TrimSpace(chatID)

	// Here you would save to config file
	fmt.Println()
	fmt.Println("📝 Configuration to be saved:")
	fmt.Printf("  Bot Token: %s...%s\n", botToken[:10], botToken[len(botToken)-4:])
	if chatID != "" {
		fmt.Printf("  Chat ID: %s\n", chatID)
	} else {
		fmt.Println("  Chat ID: (not restricted)")
	}

	// TODO: Save to config file
	fmt.Println("⚠️  Config file saving not yet implemented in setup wizard")
	fmt.Println("Please manually edit ~/.mauritania-cli.toml with the values above")

	return nil
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

			// Pretty print configuration with styled output
			ui.Header("📋 Current Configuration")

			// Database section
			dbContent := fmt.Sprintf("Type: %s\n", config.Database.Type)
			if config.Database.Type == "sqlite" {
				dbContent += fmt.Sprintf("Path: %s\n", config.Database.Path)
			}
			ui.Println(ui.InfoBox("Database", strings.TrimSuffix(dbContent, "\n")))

			// Transports section
			transportContent := fmt.Sprintf("Default: %s\n\n", config.Transports.Default)
			transportContent += "Social Media:\n"
			transportContent += fmt.Sprintf("  WhatsApp: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.whatsapp.database_path")))
			transportContent += fmt.Sprintf("  Telegram: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.telegram.bot_token")))
			transportContent += fmt.Sprintf("  Facebook: %s\n", boolToConfigStatus(cm.IsSet("transports.social_media.facebook.access_token")))
			transportContent += fmt.Sprintf("\nShipper: %s", boolToConfigStatus(cm.IsSet("transports.shipper.api_key")))
			ui.Println(ui.InfoBox("Transports", transportContent))

			// Network section
			networkContent := fmt.Sprintf("Timeout: %d seconds\n", config.Network.Timeout)
			networkContent += fmt.Sprintf("Retry Attempts: %d\n", config.Network.RetryAttempts)
			networkContent += fmt.Sprintf("Offline Mode: %t", config.Network.OfflineMode)
			ui.Println(ui.InfoBox("Network", networkContent))

			// Logging section
			loggingContent := fmt.Sprintf("Level: %s\n", config.Logging.Level)
			loggingContent += fmt.Sprintf("File: %s", config.Logging.File)
			ui.Println(ui.InfoBox("Logging", loggingContent))

			// Authentication section
			authContent := "Enabled: "
			if config.Auth.Enabled {
				authContent += "Yes"
			} else {
				authContent += "No"
			}
			ui.Println(ui.InfoBox("Authentication", authContent))
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
