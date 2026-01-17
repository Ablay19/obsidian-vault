package shell

import (
	"fmt"
	"log"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	facebook_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/facebook"
	telegram_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/telegram"
	whatsapp_transport "obsidian-automation/cmd/mauritania-cli/internal/transports/whatsapp"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// TransportClient defines the interface that all transport clients must implement
type TransportClient interface {
	// SendMessage sends a message via the transport
	SendMessage(recipient, message string) (*models.MessageResponse, error)

	// ReceiveMessage polls for new messages (webhook-based transports may not need this)
	ReceiveMessages() ([]*models.IncomingMessage, error)

	// GetStatus returns the current status of the transport
	GetStatus() (*models.TransportStatus, error)

	// ValidateCredentials validates that the transport credentials are working
	ValidateCredentials() error

	// GetRateLimit returns current rate limiting status
	GetRateLimit() (*models.RateLimit, error)
}

// ResponseSender handles sending command results back via social media
type ResponseSender struct {
	db             *database.DB
	config         *utils.Config
	transports     map[string]TransportClient
	messageHandler *utils.MessageHandler
	logger         *log.Logger
}

// NewResponseSender creates a new response sender
func NewResponseSender(db *database.DB, config *utils.Config, logger *log.Logger) *ResponseSender {
	rs := &ResponseSender{
		db:             db,
		config:         config,
		transports:     make(map[string]TransportClient),
		messageHandler: utils.NewMessageHandler(utils.NewLogger("response-sender")),
		logger:         logger,
	}

	// Initialize available transports
	rs.initializeTransports()

	return rs
}

// initializeTransports sets up available transport clients
func (rs *ResponseSender) initializeTransports() {
	// WhatsApp transport
	if waTransport, err := whatsapp_transport.NewWhatsAppTransport(rs.config, rs.logger); err == nil {
		rs.transports["whatsapp"] = waTransport
		rs.logger.Printf("WhatsApp transport initialized")
	} else {
		rs.logger.Printf("Failed to initialize WhatsApp transport: %v", err)
	}

	// Telegram transport
	if tgTransport, err := telegram_transport.NewTelegramTransport(rs.config, rs.logger); err == nil {
		rs.transports["telegram"] = tgTransport
		rs.logger.Printf("Telegram transport initialized")
	} else {
		rs.logger.Printf("Failed to initialize Telegram transport: %v", err)
	}

	// Facebook transport
	if fbTransport, err := facebook_transport.NewFacebookTransport(rs.config, rs.logger); err == nil {
		rs.transports["facebook"] = fbTransport
		rs.logger.Printf("Facebook transport initialized")
	} else {
		rs.logger.Printf("Failed to initialize Facebook transport: %v", err)
	}
}

// SendCommandResult sends a command result back to the sender
func (rs *ResponseSender) SendCommandResult(result *models.CommandResult) error {
	rs.logger.Printf("Sending command result for command %s", result.CommandID)

	// Get the original command to find sender information
	cmd, err := rs.db.GetCommand(result.CommandID)
	if err != nil {
		return fmt.Errorf("failed to get command %s: %w", result.CommandID, err)
	}

	// Get the transport client for this platform
	transport, exists := rs.transports[cmd.Platform]
	if !exists {
		return fmt.Errorf("no transport available for platform: %s", cmd.Platform)
	}

	// Format the response message
	responseMessage := rs.formatCommandResult(cmd, result)

	// Handle large responses by splitting into chunks
	chunks := rs.messageHandler.SplitMessage(responseMessage)

	rs.logger.Printf("Sending %d message chunks for command %s", len(chunks), result.CommandID)

	// Send each chunk
	for i, chunk := range chunks {
		if len(chunks) > 1 {
			rs.logger.Printf("Sending chunk %d/%d", i+1, len(chunks))
		}

		_, err := transport.SendMessage(cmd.SenderID, chunk)
		if err != nil {
			rs.logger.Printf("Failed to send chunk %d: %v", i+1, err)
			return fmt.Errorf("failed to send response chunk %d: %w", i+1, err)
		}

		// Small delay between chunks to avoid rate limiting
		if i < len(chunks)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Update command status to responded
	cmd.Status = models.StatusResponded
	if err := rs.db.SaveCommand(*cmd); err != nil {
		rs.logger.Printf("Failed to update command status: %v", err)
		// Don't return error as the response was sent successfully
	}

	rs.logger.Printf("Successfully sent response for command %s", result.CommandID)
	return nil
}

// SendErrorMessage sends an error message back to the sender
func (rs *ResponseSender) SendErrorMessage(commandID, errorMessage string) error {
	rs.logger.Printf("Sending error message for command %s: %s", commandID, errorMessage)

	// Get the original command
	cmd, err := rs.db.GetCommand(commandID)
	if err != nil {
		return fmt.Errorf("failed to get command %s: %w", commandID, err)
	}

	// Get the transport client
	transport, exists := rs.transports[cmd.Platform]
	if !exists {
		return fmt.Errorf("no transport available for platform: %s", cmd.Platform)
	}

	// Format error message
	errorResponse := fmt.Sprintf("❌ Command execution failed:\n\n%s", errorMessage)

	// Send the error message
	_, err = transport.SendMessage(cmd.SenderID, errorResponse)
	if err != nil {
		return fmt.Errorf("failed to send error message: %w", err)
	}

	// Update command status
	cmd.Status = models.StatusResponded
	if err := rs.db.SaveCommand(*cmd); err != nil {
		rs.logger.Printf("Failed to update command status: %v", err)
	}

	rs.logger.Printf("Sent error message for command %s", commandID)
	return nil
}

// SendStatusUpdate sends a status update for a running command
func (rs *ResponseSender) SendStatusUpdate(commandID string, status string, progress int) error {
	rs.logger.Printf("Sending status update for command %s: %s (%d%%)", commandID, status, progress)

	// Get the original command
	cmd, err := rs.db.GetCommand(commandID)
	if err != nil {
		return fmt.Errorf("failed to get command %s: %w", commandID, err)
	}

	// Get the transport client
	transport, exists := rs.transports[cmd.Platform]
	if !exists {
		return fmt.Errorf("no transport available for platform: %s", cmd.Platform)
	}

	// Format status message
	var emoji string
	switch status {
	case "executing":
		emoji = "⚙️"
	case "completed":
		emoji = "✅"
	case "failed":
		emoji = "❌"
	default:
		emoji = "ℹ️"
	}

	statusMessage := fmt.Sprintf("%s Command Status: %s", emoji, status)
	if progress > 0 {
		statusMessage += fmt.Sprintf(" (%d%%)", progress)
	}

	// Send the status update
	_, err = transport.SendMessage(cmd.SenderID, statusMessage)
	if err != nil {
		return fmt.Errorf("failed to send status update: %w", err)
	}

	rs.logger.Printf("Sent status update for command %s", commandID)
	return nil
}

// formatCommandResult formats a command result into a readable message
func (rs *ResponseSender) formatCommandResult(cmd *models.Command, result *models.CommandResult) string {
	var response strings.Builder

	// Add header
	response.WriteString("```\n")
	response.WriteString("$ ")
	response.WriteString(cmd.Command)
	response.WriteString("\n```\n\n")

	// Add execution info
	response.WriteString(fmt.Sprintf("⏱️ **Execution Time:** %d ms\n", result.ExecutionTime))
	response.WriteString(fmt.Sprintf("📊 **Exit Code:** %d\n\n", result.ExitCode))

	// Add output
	if result.Status == "success" {
		response.WriteString("✅ **Output:**\n")
		response.WriteString("```\n")
		if len(result.Stdout) > 0 {
			// Truncate very long output
			output := result.Stdout
			if len(output) > 2000 {
				output = output[:2000] + "\n... (output truncated)"
			}
			response.WriteString(output)
		} else {
			response.WriteString("(no output)")
		}
		response.WriteString("\n```\n")
	} else {
		response.WriteString("❌ **Error:**\n")
		response.WriteString("```\n")
		if len(result.Stderr) > 0 {
			errorOutput := result.Stderr
			if len(errorOutput) > 2000 {
				errorOutput = errorOutput[:2000] + "\n... (error output truncated)"
			}
			response.WriteString(errorOutput)
		} else {
			response.WriteString(result.Status)
		}
		response.WriteString("\n```\n")
	}

	// Add cost information if available
	if result.Cost > 0 {
		response.WriteString(fmt.Sprintf("\n💰 **Cost:** $%.4f\n", result.Cost))
	}

	return response.String()
}

// GetAvailableTransports returns a list of available transport platforms
func (rs *ResponseSender) GetAvailableTransports() []string {
	transports := make([]string, 0, len(rs.transports))
	for platform := range rs.transports {
		transports = append(transports, platform)
	}
	return transports
}

// TestTransportConnection tests if a transport is working
func (rs *ResponseSender) TestTransportConnection(platform string) error {
	transport, exists := rs.transports[platform]
	if !exists {
		return fmt.Errorf("transport not available: %s", platform)
	}

	_, err := transport.GetStatus()
	return err
}

// SendWelcomeMessage sends a welcome message to a new user
func (rs *ResponseSender) SendWelcomeMessage(platform, recipientID string) error {
	transport, exists := rs.transports[platform]
	if !exists {
		return fmt.Errorf("transport not available: %s", platform)
	}

	welcomeMessage := `🤖 **Welcome to Mauritania CLI!**

You can now execute commands remotely using social media.

**Available commands:**
• ` + "`ls`, `pwd`, `cat`, `head`, `tail`, `grep`, `find`" + `
• ` + "`git status`, `git add`, `git commit`" + `
• ` + "`npm install`, `npm run build`" + `

**Examples:**
` + "`ls -la`" + `
` + "`git status`" + `
` + "`npm install && npm run build`" + `

**Note:** All commands are executed in a safe environment with restrictions on dangerous operations.`

	_, err := transport.SendMessage(recipientID, welcomeMessage)
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	rs.logger.Printf("Sent welcome message to %s on %s", recipientID, platform)
	return nil
}
