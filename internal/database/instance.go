package database

import (
	"database/sql"
	"fmt"
	"time"
)

const HEARTBEAT_THRESHOLD = 30 * time.Second // Time in seconds before an instance is considered stale

// CheckExistingInstance checks if another instance of the bot is already running.
func CheckExistingInstance(db *sql.DB) error {
	var lastHeartbeat time.Time
	err := db.QueryRow("SELECT last_heartbeat FROM instances WHERE id = 1").Scan(&lastHeartbeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No instance recorded, safe to proceed
		}
		return fmt.Errorf("error querying last heartbeat: %v", err)
	}

	if time.Since(lastHeartbeat) < HEARTBEAT_THRESHOLD {
		return fmt.Errorf("another instance appears to be running (last heartbeat %s ago)", time.Since(lastHeartbeat).Round(time.Second))
	}

	// If heartbeat is stale, we can assume the previous instance died.
	// We'll proceed to add/update this instance's heartbeat.
	return nil
}

// AddInstance records the current process's presence in the database with a heartbeat.
func AddInstance(db *sql.DB) error {
	// Using INSERT OR REPLACE to handle both initial insert and update if a stale record exists
	_, err := db.Exec("INSERT OR REPLACE INTO instances (id, last_heartbeat) VALUES (1, ?)", time.Now())
	return err
}

// UpdateInstanceHeartbeat updates the last_heartbeat timestamp for the current instance.
func UpdateInstanceHeartbeat(db *sql.DB) error {
	_, err := db.Exec("UPDATE instances SET last_heartbeat = ? WHERE id = 1", time.Now())
	return err
}

// RemoveInstance removes the current process's PID from the database.
func RemoveInstance(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM instances WHERE id = 1")
	return err
}
