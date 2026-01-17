package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// ShipperCommandExecutor handles command execution through SM APOS Shipper
type ShipperCommandExecutor struct {
	db                *database.DB
	sessionManager    *ShipperSessionManager
	transportSelector *TransportSelector
	encryption        *utils.CommandEncryption
	securityValidator *utils.CommandSecurityValidator
	logger            *log.Logger
	maxRetries        int
	retryDelay        time.Duration
}

// NewShipperCommandExecutor creates a new shipper command executor
func NewShipperCommandExecutor(db *database.DB, sessionManager *ShipperSessionManager, transportSelector *TransportSelector, config *utils.Config, logger *log.Logger) (*ShipperCommandExecutor, error) {
	defaultConfig := utils.DefaultEncryptionConfig()
	encryption := utils.NewCommandEncryption(&defaultConfig, logger)
	securityValidator := utils.NewCommandSecurityValidator(logger)

	return &ShipperCommandExecutor{
		db:                db,
		sessionManager:    sessionManager,
		transportSelector: transportSelector,
		encryption:        encryption,
		securityValidator: securityValidator,
		logger:            logger,
		maxRetries:        3,
		retryDelay:        2 * time.Second,
	}, nil
}

// ExecuteCommand executes a command through the SM APOS Shipper service
func (sce *ShipperCommandExecutor) ExecuteCommand(command string, sessionID string, timeout int) (*models.CommandResult, error) {
	sce.logger.Printf("Executing command through shipper: %s", command[:min(50, len(command))])

	// Validate session
	session, err := sce.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Validate command security
	if err := sce.securityValidator.ValidateCommandSecurity(command); err != nil {
		return nil, fmt.Errorf("command security validation failed: %w", err)
	}

	// Sanitize command
	sanitizedCommand := sce.securityValidator.SanitizeCommand(command)

	// Encrypt command for secure transport
	encryptedCommand, err := sce.encryption.EncryptCommand(sanitizedCommand, session.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt command: %w", err)
	}

	// Execute with retry logic
	var result *models.CommandResult
	var lastErr error

	for attempt := 0; attempt <= sce.maxRetries; attempt++ {
		if attempt > 0 {
			sce.logger.Printf("Retry attempt %d for command execution", attempt)
			time.Sleep(sce.retryDelay * time.Duration(attempt)) // Exponential backoff
		}

		result, err = sce.sessionManager.transport.ExecuteCommand(session, encryptedCommand, timeout)
		if err == nil {
			break
		}

		lastErr = err
		sce.logger.Printf("Command execution attempt %d failed: %v", attempt+1, err)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("command execution failed after %d attempts: %w", sce.maxRetries+1, lastErr)
	}

	// Decrypt result if it's encrypted
	if result.Stdout != "" {
		decryptedOutput, err := sce.encryption.DecryptResult(result.Stdout, session.Token)
		if err != nil {
			sce.logger.Printf("Warning: Failed to decrypt command output: %v", err)
		} else {
			result.Stdout = decryptedOutput
		}
	}

	if result.Stderr != "" {
		decryptedError, err := sce.encryption.DecryptResult(result.Stderr, session.Token)
		if err != nil {
			sce.logger.Printf("Warning: Failed to decrypt command error: %v", err)
		} else {
			result.Stderr = decryptedError
		}
	}

	// Save result to database
	if err := sce.db.SaveCommandResult(*result); err != nil {
		sce.logger.Printf("Warning: Failed to save command result to database: %v", err)
	}

	sce.logger.Printf("Command executed successfully, result ID: %s", result.ID)
	return result, nil
}

// GetCommandStatus retrieves the status of a running command
func (sce *ShipperCommandExecutor) GetCommandStatus(commandID, sessionID string) (*models.ShipperCommandStatus, error) {
	// Validate session
	session, err := sce.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Get status from transport
	status, err := sce.sessionManager.transport.GetCommandStatus(session, commandID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command status: %w", err)
	}

	return status, nil
}

// CancelCommand cancels a running command
func (sce *ShipperCommandExecutor) CancelCommand(commandID, sessionID string) error {
	// Validate session
	session, err := sce.sessionManager.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	// Cancel through transport
	if err := sce.sessionManager.transport.CancelCommand(session, commandID); err != nil {
		return fmt.Errorf("failed to cancel command: %w", err)
	}

	sce.logger.Printf("Command %s cancelled successfully", commandID)
	return nil
}

// WaitForCommandCompletion waits for a command to complete with timeout
func (sce *ShipperCommandExecutor) WaitForCommandCompletion(commandID, sessionID string, timeout time.Duration) (*models.CommandResult, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for {
		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for command completion")
		}

		// Get command status
		status, err := sce.GetCommandStatus(commandID, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check command status: %w", err)
		}

		// Check if command is complete
		if status.Status == models.StatusCompleted || status.Status == models.StatusFailed {
			// Get the final result
			result, err := sce.db.GetCommandResult(commandID)
			if err != nil {
				return nil, fmt.Errorf("failed to get command result: %w", err)
			}
			return result, nil
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}

// GetCommandHistory retrieves command execution history for a session
func (sce *ShipperCommandExecutor) GetCommandHistory(sessionID string, limit int) ([]*models.CommandResult, error) {
	// Validate session
	session, err := sce.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// In a real implementation, you'd query the database for commands by session
	// For now, return empty slice
	sce.logger.Printf("Getting command history for session %s (limit: %d)", session.UserID, limit)
	return []*models.CommandResult{}, nil
}

// ValidateCommandPreExecution performs pre-execution validation
func (sce *ShipperCommandExecutor) ValidateCommandPreExecution(command string) error {
	// Length validation
	if len(command) > 10000 {
		return fmt.Errorf("command too long: %d characters (max 10000)", len(command))
	}

	if len(command) == 0 {
		return fmt.Errorf("command cannot be empty")
	}

	// Security validation
	if err := sce.securityValidator.ValidateCommandSecurity(command); err != nil {
		return err
	}

	// Basic syntax validation
	if command[0] == '|' || command[len(command)-1] == '|' {
		return fmt.Errorf("invalid command syntax: pipe at start or end")
	}

	return nil
}

// EstimateExecutionTime provides an estimated execution time for a command
func (sce *ShipperCommandExecutor) EstimateExecutionTime(command string) time.Duration {
	// Simple estimation based on command type
	baseTime := 5 * time.Second

	// Adjust based on command complexity
	if strings.Contains(command, "sleep") || strings.Contains(command, "wait") {
		baseTime = 30 * time.Second
	} else if strings.Contains(command, "git clone") || strings.Contains(command, "npm install") {
		baseTime = 2 * time.Minute
	} else if strings.Contains(command, "find") || strings.Contains(command, "grep") {
		baseTime = 15 * time.Second
	}

	return baseTime
}

// GetExecutionStatistics returns execution statistics
func (sce *ShipperCommandExecutor) GetExecutionStatistics() map[string]interface{} {
	// In a real implementation, you'd collect metrics from the database
	return map[string]interface{}{
		"total_commands_executed": 0,
		"average_execution_time":  "0s",
		"success_rate":            0.0,
		"most_common_commands":    []string{},
	}
}

// CleanupCompletedCommands cleans up old completed command records
func (sce *ShipperCommandExecutor) CleanupCompletedCommands(maxAge time.Duration) error {
	// In a real implementation, you'd delete old command records from the database
	sce.logger.Printf("Cleaning up completed commands older than %v", maxAge)
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
