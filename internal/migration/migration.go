package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose"
)

func Migrate(db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	rp, err := rootPath()
	if err != nil {
		return err
	}

	if err := goose.Up(db, filepath.Join(rp, "mirations")); err != nil {
		return err
	}

	return nil
}

func rootPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(wd, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return wd + "", nil
		}

		parentDir := filepath.Dir(wd)
		if parentDir == wd {
			break
		}
		wd = parentDir
	}

	return "", fmt.Errorf("no go.mod file found in or below the current directory")
}

