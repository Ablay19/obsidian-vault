-- V9__reset_instances_table.up.sql
-- This migration unconditionally resets the instances table to the correct schema.

-- Drop the table regardless of its current state.
DROP TABLE IF EXISTS instances;

-- Recreate it with the schema required by the application, including the 'pid' column.
CREATE TABLE instances (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    pid INTEGER NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
