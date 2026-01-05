-- Basic migration for development setup
CREATE TABLE IF NOT EXISTS migration_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version VARCHAR(50) NOT NULL,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Initial migration record
INSERT OR IGNORE INTO migration_history (version, applied_at) VALUES ('001_initial_setup', CURRENT_TIMESTAMP);