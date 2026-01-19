package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	whatsapp_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/whatsapp"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

func newWhatsAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whatsapp",
		Short: "Manage WhatsApp connectivity",
		Long:  `Login to WhatsApp, sync messages, send/receive messages, and manage WhatsApp connectivity.`,
	}

	cmd.AddCommand(newWhatsAppLoginCmd())
	cmd.AddCommand(newWhatsAppStatusCmd())
	cmd.AddCommand(newWhatsAppSyncCmd())
	cmd.AddCommand(newWhatsAppMessagesCmd())
	cmd.AddCommand(newWhatsAppContactsCmd())
	cmd.AddCommand(newWhatsAppChatsCmd())
	cmd.AddCommand(newWhatsAppSendCmd())
	cmd.AddCommand(newWhatsAppMediaCmd())

	return cmd
}

func newWhatsAppLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with WhatsApp (scan QR code)",
		Long: `Authenticate with WhatsApp by scanning a QR code with your phone.
Open WhatsApp → Settings → Linked Devices → Link a Device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppLogin()
		},
	}

	return cmd
}

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

func newWhatsAppSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync messages continuously (run until Ctrl+C)",
		Long: `Connect to WhatsApp and continuously sync messages to the local database.
This command:
- Downloads message history from WhatsApp servers
- Receives new messages in real-time
- Stores all messages in the database
- Runs until you press Ctrl+C`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppSync()
		},
	}

	return cmd
}

func newWhatsAppMessagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "List or search messages",
		Long:  `List messages from all chats or a specific chat, or search messages by content.`,
	}

	cmd.AddCommand(newWhatsAppMessagesListCmd())
	cmd.AddCommand(newWhatsAppMessagesSearchCmd())

	return cmd
}

func newWhatsAppMessagesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages",
		Long: `List messages from all chats or a specific chat.
Use --chat to filter by chat JID (e.g., 1234567890@s.whatsapp.net)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppMessagesList(cmd)
		},
	}

	cmd.Flags().String("chat", "", "Filter by chat JID")
	cmd.Flags().Int("limit", 20, "Maximum number of messages to return")
	cmd.Flags().Int("page", 0, "Page number for pagination (0-indexed)")

	return cmd
}

func newWhatsAppMessagesSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search messages",
		Long:  `Search messages by content across all chats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppMessagesSearch(cmd)
		},
	}

	cmd.Flags().String("query", "", "Search term (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().Int("page", 0, "Page number for pagination")

	return cmd
}

func newWhatsAppContactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage contacts",
		Long:  `Search and manage WhatsApp contacts.`,
	}

	cmd.AddCommand(newWhatsAppContactsSearchCmd())

	return cmd
}

func newWhatsAppContactsSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search contacts",
		Long:  `Search contacts by name or phone number.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppContactsSearch(cmd)
		},
	}

	cmd.Flags().String("query", "", "Search term for name or phone number (required)")

	return cmd
}

func newWhatsAppChatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "List chats",
		Long:  `List all chats sorted by recent activity.`,
	}

	cmd.AddCommand(newWhatsAppChatsListCmd())

	return cmd
}

func newWhatsAppChatsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List chats",
		Long:  `List all chats sorted by recent activity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppChatsList(cmd)
		},
	}

	cmd.Flags().String("query", "", "Filter chats by name or JID")
	cmd.Flags().Int("limit", 20, "Maximum number of chats")
	cmd.Flags().Int("page", 0, "Page number for pagination")

	return cmd
}

func newWhatsAppSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message",
		Long: `Send a text message to an individual or group.

Recipient formats:
  - Phone number: 1234567890
  - Individual JID: 1234567890@s.whatsapp.net
  - Group JID: 123456789@g.us`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppSend(cmd)
		},
	}

	cmd.Flags().String("to", "", "Recipient phone number or JID (required)")
	cmd.Flags().String("message", "", "Message text content (required)")

	return cmd
}

func newWhatsAppMediaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Manage media",
		Long:  `Download and manage media attachments.`,
	}

	cmd.AddCommand(newWhatsAppMediaDownloadCmd())

	return cmd
}

func newWhatsAppMediaDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download media for a message",
		Long: `Download media attachments (images, videos, audio, documents).

Use --message-id to specify which message's media to download.
Use --output to specify the destination file or directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhatsAppMediaDownload(cmd)
		},
	}

	cmd.Flags().String("message-id", "", "Message identifier (required)")
	cmd.Flags().String("chat", "", "Chat JID (optional, for disambiguation)")
	cmd.Flags().String("output", "", "Destination file or directory")

	return cmd
}

func runWhatsAppLogin() error {
	fmt.Println(headerStyle.Render("WhatsApp Login"))
	fmt.Println(strings.Repeat("=", 17))

	transport, err := getWhatsAppTransport()
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp transport: %w", err)
	}

	if transport.IsLoggedIn() {
		fmt.Println(successStyle.Render("[OK] Already authenticated with WhatsApp"))
		return nil
	}

	fmt.Println(infoStyle.Render("Please scan the QR code below with WhatsApp on your phone"))
	fmt.Println("   Open WhatsApp → Settings → Linked Devices → Link a Device")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := transport.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("[OK] Successfully authenticated with WhatsApp!"))
	fmt.Println("   You can now send and receive messages.")

	return nil
}

func runWhatsAppStatus() error {
	transport, err := getWhatsAppTransport()
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp transport: %w", err)
	}

	status, err := transport.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fmt.Println(headerStyle.Render("WhatsApp Status"))
	fmt.Println(strings.Repeat("=", 16))

	if status.Available {
		fmt.Println(successStyle.Render("[CONNECTED] Status: Connected"))
		fmt.Println(successStyle.Render("[OK] Authentication: Valid"))
	} else {
		fmt.Println(errorStyle.Render("[DISCONNECTED] Status: Disconnected"))
		if status.Error != "" {
			fmt.Printf("%s Error: %s\n", errorStyle.Render("[ERROR]"), status.Error)
		}
	}

	fmt.Printf("%s Last Checked: %s\n", dimStyle.Render("[TIME]"), status.LastChecked.Format("2006-01-02 15:04:05"))

	if !transport.IsLoggedIn() {
		fmt.Println()
		fmt.Printf("%s To login, run: %s\n", infoStyle.Render("[INFO]"), dimStyle.Render("mauritania-cli whatsapp login"))
	}

	return nil
}

func runWhatsAppSync() error {
	transport, err := getWhatsAppTransport()
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp transport: %w", err)
	}

	fmt.Println(headerStyle.Render("WhatsApp Sync"))
	fmt.Println(strings.Repeat("=", 14))

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	fmt.Println(infoStyle.Render("Starting sync... Press Ctrl+C to stop"))

	if err := transport.Sync(ctx); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}

func runWhatsAppMessagesList(cmd *cobra.Command) error {
	chatJID, _ := cmd.Flags().GetString("chat")
	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	messages, err := transport.ListMessages(&chatJID, nil, limit, page)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, messages, "")
}

func runWhatsAppMessagesSearch(cmd *cobra.Command) error {
	query, _ := cmd.Flags().GetString("query")
	if query == "" {
		return outputJSON(false, nil, "--query is required")
	}

	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	messages, err := transport.ListMessages(nil, &query, limit, page)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, messages, "")
}

func runWhatsAppContactsSearch(cmd *cobra.Command) error {
	query, _ := cmd.Flags().GetString("query")
	if query == "" {
		return outputJSON(false, nil, "--query is required")
	}

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	contacts, err := transport.SearchContacts(query)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, contacts, "")
}

func runWhatsAppChatsList(cmd *cobra.Command) error {
	query, _ := cmd.Flags().GetString("query")
	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	chats, err := transport.ListChats(&query, limit, page)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, chats, "")
}

func runWhatsAppSend(cmd *cobra.Command) error {
	recipient, _ := cmd.Flags().GetString("to")
	message, _ := cmd.Flags().GetString("message")

	if recipient == "" {
		return outputJSON(false, nil, "--to is required")
	}
	if message == "" {
		return outputJSON(false, nil, "--message is required")
	}

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	resp, err := transport.SendMessage(recipient, message)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, resp, "")
}

func runWhatsAppMediaDownload(cmd *cobra.Command) error {
	messageID, _ := cmd.Flags().GetString("message-id")
	chatJID, _ := cmd.Flags().GetString("chat")
	outputPath, _ := cmd.Flags().GetString("output")

	if messageID == "" {
		return outputJSON(false, nil, "--message-id is required")
	}

	transport, err := getWhatsAppTransport()
	if err != nil {
		return outputJSON(false, nil, fmt.Sprintf("failed to get WhatsApp transport: %w", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var chatPtr *string
	if chatJID != "" {
		chatPtr = &chatJID
	}

	resp, err := transport.DownloadMedia(ctx, messageID, chatPtr, outputPath)
	if err != nil {
		return outputJSON(false, nil, err.Error())
	}

	return outputJSON(true, resp, "")
}

func outputJSON(success bool, data interface{}, errMsg string) error {
	type jsonResult struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Error   *string     `json:"error"`
	}

	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}

	result := jsonResult{
		Success: success,
		Data:    data,
		Error:   errPtr,
	}

	b, _ := json.Marshal(result)
	fmt.Println(string(b))

	return nil
}

func getWhatsAppTransport() (*whatsapp_transport.WhatsAppTransport, error) {
	config := &utils.Config{}
	logger := log.New(os.Stderr, "", log.LstdFlags)

	transport, err := whatsapp_transport.NewWhatsAppTransport(config, logger)
	if err != nil {
		return nil, err
	}

	return transport, nil
}
