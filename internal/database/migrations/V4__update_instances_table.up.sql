-- V4__update_instances_table.up.sql

-- Drop the old instances table
DROP TABLE IF EXISTS instances;

-- Recreate the instances table with last_heartbeat
CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_heartbeat DATETIME NOT NULL
);
