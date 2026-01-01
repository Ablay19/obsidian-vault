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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_chat_id (chat_id),
    INDEX idx_created_at (created_at)
);