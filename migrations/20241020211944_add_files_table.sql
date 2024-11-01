-- +goose Up
-- +goose StatementBegin
CREATE TABLE files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  disk_path TEXT NOT NULL,
  workspace_path TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  hash TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  workspace_id INTEGER NOT NULL,
  UNIQUE (disk_path, workspace_path)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE files;

-- +goose StatementEnd
