-- name: AddWorkspace :exec
INSERT INTO workspaces (name, password)
VALUES (?, ?);
