// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package repository

import (
	"database/sql"
)

type File struct {
	ID            int64        `json:"id"`
	DiskPath      string       `json:"disk_path"`
	WorkspacePath string       `json:"workspace_path"`
	MimeType      string       `json:"mime_type"`
	Hash          string       `json:"hash"`
	CreatedAt     sql.NullTime `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
	WorkspaceID   int64        `json:"workspace_id"`
}

type Workspace struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Password  string       `json:"password"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}
