-- +goose Up
-- +goose StatementBegin
CREATE TABLE workspaces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  password TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (name)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE workspaces;

-- +goose StatementEnd
