package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/screen"
)

var (
	client = &http.Client{}

	serverURL = flag.String("url", "http://127.0.0.1:8080", "server URL")
	fileId    = flag.String("file", uuid.NewString(), "file to open")
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
	lastContent := ""
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			content := s.Content()
			d := diff.ComputeDiff(lastContent, content)

			if len(d) == 0 {
				continue
			}

			err := sendChunk(d)
			if err != nil {
				log.Println(err)
			}

			lastContent = content
		}
	}()
}

func sendChunk(data []diff.DiffChunk) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := *serverURL + "/publish/" + *fileId
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error in request: %v", resp.Status)
	}

	return nil
}
