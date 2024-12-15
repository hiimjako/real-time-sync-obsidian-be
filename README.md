# Obsidian Live Syncinator Server

The Obsidian-Live-Syncinator-Server is the server of the plugin [syncinator](https://github.com/hiimjako/obsidian-live-syncinator).
# Setup

Create a `.env`:
```sh
JWT_SECRET=secret
STORAGE_DIR=/data
SQLITE_FILEPATH=/data/db.sqlite3
```

Start the docker container: 
```sh
docker run ghcr.io/hiimjako/obsidian-live-syncinator-server -p 8080:8080 --env-file .env
```

## Create a new Workspace
```sh
go run ./cmd/cli -name "workspace-name" -pass "strong-pass" -db "./data/db.sqlite3"
```

> [!WARNING]  
> The `db` argument must be the same as `SQLITE_FILEPATH` env variable.


Docker compose example:
```sh 
services:
  syncinator:
    image: ghcr.io/hiimjako/obsidian-live-syncinator-server:main
    env_file: .env
    restart: always
    container_name: syncinator
    volumes:
        - ./data:/data
    ports:
      - 8080:8080
```


# Development
## Add new migration

```sh 
GOOSE_DRIVER=sqlite GOOSE_MIGRATION_DIR=./migrations goose create new_migration_name sql
```
