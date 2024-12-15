# Obsidian Live Syncinator Server

The Obsidian-Live-Syncinator-Server is the server of the plugin [syncinator](https://github.com/hiimjako/obsidian-live-syncinator).
# Setup

Start the docker container: 
```sh
docker run ghcr.io/hiimjako/obsidian-live-syncinator-server -p 8080:8080
```

## Create a new Workspace
```sh
go run ./cmd/cli -name "workspace-name" -pass "strong-pass" -db "./data/db.sqlite3"
```

> [!WARNING]  
> The `db` argument must be the same as `SQLITE_FILEPATH` env variable.


# Development
## Add new migration

```sh 
GOOSE_DRIVER=sqlite GOOSE_MIGRATION_DIR=./migrations goose create new_migration_name sql
```
