// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: files.sql

package repository

import (
	"context"
)

const addFile = `-- name: AddFile :one
INSERT INTO files (disk_path, workspace_path, mime_type, hash, workspace_id)
VALUES (?, ?, ?, ?, ?)
RETURNING id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
`

type AddFileParams struct {
	DiskPath      string `json:"diskPath"`
	WorkspacePath string `json:"workspacePath"`
	MimeType      string `json:"mimeType"`
	Hash          string `json:"hash"`
	WorkspaceID   int64  `json:"workspaceId"`
}

func (q *Queries) AddFile(ctx context.Context, arg AddFileParams) (File, error) {
	row := q.db.QueryRowContext(ctx, addFile,
		arg.DiskPath,
		arg.WorkspacePath,
		arg.MimeType,
		arg.Hash,
		arg.WorkspaceID,
	)
	var i File
	err := row.Scan(
		&i.ID,
		&i.DiskPath,
		&i.WorkspacePath,
		&i.MimeType,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.WorkspaceID,
	)
	return i, err
}

const deleteFile = `-- name: DeleteFile :exec
DELETE FROM files
WHERE id = ?
`

func (q *Queries) DeleteFile(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteFile, id)
	return err
}

const fetchAllFiles = `-- name: FetchAllFiles :many
SELECT id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
FROM files
`

func (q *Queries) FetchAllFiles(ctx context.Context) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, fetchAllFiles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []File
	for rows.Next() {
		var i File
		if err := rows.Scan(
			&i.ID,
			&i.DiskPath,
			&i.WorkspacePath,
			&i.MimeType,
			&i.Hash,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Version,
			&i.WorkspaceID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fetchFile = `-- name: FetchFile :one
SELECT id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
FROM files
WHERE id = ?
LIMIT 1
`

func (q *Queries) FetchFile(ctx context.Context, id int64) (File, error) {
	row := q.db.QueryRowContext(ctx, fetchFile, id)
	var i File
	err := row.Scan(
		&i.ID,
		&i.DiskPath,
		&i.WorkspacePath,
		&i.MimeType,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.WorkspaceID,
	)
	return i, err
}

const fetchFileFromWorkspacePath = `-- name: FetchFileFromWorkspacePath :one
SELECT id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
FROM files
WHERE workspace_path = ?
LIMIT 1
`

func (q *Queries) FetchFileFromWorkspacePath(ctx context.Context, workspacePath string) (File, error) {
	row := q.db.QueryRowContext(ctx, fetchFileFromWorkspacePath, workspacePath)
	var i File
	err := row.Scan(
		&i.ID,
		&i.DiskPath,
		&i.WorkspacePath,
		&i.MimeType,
		&i.Hash,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
		&i.WorkspaceID,
	)
	return i, err
}

const fetchFiles = `-- name: FetchFiles :many
SELECT id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
FROM files
WHERE workspace_id = ?
`

func (q *Queries) FetchFiles(ctx context.Context, workspaceID int64) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, fetchFiles, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []File
	for rows.Next() {
		var i File
		if err := rows.Scan(
			&i.ID,
			&i.DiskPath,
			&i.WorkspacePath,
			&i.MimeType,
			&i.Hash,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Version,
			&i.WorkspaceID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fetchWorkspaceFiles = `-- name: FetchWorkspaceFiles :many
SELECT id, disk_path, workspace_path, mime_type, hash, created_at, updated_at, version, workspace_id
FROM files
WHERE workspace_id = ?
`

func (q *Queries) FetchWorkspaceFiles(ctx context.Context, workspaceID int64) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, fetchWorkspaceFiles, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []File
	for rows.Next() {
		var i File
		if err := rows.Scan(
			&i.ID,
			&i.DiskPath,
			&i.WorkspacePath,
			&i.MimeType,
			&i.Hash,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Version,
			&i.WorkspaceID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUpdatedAt = `-- name: UpdateUpdatedAt :exec
UPDATE files
SET 
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`

func (q *Queries) UpdateUpdatedAt(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, updateUpdatedAt, id)
	return err
}
