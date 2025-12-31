-- name: CreateFile :one
INSERT INTO processed_files (
    hash,
    category
) VALUES (
    ?, ?
)
RETURNING *;

-- name: GetFileByHash :one
SELECT * FROM processed_files
WHERE hash = ?;

-- name: GetStats :one
SELECT
    (SELECT COUNT(*) FROM processed_files) AS total_files,
    (SELECT COUNT(*) FROM processed_files WHERE category = 'general') AS general_files,
    (SELECT COUNT(*) FROM processed_files WHERE category = 'unprocessed') AS unprocessed_files,
    (SELECT COUNT(*) FROM processed_files WHERE category = 'math') AS math_files,
    (SELECT COUNT(*) FROM processed_files WHERE category = 'physics') AS physics_files,
    (SELECT COUNT(*) FROM processed_files WHERE category = 'admin') AS admin_files;

-- name: UpdateFileText :exec
UPDATE processed_files
SET extracted_text = ?
WHERE hash = ?;