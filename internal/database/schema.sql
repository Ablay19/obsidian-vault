-- V1__initial_schema.up.sql
-- Initial schema for Obsidian Bot

-- Table: processed_files
CREATE TABLE IF NOT EXISTS processed_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT UNIQUE NOT NULL, -- Added hash column for deduplication
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    content_type TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    summary TEXT,
    topics TEXT,
    questions TEXT,
    ai_provider TEXT,
    user_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- Table: chat_history
CREATE TABLE IF NOT EXISTS chat_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    message_id INTEGER NOT NULL,
    direction TEXT NOT NULL, -- 'incoming' or 'outgoing'
    content_type TEXT NOT NULL, -- 'text', 'photo', 'document'
    text_content TEXT,
    file_path TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for chat_history
CREATE INDEX IF NOT EXISTS idx_user_id ON chat_history (user_id);
CREATE INDEX IF NOT EXISTS idx_chat_id ON chat_history (chat_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON chat_history (created_at);

-- Table: users
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    language_code TEXT,
    telegram_id BIGINT UNIQUE,
    email TEXT UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: instances
CREATE TABLE IF NOT EXISTS instances (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    pid INTEGER NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: runtime_config
CREATE TABLE IF NOT EXISTS runtime_config (
    id INTEGER PRIMARY KEY,
    config_data BLOB,
    updated_at DATETIME
);