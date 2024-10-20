package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/env"
	rtsync "github.com/hiimjako/real-time-sync-obsidian-be/pkg"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/pressly/goose"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	env := env.LoadEnv()

	err := run(env)
	if err != nil {
		log.Fatal(err)
	}
}

func run(env *env.EnvVariables) error {
	log.Printf("running migrations")
	err := migrate(env.SqliteFilepath)
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", net.JoinHostPort(env.Host, env.Port))
	if err != nil {
		return err
	}
	log.Printf("listening on ws://%v", l.Addr())

	disk := filestorage.NewDisk(env.StorageDir)

	handler := rtsync.New(disk)
	defer handler.Close()

	s := &http.Server{
		Handler:      handler,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}

func migrate(sqliteFilepath string) error {
	db, err := sql.Open("sqlite3", sqliteFilepath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, filepath.Join(".", "migrations")); err != nil {
		return err
	}

	return nil
}
