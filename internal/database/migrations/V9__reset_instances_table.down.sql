-- V9__reset_instances_table.down.sql
-- This migration reverts the instances table to the schema with last_heartbeat.

-- Drop the table to ensure a clean state.
DROP TABLE IF EXISTS instances;

-- Recreate it with the 'last_heartbeat' column, as defined in a previous version.
CREATE TABLE instances (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_heartbeat DATETIME NOT NULL
);
