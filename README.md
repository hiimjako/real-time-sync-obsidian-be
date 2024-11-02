# Development
## Add new migration

```sh 
GOOSE_DRIVER=sqlite GOOSE_MIGRATION_DIR=./migrations goose create new_migration_name sql
```

# Create a new Workspace
```sh
go run ./cmd/cli -name "workspace-name" -pass "strong-pass" -db "./data/db.sqlite3"
```
