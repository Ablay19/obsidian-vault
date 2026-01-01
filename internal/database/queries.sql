-- name: GetInstancePID :one
SELECT pid FROM instances WHERE id = 1;

-- name: AddInstance :exec
INSERT INTO instances (id, pid, started_at) VALUES (1, ?, ?);

-- name_order: DeleteInstance :exec
DELETE FROM instances WHERE id = 1;

-- name: FileExistsByHash :one
SELECT 1 FROM processed_files WHERE hash = ?;
