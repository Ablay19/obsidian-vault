package utils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/database"
	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// OfflineQueue manages commands that failed due to network issues
type OfflineQueue struct {
	db             *database.DB
	networkMonitor *NetworkMonitor
	logger         *log.Logger
	mu             sync.RWMutex
	retryConfig    RetryConfig
	isProcessing   bool
	stopChan       chan struct{}
	wakeChan       chan struct{}
}

// NewOfflineQueue creates a new offline queue manager
func NewOfflineQueue(db *database.DB, networkMonitor *NetworkMonitor, logger *log.Logger) *OfflineQueue {
	oq := &OfflineQueue{
		db:             db,
		networkMonitor: networkMonitor,
		logger:         logger,
		retryConfig:    MobileRetryConfig(), // Use mobile config by default
		stopChan:       make(chan struct{}),
		wakeChan:       make(chan struct{}, 1),
	}

	// Start processing when network comes back online
	networkMonitor.SetStatusCallback(func(status NetworkStatus) {
		if status.IsOnline && !oq.isProcessing {
			select {
			case oq.wakeChan <- struct{}{}:
			default:
			}
		}
	})

	return oq
}

// Start begins offline queue processing
func (oq *OfflineQueue) Start() {
	go oq.processingLoop()
}

// Stop stops offline queue processing
func (oq *OfflineQueue) Stop() {
	close(oq.stopChan)
}

// AddFailedCommand adds a command that failed due to network issues
func (oq *OfflineQueue) AddFailedCommand(cmd models.Command, error error) error {
	oq.mu.Lock()
	defer oq.mu.Unlock()

	oq.logger.Printf("Adding failed command %s to offline queue: %v", cmd.ID, error)

	// Mark command as failed in database
	cmd.Status = models.StatusQueued // Keep it queued for retry
	if err := oq.db.SaveCommand(cmd); err != nil {
		return fmt.Errorf("failed to save failed command: %w", err)
	}

	// Wake up processing loop
	select {
	case oq.wakeChan <- struct{}{}:
	default:
	}

	return nil
}

// GetQueuedCommands returns all commands waiting for retry
func (oq *OfflineQueue) GetQueuedCommands() ([]models.Command, error) {
	return oq.db.GetPendingCommands()
}

// processingLoop runs the offline queue processing
func (oq *OfflineQueue) processingLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-oq.stopChan:
			return
		case <-oq.wakeChan:
			oq.processQueue()
		case <-ticker.C:
			if oq.networkMonitor.IsOnline() {
				oq.processQueue()
			}
		}
	}
}

// processQueue attempts to retry queued commands
func (oq *OfflineQueue) processQueue() {
	oq.mu.Lock()
	if oq.isProcessing {
		oq.mu.Unlock()
		return
	}
	oq.isProcessing = true
	oq.mu.Unlock()

	defer func() {
		oq.mu.Lock()
		oq.isProcessing = false
		oq.mu.Unlock()
	}()

	// Get pending commands
	commands, err := oq.db.GetPendingCommands()
	if err != nil {
		oq.logger.Printf("Failed to get pending commands: %v", err)
		return
	}

	if len(commands) == 0 {
		return
	}

	oq.logger.Printf("Processing %d queued commands", len(commands))

	// Process commands (in a real implementation, this would send them via transports)
	for _, cmd := range commands {
		if err := oq.processCommand(cmd); err != nil {
			oq.logger.Printf("Failed to process command %s: %v", cmd.ID, err)
			continue
		}
	}
}

// processCommand attempts to send a queued command
func (oq *OfflineQueue) processCommand(cmd models.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), oq.retryConfig.Timeout)
	defer cancel()

	oq.logger.Printf("Retrying command: %s", cmd.Command)

	// In a real implementation, this would use the appropriate transport
	// For now, we'll simulate network operations
	err := RetryOperation(ctx, oq.retryConfig, func() error {
		// Simulate network operation that might fail
		if !oq.networkMonitor.IsOnline() {
			return fmt.Errorf("network offline")
		}

		// Simulate some commands succeeding, some failing
		// In real implementation, this would be actual transport logic
		time.Sleep(100 * time.Millisecond) // Simulate network delay

		// For demo purposes, succeed for some commands
		if cmd.ID[len(cmd.ID)-1] > '5' { // Arbitrary success condition
			return nil
		}
		return fmt.Errorf("simulated network error")
	})

	if err != nil {
		oq.logger.Printf("Command %s failed after retries: %v", cmd.ID, err)
		return err
	}

	// Mark command as completed (in real implementation, this would be done by transport)
	cmd.Status = models.StatusCompleted
	if saveErr := oq.db.SaveCommand(cmd); saveErr != nil {
		oq.logger.Printf("Warning: failed to update command status: %v", saveErr)
	}

	// Save command result
	result := models.CommandResult{
		ID:            generateResultID(),
		CommandID:     cmd.ID,
		Status:        "success",
		ExitCode:      0,
		Stdout:        "Command executed successfully via offline retry",
		Stderr:        "",
		ExecutionTime: 100,
		TransportUsed: string(cmd.TransportID),
		Cost:          0.0,
		CompletedAt:   time.Now(),
	}

	if saveErr := oq.db.SaveCommandResult(result); saveErr != nil {
		oq.logger.Printf("Warning: failed to save command result: %v", saveErr)
	}

	oq.logger.Printf("Command %s completed successfully", cmd.ID)
	return nil
}

// generateResultID creates a unique result ID
func generateResultID() string {
	return fmt.Sprintf("result-%d", time.Now().UnixNano())
}

// QueueStats represents offline queue statistics
type QueueStats struct {
	TotalQueued   int
	Processing    bool
	LastProcessed time.Time
	NetworkStatus NetworkStatus
}

// GetStats returns current queue statistics
func (oq *OfflineQueue) GetStats() QueueStats {
	commands, err := oq.db.GetPendingCommands()
	totalQueued := 0
	if err == nil {
		totalQueued = len(commands)
	}

	oq.mu.RLock()
	processing := oq.isProcessing
	oq.mu.RUnlock()

	return QueueStats{
		TotalQueued:   totalQueued,
		Processing:    processing,
		LastProcessed: time.Now(), // TODO: track actual last processed time
		NetworkStatus: oq.networkMonitor.GetStatus(),
	}
}

// ForceRetry triggers immediate retry of all queued commands
func (oq *OfflineQueue) ForceRetry() {
	oq.logger.Printf("Forcing retry of all queued commands")
	select {
	case oq.wakeChan <- struct{}{}:
	default:
	}
}
