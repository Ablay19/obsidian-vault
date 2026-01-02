CREATE TABLE IF NOT EXISTS processed_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    extracted_text TEXT,
    summary TEXT,
    topics TEXT,
    questions TEXT,
    ai_provider TEXT
);

CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY CHECK (id = 1), -- Ensures only one row can exist
    pid INTEGER NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS runtime_config (
    id INTEGER PRIMARY KEY,
    config_data BLOB,
    updated_at DATETIME
);
