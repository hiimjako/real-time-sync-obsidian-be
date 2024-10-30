-- +goose Up
-- +goose StatementBegin
CREATE TABLE files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  disk_path TEXT NOT NULL,
  virtual_path TEXT NOT NULL,
  mime_type TEXT,
  checksum TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  workspace_id INTEGER,
  UNIQUE (disk_path)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE files;

-- +goose StatementEnd
