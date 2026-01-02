-- name: GetInstancePID :one
SELECT pid FROM instances WHERE id = 1;

-- name: AddInstance :exec
INSERT INTO instances (id, pid, started_at) VALUES (1, ?, ?);

-- name: DeleteInstance :exec
DELETE FROM instances WHERE id = 1;

-- name: FileExistsByHash :one
SELECT 1 FROM processed_files WHERE hash = ?;

-- name: GetAllProcessedFiles :many
SELECT hash, category, timestamp, extracted_text FROM processed_files ORDER BY timestamp DESC;

-- name: GetRecentProcessedFiles :many
SELECT id, hash, category, timestamp, extracted_text, summary, topics, questions, ai_provider FROM processed_files ORDER BY timestamp DESC LIMIT 10;
