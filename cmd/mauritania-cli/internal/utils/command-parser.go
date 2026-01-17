package utils

import (
	"fmt"
	"strings"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// CommandParser handles parsing and validation of commands
type CommandParser struct {
	allowedCommands  map[string]bool
	maxCommandLength int
}

// NewCommandParser creates a new command parser
func NewCommandParser() *CommandParser {
	return &CommandParser{
		allowedCommands: map[string]bool{
			// Development commands
			"git":    true,
			"npm":    true,
			"yarn":   true,
			"go":     true,
			"python": true,
			"pip":    true,

			// System commands (restricted)
			"ls":   true,
			"pwd":  true,
			"cat":  true,
			"head": true,
			"tail": true,
			"grep": true,
			"find": true,
			"ps":   true,
			"df":   true,
			"du":   true,

			// Network commands
			"curl":     true,
			"wget":     true,
			"ping":     true,
			"nslookup": true,
			"dig":      true,
		},
		maxCommandLength: 4000, // Social media message limit
	}
}

// ParseCommand parses a raw command string into structured format
func (cp *CommandParser) ParseCommand(rawCommand string) (*models.Command, error) {
	if len(rawCommand) > cp.maxCommandLength {
		return nil, fmt.Errorf("command exceeds maximum length of %d characters", cp.maxCommandLength)
	}

	if strings.TrimSpace(rawCommand) == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	// Parse command parts
	parts := strings.Fields(rawCommand)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid command format")
	}

	command := &models.Command{
		Command:  rawCommand,
		Priority: models.PriorityNormal, // Default priority
	}

	// Check if command is allowed
	baseCmd := parts[0]
	if !cp.allowedCommands[baseCmd] {
		return nil, fmt.Errorf("command '%s' is not allowed", baseCmd)
	}

	// Parse priority indicators
	if cp.hasPriorityIndicator(rawCommand) {
		priority, err := cp.parsePriority(rawCommand)
		if err != nil {
			return nil, err
		}
		command.Priority = priority
		// Remove priority indicator from command
		command.Command = cp.stripPriorityIndicator(rawCommand)
	}

	return command, nil
}

// ValidateCommand validates a parsed command
func (cp *CommandParser) ValidateCommand(command *models.Command) error {
	if command == nil {
		return fmt.Errorf("command cannot be nil")
	}

	if len(command.Command) == 0 {
		return fmt.Errorf("command text cannot be empty")
	}

	if len(command.Command) > cp.maxCommandLength {
		return fmt.Errorf("command exceeds maximum length")
	}

	// Validate priority
	switch command.Priority {
	case models.PriorityLow, models.PriorityNormal, models.PriorityHigh, models.PriorityUrgent:
		// Valid
	default:
		return fmt.Errorf("invalid command priority: %s", command.Priority)
	}

	return nil
}

// IsAllowedCommand checks if a base command is allowed
func (cp *CommandParser) IsAllowedCommand(baseCmd string) bool {
	return cp.allowedCommands[baseCmd]
}

// AddAllowedCommand adds a command to the allowed list
func (cp *CommandParser) AddAllowedCommand(command string) {
	cp.allowedCommands[command] = true
}

// RemoveAllowedCommand removes a command from the allowed list
func (cp *CommandParser) RemoveAllowedCommand(command string) {
	delete(cp.allowedCommands, command)
}

// GetAllowedCommands returns a list of allowed commands
func (cp *CommandParser) GetAllowedCommands() []string {
	commands := make([]string, 0, len(cp.allowedCommands))
	for cmd := range cp.allowedCommands {
		commands = append(commands, cmd)
	}
	return commands
}

// hasPriorityIndicator checks if command has priority indicators
func (cp *CommandParser) hasPriorityIndicator(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(command))
	return strings.HasPrefix(cmd, "urgent:") ||
		strings.HasPrefix(cmd, "high:") ||
		strings.HasPrefix(cmd, "low:")
}

// parsePriority extracts priority from command string
func (cp *CommandParser) parsePriority(command string) (models.CommandPriority, error) {
	cmd := strings.ToLower(strings.TrimSpace(command))

	if strings.HasPrefix(cmd, "urgent:") {
		return models.PriorityUrgent, nil
	}
	if strings.HasPrefix(cmd, "high:") {
		return models.PriorityHigh, nil
	}
	if strings.HasPrefix(cmd, "low:") {
		return models.PriorityLow, nil
	}

	return models.PriorityNormal, fmt.Errorf("unknown priority indicator")
}

// stripPriorityIndicator removes priority prefix from command
func (cp *CommandParser) stripPriorityIndicator(command string) string {
	cmd := strings.TrimSpace(command)

	if strings.HasPrefix(strings.ToLower(cmd), "urgent:") {
		return strings.TrimSpace(cmd[7:])
	}
	if strings.HasPrefix(strings.ToLower(cmd), "high:") {
		return strings.TrimSpace(cmd[5:])
	}
	if strings.HasPrefix(strings.ToLower(cmd), "low:") {
		return strings.TrimSpace(cmd[4:])
	}

	return cmd
}
