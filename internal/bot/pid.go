package bot

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

const pidFile = "bot.pid"

// CreatePIDFile creates a PID file to ensure only one instance of the bot is running.
func CreatePIDFile() error {
	// Check if PID file exists
	if _, err := os.Stat(pidFile); err == nil {
		// PID file exists, check if the process is running
		pidBytes, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("failed to read existing PID file: %w", err)
		}
		pid, err := strconv.Atoi(string(pidBytes))
		if err != nil {
			return fmt.Errorf("failed to parse PID from existing PID file: %w", err)
		}

		// Check if the process is running
		process, err := os.FindProcess(pid)
		if err == nil {
			// On Unix systems, os.FindProcess always succeeds, so we need to send a signal to check for existence
			err := process.Signal(syscall.Signal(0))
			if err == nil {
				return fmt.Errorf("bot is already running with PID %d", pid)
			}
		}
	}

	// Create a new PID file
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}
	return nil
}

// RemovePIDFile removes the PID file.
func RemovePIDFile() {
	os.Remove(pidFile)
}
