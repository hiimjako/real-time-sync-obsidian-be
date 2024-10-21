package rtsync

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type InternalWSMessage struct {
	SenderId string
	Message  DiffChunkMessage
}

type DiffChunkMessage struct {
	FileId string
	Chunks []diff.DiffChunk
}

func (rts *realTimeSyncServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	err := rts.subscribe(w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (rts *realTimeSyncServer) subscribe(w http.ResponseWriter, r *http.Request) error {
	s, err := NewSubscriber(w, r)
	if err != nil {
		return err
	}
	defer s.Close()

	rts.addSubscriber(s)
	defer rts.deleteSubscriber(s)

	ctx, cancel := context.WithCancel(r.Context())
	go func() {
		for {
			if !s.IsOpen() {
				cancel()
				return
			}

			data, err := s.ReadMessage()
			if err != nil {
				log.Println(err)
			}

			rts.processMessage(s, data)
		}
	}()

	for {
		select {
		case msg := <-s.msgs:
			if msg.SenderId == s.clientId {
				continue
			}

			err := s.WriteMessage(msg.Message, time.Second*1)
			if err != nil {
				log.Println("error writing message to client", err)
			}
		case <-ctx.Done():
			return nil
		case <-rts.ctx.Done():
			return nil
		}

	}
}

func (rts *realTimeSyncServer) processMessage(s *subscriber, data DiffChunkMessage) {
	localCopy := rts.files[data.FileId]
	for _, d := range data.Chunks {
		localCopy = diff.ApplyDiff(localCopy, d)
	}
	diffs := diff.ComputeDiff(rts.files[data.FileId], localCopy)
	rts.files[data.FileId] = localCopy

	rts.storageQueue <- data

	rts.broadcastPublish(InternalWSMessage{
		SenderId: s.clientId,
		Message: DiffChunkMessage{
			FileId: data.FileId,
			Chunks: diffs,
		},
	})
}

// broadcastPublish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (rts *realTimeSyncServer) broadcastPublish(msg InternalWSMessage) {
	rts.subscribersMu.Lock()
	defer rts.subscribersMu.Unlock()

	err := rts.publishLimiter.Wait(context.Background())
	if err != nil {
		log.Print(err)
	}

	for s := range rts.subscribers {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}
