package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/env"
	rtsync "github.com/hiimjako/real-time-sync-obsidian-be/pkg"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
)

var (
	host = flag.String("addr", "127.0.0.1", "host to expose server")
	port = flag.String("port", "8080", "port to expose server")
)

func main() {
	flag.Parse()

	err := run(*host, *port)
	if err != nil {
		log.Fatal(err)
	}
}

func run(host, port string) error {
	env := env.LoadEnv()

	l, err := net.Listen("tcp", net.JoinHostPort(host, port))
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
