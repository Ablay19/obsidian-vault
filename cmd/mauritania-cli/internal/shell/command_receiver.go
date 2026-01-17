package shell

import (
	"fmt"
	"log"
	"strings"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// CommandReceiver handles incoming social media messages and converts them to commands
type CommandReceiver struct {
	db             *database.DB
	config         *utils.Config
	commandParser  *utils.CommandParser
	messageHandler *utils.MessageHandler
	logger         *log.Logger
	messageQueue   *utils.MessageQueue
}

// NewCommandReceiver creates a new command receiver
func NewCommandReceiver(db *database.DB, config *utils.Config, logger *log.Logger) *CommandReceiver {
	return &CommandReceiver{
		db:             db,
		config:         config,
		commandParser:  utils.NewCommandParser(),
		messageHandler: utils.NewMessageHandler(utils.NewLogger("command-receiver")),
		logger:         logger,
		messageQueue:   utils.NewMessageQueue(utils.NewLogger("message-queue")),
	}
}

// ProcessIncomingMessage processes an incoming social media message
func (cr *CommandReceiver) ProcessIncomingMessage(message *models.IncomingMessage) (*models.Command, error) {
	cr.logger.Printf("Processing incoming message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])

	// Clean the message content
	content := strings.TrimSpace(message.Content)
	if content == "" {
		return nil, fmt.Errorf("empty message content")
	}

	// Check if this is part of a multi-part message
	chunks, isComplete := cr.messageQueue.AddChunk(message.SenderID, content)

	if !isComplete {
		// Message chunks are still being collected
		cr.logger.Printf("Waiting for more message chunks from %s", message.SenderID)
		return nil, nil // Not an error, just incomplete
	}

	// Combine chunks if we have multiple parts
	var finalContent string
	if len(chunks) > 1 {
		finalContent = cr.messageHandler.CombineMessage(chunks)
		cr.logger.Printf("Combined %d message chunks into final content", len(chunks))
	} else {
		finalContent = chunks[0]
	}

	// Parse the command
	command, err := cr.commandParser.ParseCommand(finalContent)
	if err != nil {
		cr.logger.Printf("Failed to parse command: %v", err)
		return nil, fmt.Errorf("invalid command format: %w", err)
	}

	// Validate the command
	if err := cr.commandParser.ValidateCommand(command); err != nil {
		cr.logger.Printf("Command validation failed: %v", err)
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Create the full command record
	cmd := &models.SocialMediaCommand{
		ID:          generateCommandID(message),
		SenderID:    message.SenderID,
		Platform:    message.Transport,
		Command:     command.Command,
		Priority:    command.Priority,
		Status:      models.StatusReceived,
		Timestamp:   message.Timestamp,
		TransportID: message.ID,
	}

	// Validate sender permissions
	if err := cr.validateSenderPermissions(cmd.SenderID, cmd.Platform); err != nil {
		cr.logger.Printf("Sender permission validation failed: %v", err)
		return nil, fmt.Errorf("unauthorized sender: %w", err)
	}

	cr.logger.Printf("Successfully processed command: %s from %s", cmd.ID, cmd.SenderID)
	return cmd, nil
}

// ProcessWebhookMessage processes a message received via webhook
func (cr *CommandReceiver) ProcessWebhookMessage(platform string, payload []byte) ([]*models.Command, error) {
	// This would delegate to the appropriate transport's webhook processor
	// For now, return empty slice
	cr.logger.Printf("Processing webhook message for platform: %s", platform)
	return []*models.Command{}, nil
}

// ValidateAndQueueCommand validates a command and queues it for execution
func (cr *CommandReceiver) ValidateAndQueueCommand(cmd *models.Command) error {
	// Additional validation beyond basic parsing
	if err := cr.performSecurityChecks(cmd); err != nil {
		return fmt.Errorf("security check failed: %w", err)
	}

	// Set status to queued
	cmd.Status = models.StatusQueued

	// Save to database
	if err := cr.db.SaveCommand(*cmd); err != nil {
		cr.logger.Printf("Failed to save command to database: %v", err)
		return fmt.Errorf("failed to queue command: %w", err)
	}

	cr.logger.Printf("Command queued for execution: %s", cmd.ID)
	return nil
}

// validateSenderPermissions validates that the sender is authorized
func (cr *CommandReceiver) validateSenderPermissions(senderID, platform string) error {
	// Check against configured allowed users
	authConfig := cr.config.Auth

	// If no restrictions configured, allow all
	if len(authConfig.AllowedUsers) == 0 {
		return nil
	}

	// Check if sender is in allowed list
	for _, allowedUser := range authConfig.AllowedUsers {
		if allowedUser == senderID {
			return nil
		}
	}

	return fmt.Errorf("sender %s not in allowed users list", senderID)
}

// performSecurityChecks performs additional security validations
func (cr *CommandReceiver) performSecurityChecks(cmd *models.Command) error {
	// Rate limiting check (per user)
	// This would integrate with a user-specific rate limiter

	// Command allowlist check
	if !cr.commandParser.IsAllowedCommand(strings.Fields(cmd.Command)[0]) {
		return fmt.Errorf("command not in allowlist: %s", strings.Fields(cmd.Command)[0])
	}

	// Length and complexity checks
	if len(cmd.Command) > 10000 { // Reasonable upper bound
		return fmt.Errorf("command too long: %d characters", len(cmd.Command))
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"rm -rf /",
		"sudo ",
		"chmod 777",
		"dd if=",
		"mkfs",
		"fdisk",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmd.Command, pattern) {
			return fmt.Errorf("command contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// GetPendingCommands returns commands that are ready for execution
func (cr *CommandReceiver) GetPendingCommands() ([]models.Command, error) {
	commands, err := cr.db.GetPendingCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending commands: %w", err)
	}

	return commands, nil
}

// CleanupOldMessages cleans up old incomplete message chunks
func (cr *CommandReceiver) CleanupOldMessages() {
	cr.messageQueue.Cleanup()
	cr.logger.Printf("Cleaned up old message chunks")
}

// generateCommandID generates a unique command ID
func generateCommandID(message *models.IncomingMessage) string {
	return fmt.Sprintf("cmd_%s_%d", message.SenderID, message.Timestamp.Unix())
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
