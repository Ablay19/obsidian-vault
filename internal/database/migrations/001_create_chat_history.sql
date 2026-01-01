-- Add to database/migrations/
CREATE TABLE IF NOT EXISTS chat_history (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    message_id INTEGER NOT NULL,
    direction VARCHAR(10) NOT NULL, -- 'incoming' or 'outgoing'
    content_type VARCHAR(20) NOT NULL, -- 'text', 'photo', 'document'
    text_content TEXT,
    file_path TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_id ON chat_history (user_id);
CREATE INDEX IF NOT EXISTS idx_chat_id ON chat_history (chat_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON chat_history (created_at);