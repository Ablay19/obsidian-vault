-- V4__update_instances_table.down.sql
-- This migration reverts the changes made by V4__update_instances_table.up.sql

-- Drop the instances table with last_heartbeat
DROP TABLE IF EXISTS instances;

-- Recreate the instances table to its state before V4 (as per V1__initial_schema.up.sql)
CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
