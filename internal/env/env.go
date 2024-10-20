package env

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type EnvVariables struct {
	StorageDir     string `env:"STORAGE_DIR,default=./data"`
	SqliteFilepath string `env:"SQLITE_FILEPATH,default=./data/db.sqlite3"`
}

func LoadEnv(paths ...string) *EnvVariables {
	path := ".env"
	if len(paths) > 0 {
		path = paths[0]
	}

	if err := godotenv.Load(path); err != nil {
		panic(fmt.Errorf("error loading %s file", path))
	}

	var env EnvVariables
	ctx := context.Background()
	if err := envconfig.Process(ctx, &env); err != nil {
		panic(err)
	}

	err := os.MkdirAll(env.StorageDir, 0755)
	if err != nil {
		panic(err)
	}

	return &env
}
