-- Cloudflare D1 Database Schema for Obsidian Bot
-- Enhanced schema with better indexing and performance optimizations

-- Users table with enhanced fields
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT UNIQUE,
    telegram_id INTEGER UNIQUE,
    google_id TEXT UNIQUE,
    first_name TEXT,
    last_name TEXT,
    language_code TEXT,
    timezone TEXT DEFAULT 'UTC',
    is_active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    
    -- Performance indexes
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_telegram_id (telegram_id),
    INDEX idx_google_id (google_id),
    INDEX idx_created_at (created_at),
    INDEX idx_is_active (is_active)
);

-- Enhanced chat messages with AI tracking
CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    content_type TEXT NOT NULL CHECK (content_type IN ('text', 'image', 'document', 'audio', 'video')),
    text_content TEXT,
    file_path TEXT,
    file_size INTEGER,
    mime_type TEXT,
    
    -- AI-specific fields
    ai_provider_used TEXT,
    ai_model_used TEXT,
    tokens_used INTEGER DEFAULT 0,
    cost_in_cents INTEGER DEFAULT 0,
    response_time_ms INTEGER,
    cache_hit BOOLEAN DEFAULT false,
    
    -- Processing fields
    processing_status TEXT DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Performance indexes
    INDEX idx_user_chat_created (user_id, chat_id, created_at),
    INDEX idx_user_direction_created (user_id, direction, created_at),
    INDEX idx_created_at (created_at),
    INDEX idx_ai_provider (ai_provider_used),
    INDEX idx_processing_status (processing_status),
    INDEX idx_content_type_created (content_type, created_at)
);

-- AI Providers configuration and status
CREATE TABLE IF NOT EXISTS ai_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    api_endpoint TEXT NOT NULL,
    model TEXT,
    max_tokens INTEGER DEFAULT 4096,
    cost_per_token REAL DEFAULT 0.0001,
    latency_ms INTEGER DEFAULT 1000,
    
    -- Configuration
    is_enabled BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    rate_limit_per_minute INTEGER DEFAULT 60,
    supports_streaming BOOLEAN DEFAULT false,
    supports_images BOOLEAN DEFAULT false,
    
    -- Performance tracking
    total_requests INTEGER DEFAULT 0,
    successful_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    avg_response_time_ms INTEGER DEFAULT 0,
    total_cost_cents INTEGER DEFAULT 0,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Performance indexes
    INDEX idx_name (name),
    INDEX idx_is_enabled (is_enabled),
    INDEX idx_is_default (is_default),
    INDEX idx_performance (is_enabled, successful_requests, total_requests)
);

-- AI API Keys management
CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    key_id TEXT NOT NULL,
    key_name TEXT,
    key_hash TEXT NOT NULL UNIQUE, -- Hash of the actual key for security
    
    -- Status
    is_enabled BOOLEAN DEFAULT true,
    is_blocked BOOLEAN DEFAULT false,
    is_revoked BOOLEAN DEFAULT false,
    
    -- Usage tracking
    requests_this_month INTEGER DEFAULT 0,
    tokens_this_month INTEGER DEFAULT 0,
    cost_this_month_cents INTEGER DEFAULT 0,
    
    -- Limits
    monthly_request_limit INTEGER,
    monthly_token_limit INTEGER,
    monthly_cost_limit_cents INTEGER,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used DATETIME,
    expires_at DATETIME,
    
    -- Foreign key
    FOREIGN KEY (provider_id) REFERENCES ai_providers(id) ON DELETE CASCADE,
    
    -- Performance indexes
    INDEX idx_provider_id (provider_id),
    INDEX idx_key_hash (key_hash),
    INDEX idx_is_enabled (is_enabled),
    INDEX idx_expires_at (expires_at),
    INDEX idx_usage (requests_this_month, tokens_this_month)
);

-- WhatsApp media tracking
CREATE TABLE IF NOT EXISTS whatsapp_media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id INTEGER NOT NULL,
    media_id TEXT NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('image', 'document', 'audio', 'video', 'sticker')),
    mime_type TEXT,
    file_size INTEGER,
    storage_path TEXT,
    storage_type TEXT DEFAULT 'local' CHECK (storage_type IN ('local', 'r2')),
    cdn_url TEXT,
    hash_sha256 TEXT NOT NULL,
    
    -- Processing status
    download_status TEXT DEFAULT 'pending' CHECK (download_status IN ('pending', 'downloading', 'completed', 'failed')),
    processing_status TEXT DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    downloaded_at DATETIME,
    
    -- Foreign key
    FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE,
    
    -- Performance indexes
    INDEX idx_message_id (message_id),
    INDEX idx_media_id (media_id),
    INDEX idx_hash_sha256 (hash_sha256),
    INDEX idx_storage_type (storage_type),
    INDEX idx_download_status (download_status),
    INDEX idx_created_at (created_at)
);

-- SSH user management (migrated from local SQLite)
CREATE TABLE IF NOT EXISTS ssh_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    ssh_public_key TEXT,
    email TEXT UNIQUE,
    
    -- Status and limits
    is_active BOOLEAN DEFAULT true,
    is_locked BOOLEAN DEFAULT false,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until DATETIME,
    
    -- Sessions
    last_login DATETIME,
    last_login_ip TEXT,
    session_expires_at DATETIME,
    
    -- Permissions
    allowed_commands TEXT DEFAULT 'all', -- JSON array of allowed commands, or 'all'
    allowed_ip_addresses TEXT, -- JSON array of allowed IP addresses, or null
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Performance indexes
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_is_active (is_active),
    INDEX idx_is_locked (is_locked),
    INDEX idx_created_at (created_at)
);

-- SSH user activity audit log
CREATE TABLE IF NOT EXISTS ssh_user_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete', 'login', 'logout', 'failed_login', 'password_change', 'key_change')),
    ip_address TEXT,
    user_agent TEXT,
    details TEXT, -- JSON object with additional details
    
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key reference (soft, as username might change)
    -- INDEX idx_username (username),
    INDEX idx_action (action),
    INDEX idx_timestamp (timestamp)
);

-- AI usage analytics and cost tracking
CREATE TABLE IF NOT EXISTS ai_usage_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL, -- YYYY-MM-DD format
    provider TEXT NOT NULL,
    
    -- Usage metrics
    total_requests INTEGER DEFAULT 0,
    successful_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    total_cost_cents INTEGER DEFAULT 0,
    
    -- Performance metrics
    avg_response_time_ms INTEGER DEFAULT 0,
    min_response_time_ms INTEGER,
    max_response_time_ms INTEGER,
    
    -- Cache metrics
    cache_hits INTEGER DEFAULT 0,
    cache_misses INTEGER DEFAULT 0,
    cache_hit_rate REAL DEFAULT 0.0,
    
    -- User metrics
    unique_users INTEGER DEFAULT 0,
    top_prompt_types TEXT, -- JSON array of most common prompt types
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint
    UNIQUE(date, provider),
    
    -- Performance indexes
    INDEX idx_date_provider (date, provider),
    INDEX idx_date (date),
    INDEX idx_provider (provider)
);

-- Configuration settings (replaces hardcoded values)
CREATE TABLE IF NOT EXISTS app_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL, -- 'ai', 'storage', 'auth', 'ssh', etc.
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT DEFAULT 'string' CHECK (value_type IN ('string', 'integer', 'boolean', 'json')),
    description TEXT,
    
    is_encrypted BOOLEAN DEFAULT false, -- For sensitive values
    is_readonly BOOLEAN DEFAULT false, -- UI flag
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint
    UNIQUE(category, key),
    
    -- Performance indexes
    INDEX idx_category (category),
    INDEX idx_category_key (category, key)
);

-- Insert default settings
INSERT OR IGNORE INTO app_settings (category, key, value, value_type, description) VALUES
('ai', 'default_provider', 'gemini', 'string', 'Default AI provider to use'),
('ai', 'cache_ttl_seconds', '3600', 'integer', 'Default cache TTL in seconds'),
('storage', 'type', 'local', 'string', 'Storage backend: local or r2'),
('storage', 'max_file_size_mb', '50', 'integer', 'Maximum file size in MB'),
('auth', 'session_timeout_minutes', '1440', 'integer', 'Session timeout in minutes'),
('ssh', 'max_failed_attempts', '5', 'integer', 'Maximum failed login attempts'),
('ssh', 'lockout_duration_minutes', '30', 'integer', 'Account lockout duration');

-- Insert default AI providers
INSERT OR IGNORE INTO ai_providers (name, display_name, api_endpoint, model, max_tokens, cost_per_token, latency_ms, is_enabled, is_default) VALUES
('gemini', 'Google Gemini', 'https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent', 'gemini-pro', 8192, 0.000125, 800, true, true),
('groq', 'Groq', 'https://api.groq.com/openai/v1/chat/completions', 'mixtral-8x7b-32768', 4096, 0.00005, 400, true, false),
('claude', 'Anthropic Claude', 'https://api.anthropic.com/v1/messages', 'claude-3-sonnet-20240229', 100000, 0.0008, 1200, true, false),
('gpt4', 'OpenAI GPT-4', 'https://api.openai.com/v1/chat/completions', 'gpt-4-turbo-preview', 8192, 0.00003, 600, true, false);