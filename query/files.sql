-- name: AddFile :one
INSERT INTO files (disk_path, workspace_path, mime_type, hash, workspace_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: FetchFile :one
SELECT *
FROM files
WHERE id = ?
LIMIT 1;

-- name: FetchWorkspaceFiles :many
SELECT *
FROM files
WHERE workspace_id = ?;

-- name: DeleteFile :exec
DELETE FROM files
WHERE id = ?;

