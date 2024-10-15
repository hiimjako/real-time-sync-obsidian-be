package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/screen"
)

var (
	fileId = flag.String("file", uuid.NewString(), "file to open")
)

func main() {
	flag.Parse()

	screen, err := screen.NewScreen()
	if err != nil {
		log.Fatalf("error while creating screen %v", err)
	}

	errc := make(chan error, 1)
	go func() {
		errc <- screen.Init()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}
}
