package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// DB provides file-based storage for development
type DB struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewDB creates a new file-based database
func NewDB(dataDir string, tursoURL string, authToken string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &DB{
		dataDir: dataDir,
	}, nil
}

// Close is a no-op for file-based storage
func (db *DB) Close() error {
	return nil
}

// Helper methods for file-based storage
func (db *DB) loadCommands() ([]models.Command, error) {
	filePath := filepath.Join(db.dataDir, "commands.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []models.Command{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var commands []models.Command
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, err
	}

	return commands, nil
}

func (db *DB) saveCommands(commands []models.Command) error {
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(db.dataDir, "commands.json")
	return os.WriteFile(filePath, data, 0644)
}

// SaveCommand saves a command to file storage
func (db *DB) SaveCommand(cmd models.Command) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	commands, err := db.loadCommands()
	if err != nil {
		return err
	}

	// Update or add command
	found := false
	for i, existing := range commands {
		if existing.ID == cmd.ID {
			commands[i] = cmd
			found = true
			break
		}
	}
	if !found {
		commands = append(commands, cmd)
	}

	return db.saveCommands(commands)
}

// GetCommand retrieves a command by ID
func (db *DB) GetCommand(id string) (*models.Command, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	commands, err := db.loadCommands()
	if err != nil {
		return nil, err
	}

	for _, cmd := range commands {
		if cmd.ID == id {
			return &cmd, nil
		}
	}

	return nil, fmt.Errorf("command not found: %s", id)
}

// SaveCommandResult saves a command result
func (db *DB) SaveCommandResult(result models.CommandResult) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	filePath := filepath.Join(db.dataDir, "results.json")

	var results []models.CommandResult
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &results); err != nil {
			return err
		}
	}

	// Update or add result
	found := false
	for i, existing := range results {
		if existing.CommandID == result.CommandID {
			results[i] = result
			found = true
			break
		}
	}
	if !found {
		results = append(results, result)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetCommandResult retrieves a command result by command ID
func (db *DB) GetCommandResult(commandID string) (*models.CommandResult, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	filePath := filepath.Join(db.dataDir, "results.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("result not found")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []models.CommandResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		if result.CommandID == commandID {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("result not found")
}

// GetPendingCommands returns commands that are queued or executing
func (db *DB) GetPendingCommands() ([]models.Command, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	commands, err := db.loadCommands()
	if err != nil {
		return nil, err
	}

	var pending []models.Command
	for _, cmd := range commands {
		if cmd.Status == models.StatusQueued || cmd.Status == models.StatusExecuting {
			pending = append(pending, cmd)
		}
	}

	// Sort by priority and timestamp (simple implementation)
	// TODO: Implement proper sorting
	return pending, nil
}

// SaveServiceConfig saves a service configuration (stub)
func (db *DB) SaveServiceConfig(config models.ServiceConfig) error {
	// TODO: Implement file-based storage for service configs
	return nil
}

// GetServiceConfigs returns all service configurations (stub)
func (db *DB) GetServiceConfigs() ([]models.ServiceConfig, error) {
	// TODO: Implement file-based storage for service configs
	return []models.ServiceConfig{}, nil
}

// SaveCommandHistory saves command execution history (stub)
func (db *DB) SaveCommandHistory(history models.CommandHistory) error {
	// TODO: Implement file-based storage for command history
	return nil
}

// DeleteCommand removes a command from storage
func (db *DB) DeleteCommand(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	commands, err := db.loadCommands()
	if err != nil {
		return err
	}

	// Remove command
	newCommands := make([]models.Command, 0, len(commands))
	for _, cmd := range commands {
		if cmd.ID != id {
			newCommands = append(newCommands, cmd)
		}
	}

	return db.saveCommands(newCommands)
}

// GetQueuedCommands returns commands that are queued or pending
func (db *DB) GetQueuedCommands() ([]models.Command, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	commands, err := db.loadCommands()
	if err != nil {
		return nil, err
	}

	var queued []models.Command
	for _, cmd := range commands {
		if cmd.Status == models.StatusQueued || cmd.Status == models.StatusReceived {
			queued = append(queued, cmd)
		}
	}

	return queued, nil
}
