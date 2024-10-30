-- name: AddWorkspace :exec
INSERT INTO workspaces (name, password)
VALUES (?, ?);

-- name: FetchWorkspace :one
SELECT id, name, password 
FROM workspaces
WHERE name = ?
LIMIT 1
