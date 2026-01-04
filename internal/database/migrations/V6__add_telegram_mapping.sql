-- V6__add_telegram_mapping.sql
-- Link Telegram and Google/Dashboard accounts via email
ALTER TABLE users ADD COLUMN telegram_id BIGINT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_id ON users (telegram_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
