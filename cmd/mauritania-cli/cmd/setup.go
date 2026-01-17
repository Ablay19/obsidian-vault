package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// newConfigSetupCmd creates the config setup command
func newConfigSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard for configuring transports",
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
