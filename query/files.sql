-- name: AddFile :exec
INSERT INTO files (id, name, path)
VALUES ($1, $2, $3);


