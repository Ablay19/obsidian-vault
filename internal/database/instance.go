package database

import (
	"database/sql"
	"fmt"
	"os"
	"syscall"
	"time"
)

// CheckExistingInstance checks if another instance of the bot is already running.
func CheckExistingInstance(db *sql.DB) (int, error) {
	var pid int
	err := db.QueryRow("SELECT pid FROM instances WHERE id = 1").Scan(&pid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No instance running
		}
		return 0, err
	}

	// Check if the process with the stored PID is still running
	process, err := os.FindProcess(pid)
	if err != nil {
		// On Unix-like systems, FindProcess always succeeds.
		// So, an error here is unexpected.
		return pid, fmt.Errorf("error finding process: %v", err)
	}

	// Sending signal 0 to a process checks if it exists.
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return pid, fmt.Errorf("another instance is already running with PID: %d", pid)
	}

	// If the process does not exist, remove the stale PID from the database
	_, err = db.Exec("DELETE FROM instances WHERE id = 1")
	return 0, err
}

// AddInstance records the current process's PID in the database.
func AddInstance(db *sql.DB) error {
	pid := os.Getpid()
	_, err := db.Exec("INSERT INTO instances (id, pid, started_at) VALUES (1, ?, ?)", pid, time.Now())
	return err
}

// RemoveInstance removes the current process's PID from the database.
func RemoveInstance(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM instances WHERE id = 1")
	return err
}
