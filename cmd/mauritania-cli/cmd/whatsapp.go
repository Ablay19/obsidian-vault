package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	whatsapp_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/whatsapp"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// newWhatsAppCmd creates the whatsapp command
func newWhatsAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whatsapp",
		Short: "Manage WhatsApp connectivity",
		Long:  `Login to WhatsApp, check status, and manage WhatsApp connectivity.`,
	}

	// Add subcommands
	cmd.AddCommand(newWhatsAppLoginCmd())
	cmd.AddCommand(newWhatsAppStatusCmd())

	return cmd
}

// newWhatsAppLoginCmd creates the whatsapp login command
func newWhatsAppLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to WhatsApp using QR code",
		Long:  `Authenticate with WhatsApp by scanning a QR code with your phone.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppLogin()
		},
	}

	return cmd
}

// newWhatsAppStatusCmd creates the whatsapp status command
func newWhatsAppStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check WhatsApp connection status",
		Long:  `Display the current WhatsApp connection and authentication status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppStatus()
		},
	}

	return cmd
}

func runWhatsAppLogin() error {
	fmt.Println("🔐 WhatsApp Login")
	fmt.Println("=================")

	// Get WhatsApp transport
	transport, err := getWhatsAppTransport()
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp transport: %w", err)
	}

	// Check if already logged in
	if transport.IsLoggedIn() {
		fmt.Println("✅ Already logged in to WhatsApp!")
		return nil
	}

	fmt.Println("📱 Please scan the QR code below with WhatsApp on your phone")
	fmt.Println("   Open WhatsApp → Linked Devices → Link a Device")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := transport.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println()
	fmt.Println("🎉 Successfully logged in to WhatsApp!")
	fmt.Println("   You can now send and receive messages.")
	return nil
}

func runWhatsAppStatus() error {
	// Get WhatsApp transport
	transport, err := getWhatsAppTransport()
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp transport: %w", err)
	}

	status, err := transport.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println("📱 WhatsApp Status")
	fmt.Println("==================")

	if status.Available {
		fmt.Println("✅ Status: Connected")
		fmt.Println("✅ Authentication: Valid")
	} else {
		fmt.Println("❌ Status: Disconnected")
		if status.Error != "" {
			fmt.Printf("❌ Error: %s\n", status.Error)
		}
	}

	fmt.Printf("🕒 Last Checked: %s\n", status.LastChecked.Format("2006-01-02 15:04:05"))

	if !transport.IsLoggedIn() {
		fmt.Println()
		fmt.Println("💡 To login, run: mauritania-cli whatsapp login")
	}

	return nil
}

// getWhatsAppTransport gets the WhatsApp transport instance
func getWhatsAppTransport() (*whatsapp_transport.WhatsAppTransport, error) {
	// This is a simplified version - in reality you'd get it from the app context
	// For now, we'll create a new instance
	config := &utils.Config{}
	logger := log.New(os.Stderr, "", log.LstdFlags)

	transport, err := whatsapp_transport.NewWhatsAppTransport(config, logger)
	if err != nil {
		return nil, err
	}

	return transport, nil
}
