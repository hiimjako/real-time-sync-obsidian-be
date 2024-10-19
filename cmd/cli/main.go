package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	rtsync "github.com/hiimjako/real-time-sync-obsidian-be/pkg"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/screen"
)

var (
	serverURL = flag.String("url", "127.0.0.1:8080", "server URL")
)

func main() {
	log.SetOutput(os.Stderr)
	flag.Parse()

	screen, err := screen.NewScreen()
	if err != nil {
		log.Fatalf("error while creating screen %v", err)
	}

	errc := make(chan error, 1)
	go func() {
		errc <- screen.Init()
	}()

	go pollText(&screen)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}
}

func pollText(s *screen.Screen) {
	var mu sync.Mutex

	ctx := context.Background()
	url := "ws://" + *serverURL + rtsync.PathWebSocket
	ws, _, err := websocket.Dial(ctx, url, nil)
	logOnError(err)

	lastContent := ""
	go func() {
		// listen for changes in ws
		for {
			var msg rtsync.DiffChunkMessage
			err = wsjson.Read(ctx, ws, &msg)
			logOnError(err)

			mu.Lock()
			lastContent = s.ApplyDiff(msg.Chunks)
			s.Render()
			mu.Unlock()
		}
	}()

	go func() {
		// send local changes to server
		for {
			<-time.After(1 * time.Millisecond)

			mu.Lock()
			content := s.Content()
			d := diff.ComputeDiff(lastContent, content)

			log.Println(d)
			if len(d) == 0 {
				mu.Unlock()
				continue
			}

			err = wsjson.Write(ctx, ws, rtsync.DiffChunkMessage{
				FileId: "cli",
				Chunks: d,
			})
			logOnError(err)

			lastContent = content
			mu.Unlock()
		}
	}()
}

func logOnError(err error) {
	if err != nil {
		log.Println(err)
	}
}
