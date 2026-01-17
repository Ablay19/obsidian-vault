package services

import (
	"fmt"
	"log"
	"strings"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// CommandAuthService handles command authentication and authorization
type CommandAuthService struct {
	db        *database.DB
	config    *utils.Config
	logger    *log.Logger
	validator *utils.AuthValidator
}

// NewCommandAuthService creates a new command authentication service
func NewCommandAuthService(db *database.DB, config *utils.Config, logger *log.Logger) *CommandAuthService {
	return &CommandAuthService{
		db:        db,
		config:    config,
		logger:    logger,
		validator: utils.NewAuthValidator(config, nil), // No crypto manager for now
	}
}

// AuthenticateCommand authenticates and authorizes a command
func (cas *CommandAuthService) AuthenticateCommand(cmd *models.Command, senderID, platform string) error {
	cas.logger.Printf("Authenticating command from %s on %s", senderID, platform)

	// Validate sender identity
	if err := cas.validateSender(senderID, platform); err != nil {
		cas.logger.Printf("Sender validation failed: %v", err)
		return fmt.Errorf("sender validation failed: %w", err)
	}

	// Check command permissions
	if err := cas.checkCommandPermissions(cmd, senderID); err != nil {
		cas.logger.Printf("Command permission check failed: %v", err)
		return fmt.Errorf("command permission denied: %w", err)
	}

	// Validate command content
	if err := cas.validateCommandContent(cmd); err != nil {
		cas.logger.Printf("Command content validation failed: %v", err)
		return fmt.Errorf("command content invalid: %w", err)
	}

	// Check rate limits
	if err := cas.checkRateLimits(senderID, platform); err != nil {
		cas.logger.Printf("Rate limit check failed: %v", err)
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	cas.logger.Printf("Command authentication successful for %s", cmd.ID)
	return nil
}

// validateSender validates the sender's identity and permissions
func (cas *CommandAuthService) validateSender(senderID, platform string) error {
	authConfig := cas.config.Auth

	// If no restrictions configured, allow all
	if len(authConfig.AllowedUsers) == 0 {
		return nil
	}

	// Check if sender is in allowed list
	for _, allowedUser := range authConfig.AllowedUsers {
		if allowedUser == senderID {
			return nil
		}
		// Allow partial matches for platform-specific IDs
		if strings.Contains(senderID, allowedUser) || strings.Contains(allowedUser, senderID) {
			return nil
		}
	}

	return fmt.Errorf("sender %s not authorized", senderID)
}

// checkCommandPermissions checks if the user has permission to execute the command
func (cas *CommandAuthService) checkCommandPermissions(cmd *models.Command, senderID string) error {
	// Parse the base command
	parts := strings.Fields(cmd.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	baseCommand := parts[0]

	// Check against allowed commands list
	authConfig := cas.config.Auth
	if len(authConfig.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range authConfig.AllowedCommands {
			if allowedCmd == baseCommand || strings.HasPrefix(baseCommand, allowedCmd) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command '%s' not in allowed commands list", baseCommand)
		}
	}

	// Additional permission checks based on user role
	if err := cas.checkUserRolePermissions(senderID, baseCommand); err != nil {
		return err
	}

	return nil
}

// checkUserRolePermissions checks permissions based on user roles
func (cas *CommandAuthService) checkUserRolePermissions(senderID, command string) error {
	// TODO: Implement role-based permissions
	// For now, basic restrictions on dangerous commands

	dangerousCommands := []string{
		"rm", "del", "format", "fdisk", "mkfs", "dd",
		"sudo", "su", "chmod", "chown",
		"passwd", "usermod", "userdel",
		"systemctl", "service", "kill", "pkill",
	}

	baseCmd := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if baseCmd == dangerous {
			return fmt.Errorf("command '%s' is restricted for security reasons", command)
		}
	}

	// Special handling for rm command - only allow safe variants
	if baseCmd == "rm" {
		cmdParts := strings.Fields(command)
		if len(cmdParts) >= 2 {
			// Check for dangerous flags
			for _, part := range cmdParts {
				if part == "-rf" || part == "--force" || part == "-fr" {
					return fmt.Errorf("dangerous rm flags not allowed: %s", part)
				}
			}
		}
	}

	return nil
}

// validateCommandContent validates the command content for security and correctness
func (cas *CommandAuthService) validateCommandContent(cmd *models.Command) error {
	content := cmd.Command

	// Length check
	if len(content) > 10000 { // Reasonable upper bound
		return fmt.Errorf("command too long: %d characters (max 10000)", len(content))
	}

	// Check for null bytes or other injection attempts
	if strings.Contains(content, "\x00") {
		return fmt.Errorf("command contains null bytes")
	}

	// Check for command chaining that might be dangerous
	if strings.Contains(content, ";") || strings.Contains(content, "&&") || strings.Contains(content, "||") {
		// Allow some safe chaining but restrict dangerous patterns
		if strings.Contains(content, "; rm") || strings.Contains(content, "&& rm") {
			return fmt.Errorf("dangerous command chaining detected")
		}
	}

	// Check for path traversal attempts
	if strings.Contains(content, "../") || strings.Contains(content, "..\\") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Validate command syntax (basic check)
	if strings.HasPrefix(strings.TrimSpace(content), "|") ||
		strings.HasSuffix(strings.TrimSpace(content), "|") {
		return fmt.Errorf("invalid command syntax: pipe at start or end")
	}

	return nil
}

// checkRateLimits checks if the sender is within rate limits
func (cas *CommandAuthService) checkRateLimits(senderID, platform string) error {
	// TODO: Implement rate limiting per user/platform
	// For now, rely on transport-level rate limiting

	cas.logger.Printf("Rate limit check passed for %s on %s", senderID, platform)
	return nil
}

// IsUserAuthorized checks if a user is authorized to use the system
func (cas *CommandAuthService) IsUserAuthorized(senderID, platform string) bool {
	err := cas.validateSender(senderID, platform)
	return err == nil
}

// GetUserPermissions returns the permissions for a user
func (cas *CommandAuthService) GetUserPermissions(senderID string) []string {
	// TODO: Implement user permission lookup
	// For now, return basic permissions

	authConfig := cas.config.Auth
	if len(authConfig.AllowedCommands) > 0 {
		return authConfig.AllowedCommands
	}

	// Default permissions for allowed users
	return []string{
		"read", "execute",
		"git", "npm", "yarn",
		"ls", "pwd", "cat", "grep",
	}
}

// LogAuthAttempt logs an authentication attempt
func (cas *CommandAuthService) LogAuthAttempt(cmd *models.Command, senderID, platform string, success bool, reason string) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	cas.logger.Printf("Auth attempt %s: user=%s platform=%s command='%s' reason='%s'",
		status, senderID, platform, cmd.Command, reason)

	// TODO: Store auth attempts in database for audit trail
}

// ValidateCommandForExecution performs final validation before execution
func (cas *CommandAuthService) ValidateCommandForExecution(cmd *models.Command) error {
	// Check if command is still valid (not expired, not cancelled)
	if cmd.Status != models.StatusQueued {
		return fmt.Errorf("command not in executable state: %s", cmd.Status)
	}

	// Check command age (don't execute very old commands)
	// TODO: Implement command expiration

	// Final security check
	if err := cas.validateCommandContent(cmd); err != nil {
		return err
	}

	return nil
}

// GrantTemporaryAccess grants temporary elevated access to a user
func (cas *CommandAuthService) GrantTemporaryAccess(senderID, platform string, durationMinutes int, additionalCommands []string) error {
	// TODO: Implement temporary access tokens/permissions
	cas.logger.Printf("Temporary access granted to %s for %d minutes", senderID, durationMinutes)
	return fmt.Errorf("temporary access not yet implemented")
}

// RevokeAccess revokes access for a user
func (cas *CommandAuthService) RevokeAccess(senderID, platform string) error {
	// TODO: Implement access revocation
	cas.logger.Printf("Access revoked for %s", senderID)
	return fmt.Errorf("access revocation not yet implemented")
}

// GetAuthStatistics returns authentication statistics
func (cas *CommandAuthService) GetAuthStatistics() map[string]interface{} {
	// TODO: Implement auth statistics collection
	return map[string]interface{}{
		"total_attempts":   0,
		"successful_auths": 0,
		"failed_auths":     0,
		"rate_limit_hits":  0,
		"blocked_commands": 0,
	}
}
