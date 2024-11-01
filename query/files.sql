-- name: AddFile :one
INSERT INTO files (path, virtual_path, mime_type, hash, workspace_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: FetchFile :one
SELECT path, virtual_path, mime_type, hash, created_at, updated_at, workspace_id 
FROM files
WHERE id = ?
LIMIT 1;

-- name: FetchWorkspaceFiles :many
SELECT id, path, virtual_path, mime_type, hash, created_at, updated_at
FROM files
WHERE workspace_id = ?;

