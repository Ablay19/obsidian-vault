-- Add runtime_config table
CREATE TABLE IF NOT EXISTS runtime_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    config_data BLOB,
    updated_at DATETIME
);

-- Insert initial config
INSERT OR IGNORE INTO runtime_config (id, config_data, updated_at) VALUES (1, '{}', CURRENT_TIMESTAMP);