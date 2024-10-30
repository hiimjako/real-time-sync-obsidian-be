-- name: AddFile :exec
INSERT INTO files (disk_path, virtual_path, mime_type, checksum, workspace_id)
VALUES (?, ?, ?, ?, ?);

