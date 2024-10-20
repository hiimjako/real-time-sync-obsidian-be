-- name: AddFile :exec
INSERT INTO files (id, name, path)
VALUES (?, ?, ?);


