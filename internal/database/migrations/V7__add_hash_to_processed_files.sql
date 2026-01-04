-- V7__add_hash_to_processed_files.sql
ALTER TABLE processed_files ADD COLUMN hash TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_processed_files_hash ON processed_files (hash);
