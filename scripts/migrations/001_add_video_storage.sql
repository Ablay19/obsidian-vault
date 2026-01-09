-- Add video storage tables
CREATE TABLE videos (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT,
    description TEXT,
    original_prompt TEXT,
    video_format TEXT DEFAULT 'mp4',
    file_size_bytes INTEGER,
    chunk_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    processing_status TEXT DEFAULT 'pending',
    error_message TEXT,
    download_token TEXT UNIQUE,
    retention_hours INTEGER DEFAULT 72,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE video_chunks (
    video_id TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    chunk_data BLOB NOT NULL,
    chunk_hash TEXT,
    PRIMARY KEY (video_id, chunk_index),
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_videos_user_id ON videos(user_id);
CREATE INDEX idx_videos_created_at ON videos(created_at);
CREATE INDEX idx_videos_expires_at ON videos(expires_at);
CREATE INDEX idx_videos_download_token ON videos(download_token);