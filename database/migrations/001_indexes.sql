-- Database indexes and optimizations for better query performance
-- This file contains SQL commands to create indexes and optimize database queries

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users(updated_at);

-- Sessions table indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);

-- Processing files table indexes
CREATE INDEX IF NOT EXISTS idx_processing_files_user_id ON processing_files(user_id);
CREATE INDEX IF NOT EXISTS idx_processing_files_status ON processing_files(status);
CREATE INDEX IF NOT EXISTS idx_processing_files_created_at ON processing_files(created_at);
CREATE INDEX IF NOT EXISTS idx_processing_files_updated_at ON processing_files(updated_at);
CREATE INDEX IF NOT EXISTS idx_processing_files_hash ON processing_files(hash);
CREATE INDEX IF NOT EXISTS idx_processing_files_ai_provider ON processing_files(ai_provider);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_processing_files_user_status ON processing_files(user_id, status);
CREATE INDEX IF NOT EXISTS idx_processing_files_status_created ON processing_files(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_processing_files_user_created ON processing_files(user_id, created_at DESC);

-- API keys table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

-- WhatsApp messages table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_sender ON whatsapp_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_type ON whatsapp_messages(message_type);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_conversation ON whatsapp_messages(conversation_id, timestamp DESC);

-- AI conversations table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id ON ai_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_created_at ON ai_conversations(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_updated_at ON ai_conversations(updated_at);

-- AI messages table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_id ON ai_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_messages_created_at ON ai_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_messages_role ON ai_messages(role);

-- Vector store indexes (if using PostgreSQL with pgvector)
-- Note: These require pgvector extension to be installed
-- CREATE EXTENSION IF NOT EXISTS vector;
-- CREATE INDEX IF NOT EXISTS idx_vectors_embedding ON vectors USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Full-text search indexes
CREATE INDEX IF NOT EXISTS idx_processing_files_content_fts ON processing_files USING gin(to_tsvector('english', summary || ' ' || topics || ' ' || questions));
CREATE INDEX IF NOT EXISTS idx_ai_messages_content_fts ON ai_messages USING gin(to_tsvector('english', content));

-- Rate limiting table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_rate_limits_user_id ON rate_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_rate_limits_window_start ON rate_limits(window_start);
CREATE INDEX IF NOT EXISTS idx_rate_limits_expires_at ON rate_limits(expires_at);

-- Audit logs table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp_user ON audit_logs(timestamp DESC, user_id);

-- Metrics table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp_name ON metrics(timestamp DESC, name);

-- Notifications table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(notification_type);

-- Teams table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
CREATE INDEX IF NOT EXISTS idx_teams_created_at ON teams(created_at);

-- Team members table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role);

-- Invitations table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email);
CREATE INDEX IF NOT EXISTS idx_invitations_team_id ON invitations(team_id);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_at ON invitations(expires_at);
CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status);

-- Webhooks table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_url ON webhooks(url);
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON webhooks(events);
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(active);

-- Webhook deliveries table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created_at ON webhook_deliveries(created_at DESC);

-- Usage tracking table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_usage_user_id ON usage_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_date ON usage_tracking(date DESC);
CREATE INDEX IF NOT EXISTS idx_usage_resource ON usage_tracking(resource);
CREATE INDEX IF NOT EXISTS idx_usage_user_date ON usage_tracking(user_id, date DESC);

-- Billing records table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_billing_records_user_id ON billing_records(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_records_date ON billing_records(date DESC);
CREATE INDEX IF NOT EXISTS idx_billing_records_status ON billing_records(status);
CREATE INDEX IF NOT EXISTS idx_billing_records_amount ON billing_records(amount);

-- Subscriptions table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires_at ON subscriptions(expires_at);

-- Performance optimization: Analyze table statistics
ANALYZE users;
ANALYZE sessions;
ANALYZE processing_files;
ANALYZE api_keys;
ANALYZE whatsapp_messages;
ANALYZE ai_conversations;
ANALYZE ai_messages;
ANALYZE rate_limits;
ANALYZE audit_logs;
ANALYZE metrics;
ANALYZE notifications;
ANALYZE teams;
ANALYZE team_members;
ANALYZE invitations;
ANALYZE webhooks;
ANALYZE webhook_deliveries;
ANALYZE usage_tracking;
ANALYZE billing_records;
ANALYZE subscriptions;

-- Create partial indexes for better performance on filtered queries
CREATE INDEX IF NOT EXISTS idx_processing_files_active ON processing_files(created_at DESC) WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS idx_processing_files_failed ON processing_files(created_at DESC) WHERE status = 'failed';
CREATE INDEX IF NOT EXISTS idx_processing_files_completed ON processing_files(created_at DESC) WHERE status = 'completed';

CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(expires_at) WHERE expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(expires_at) WHERE (expires_at IS NULL OR expires_at > NOW());
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(created_at DESC) WHERE status = 'unread';

-- Indexes for JSON/JSONB columns (PostgreSQL specific)
-- CREATE INDEX IF NOT EXISTS idx_processing_files_metadata ON processing_files USING gin(metadata);
-- CREATE INDEX IF NOT EXISTS idx_processing_files_topics ON processing_files USING gin(topics);
-- CREATE INDEX IF NOT EXISTS idx_processing_files_questions ON processing_files USING gin(questions);

-- Create covering indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_processing_files_list ON processing_files(user_id, status, created_at DESC) INCLUDE (id, file_name, summary);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_list ON ai_conversations(user_id, created_at DESC) INCLUDE (id, title, updated_at);
CREATE INDEX IF NOT EXISTS idx_notifications_list ON notifications(user_id, status, created_at DESC) INCLUDE (id, title, message, notification_type);

-- Index for foreign key constraints (automatically created by PostgreSQL, but good practice to be explicit)
-- These are typically created automatically when defining foreign keys, but we can ensure they exist