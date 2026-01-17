package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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

// SaveCommandHistory saves command execution history
func (db *DB) SaveCommandHistory(history models.CommandHistory) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	filePath := filepath.Join(db.dataDir, "history.json")

	var histories []models.CommandHistory
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &histories); err != nil {
			return err
		}
	}

	// Add new history entry
	histories = append(histories, history)

	data, err := json.MarshalIndent(histories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
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

// Helper methods for session-based storage
func (db *DB) loadSessions() ([]models.ShipperSession, error) {
	filePath := filepath.Join(db.dataDir, "sessions.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []models.ShipperSession{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var sessions []models.ShipperSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (db *DB) saveSessions(sessions []models.ShipperSession) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(db.dataDir, "sessions.json")
	return os.WriteFile(filePath, data, 0644)
}

// SaveSession saves a shipper session to persistent storage
func (db *DB) SaveSession(session models.ShipperSession) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return err
	}

	// Update or add session
	found := false
	for i, existing := range sessions {
		if existing.ID == session.ID {
			sessions[i] = session
			found = true
			break
		}
	}
	if !found {
		sessions = append(sessions, session)
	}

	return db.saveSessions(sessions)
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(sessionID string) (*models.ShipperSession, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.ID == sessionID {
			return &session, nil
		}
	}

	return nil, fmt.Errorf("session not found: %s", sessionID)
}

// GetUserSession retrieves the active session for a user
func (db *DB) GetUserSession(userID string) (*models.ShipperSession, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.UserID == userID && session.ExpiresAt.After(time.Now()) {
			return &session, nil
		}
	}

	return nil, fmt.Errorf("active session not found for user: %s", userID)
}

// DeleteSession removes a session from storage
func (db *DB) DeleteSession(sessionID string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return err
	}

	// Remove session
	newSessions := make([]models.ShipperSession, 0, len(sessions))
	for _, session := range sessions {
		if session.ID != sessionID {
			newSessions = append(newSessions, session)
		}
	}

	return db.saveSessions(newSessions)
}

// GetActiveSessions returns all non-expired sessions
func (db *DB) GetActiveSessions() ([]models.ShipperSession, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return nil, err
	}

	var activeSessions []models.ShipperSession
	now := time.Now()
	for _, session := range sessions {
		if session.ExpiresAt.After(now) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// CleanupExpiredSessions removes expired sessions from storage
func (db *DB) CleanupExpiredSessions() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	sessions, err := db.loadSessions()
	if err != nil {
		return err
	}

	var activeSessions []models.ShipperSession
	now := time.Now()
	for _, session := range sessions {
		if session.ExpiresAt.After(now) {
			activeSessions = append(activeSessions, session)
		}
	}

	return db.saveSessions(activeSessions)
}

// Migration represents a database migration
type Migration struct {
	Version   int       `json:"version"`
	Name      string    `json:"name"`
	AppliedAt time.Time `json:"applied_at"`
}

// RunMigrations runs database migrations to ensure schema is up to date
func (db *DB) RunMigrations() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Load current migration state
	migrations, err := db.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get latest migration version
	latestVersion := 0
	for _, migration := range migrations {
		if migration.Version > latestVersion {
			latestVersion = migration.Version
		}
	}

	// Define migrations to run
	migrationsToRun := []struct {
		version int
		name    string
		run     func() error
	}{
		{
			version: 1,
			name:    "create_initial_schema",
			run: func() error {
				// Ensure data directory structure exists
				dirs := []string{"commands", "results", "sessions", "history", "configs"}
				for _, dir := range dirs {
					dirPath := filepath.Join(db.dataDir, dir)
					if err := os.MkdirAll(dirPath, 0755); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", dir, err)
					}
				}
				return nil
			},
		},
		{
			version: 2,
			name:    "add_session_persistence",
			run: func() error {
				// Migration already handled by adding session methods
				// Just ensure sessions.json exists
				filePath := filepath.Join(db.dataDir, "sessions.json")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					return db.saveSessions([]models.ShipperSession{})
				}
				return nil
			},
		},
		{
			version: 3,
			name:    "add_command_history",
			run: func() error {
				// Ensure command history storage exists
				filePath := filepath.Join(db.dataDir, "history.json")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					return os.WriteFile(filePath, []byte("[]"), 0644)
				}
				return nil
			},
		},
	}

	// Run pending migrations
	for _, migration := range migrationsToRun {
		if migration.version > latestVersion {
			if err := migration.run(); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migration.name, err)
			}

			// Record migration as applied
			newMigration := Migration{
				Version:   migration.version,
				Name:      migration.name,
				AppliedAt: time.Now(),
			}
			migrations = append(migrations, newMigration)
			if err := db.saveMigrations(migrations); err != nil {
				return fmt.Errorf("failed to save migration record: %w", err)
			}
		}
	}

	return nil
}

// Helper methods for migrations
func (db *DB) loadMigrations() ([]Migration, error) {
	filePath := filepath.Join(db.dataDir, "migrations.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []Migration{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	if err := json.Unmarshal(data, &migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

func (db *DB) saveMigrations(migrations []Migration) error {
	data, err := json.MarshalIndent(migrations, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(db.dataDir, "migrations.json")
	return os.WriteFile(filePath, data, 0644)
}
