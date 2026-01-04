-- name: GetInstancePID :one
SELECT pid FROM instances WHERE id = 1;

-- name: AddInstance :exec
INSERT INTO instances (id, pid, started_at) VALUES (1, ?, ?);

-- name: DeleteInstance :exec
DELETE FROM instances WHERE id = 1;

-- name: UpdateInstanceHeartbeat :exec
UPDATE instances SET started_at = ? WHERE id = 1;

-- name: FileExistsByHash :one
SELECT 1 FROM processed_files WHERE hash = ?;

-- name: InsertProcessedFile :exec
INSERT INTO processed_files (hash, file_name, file_path, content_type, status, summary, topics, questions, ai_provider, user_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAllProcessedFiles :many
SELECT id, hash, file_name, file_path, content_type, status, summary, topics, questions, ai_provider, user_id, created_at, updated_at FROM processed_files ORDER BY created_at DESC;

-- name: GetRecentProcessedFiles :many
SELECT id, hash, file_name, file_path, content_type, status, summary, topics, questions, ai_provider, user_id, created_at, updated_at FROM processed_files ORDER BY created_at DESC LIMIT 10;

-- name: SaveChatMessage :exec
INSERT INTO chat_history (user_id, chat_id, message_id, direction, content_type, text_content, file_path, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListChatMessages :many
SELECT id, user_id, chat_id, message_id, direction, content_type, text_content, file_path, created_at
FROM chat_history
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: ListChatMessagesGlobal :many
SELECT id, user_id, chat_id, message_id, direction, content_type, text_content, file_path, created_at
FROM chat_history
ORDER BY created_at DESC
LIMIT ?;

-- name: UpsertUser :exec
INSERT INTO users (id, username, first_name, last_name, language_code)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	username = excluded.username,
	first_name = excluded.first_name,
	last_name = excluded.last_name,
	language_code = excluded.language_code;

-- name: LinkTelegramToEmailByEmail :exec
UPDATE users
SET telegram_id = ?
WHERE email = ?;

-- name: LinkTelegramToEmailByID :exec
UPDATE users
SET email = ?
WHERE id = ? AND email IS NULL;

-- name: ListUsers :many
SELECT id, username, first_name, last_name, language_code, created_at FROM users ORDER BY created_at DESC;

-- name: GetProcessedFileCounts :one
SELECT COUNT(id) AS total_files,
       COUNT(CASE WHEN content_type LIKE 'image/%' THEN 1 END) AS image_files,
       COUNT(CASE WHEN content_type = 'application/pdf' THEN 1 END) AS pdf_files,
       MAX(created_at) AS last_activity
FROM processed_files;
