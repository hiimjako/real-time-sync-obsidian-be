package rtsync

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
)

type MessageType = int

const (
	ChunkEventType  MessageType = iota
	CreateEventType MessageType = iota
	DeleteEventType MessageType = iota
)

type WsMessageHeader struct {
	SenderId string      `json:"-"`
	FileId   int64       `json:"fileId"`
	Type     MessageType `json:"type"`
}

type EventMessage struct {
	WsMessageHeader
}

type ChunkMessage struct {
	WsMessageHeader
	Chunks []diff.DiffChunk `json:"chunks"`
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
	s, err := NewSubscriber(rts.ctx, w, r, rts.processChunkMessage, rts.broadcastEventMessage)
	if err != nil {
		return err
	}

	rts.addSubscriber(s)
	defer rts.deleteSubscriber(s)

	s.Listen()

	return nil
}

func (rts *realTimeSyncServer) processChunkMessage(data ChunkMessage) {
	rts.mut.Lock()
	defer rts.mut.Unlock()

	file := rts.files[data.FileId]
	localCopy := file.Content
	for _, d := range data.Chunks {
		localCopy = diff.ApplyDiff(localCopy, d)
	}
	diffs := diff.ComputeDiff(file.Content, localCopy)

	file.Content = localCopy
	rts.files[data.FileId] = file

	if len(diffs) > 0 {
		rts.storageQueue <- data
		rts.broadcastChunkMessage(ChunkMessage{
			WsMessageHeader: data.WsMessageHeader,
			Chunks:          diffs,
		})
	}
}

// broadcastPublish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (rts *realTimeSyncServer) broadcastChunkMessage(msg ChunkMessage) {
	rts.subscribersMu.Lock()
	defer rts.subscribersMu.Unlock()

	err := rts.publishLimiter.Wait(context.Background())
	if err != nil {
		log.Print(err)
	}

	for s := range rts.subscribers {
		select {
		case s.chunkMsgQueue <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (rts *realTimeSyncServer) broadcastEventMessage(msg EventMessage) {
	rts.subscribersMu.Lock()
	defer rts.subscribersMu.Unlock()

	err := rts.publishLimiter.Wait(context.Background())
	if err != nil {
		log.Print(err)
	}

	for s := range rts.subscribers {
		select {
		case s.eventMsgQueue <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (rts *realTimeSyncServer) writeChunks() {
	for {
		select {
		case data := <-rts.storageQueue:
			for _, d := range data.Chunks {
				file, err := rts.db.FetchFile(context.Background(), data.FileId)
				if err != nil {
					log.Println(err)
					return
				}

				err = rts.storage.PersistChunk(file.DiskPath, d)
				if err != nil {
					log.Println(err)
				}

				err = rts.db.UpdateUpdatedAt(context.Background(), data.FileId)
				if err != nil {
					log.Println(err)
				}
			}
		case <-rts.ctx.Done():
			return
		}
	}
}

func (rts *realTimeSyncServer) addSubscriber(s *subscriber) {
	rts.subscribersMu.Lock()
	rts.subscribers[s] = struct{}{}
	rts.subscribersMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (rts *realTimeSyncServer) deleteSubscriber(s *subscriber) {
	rts.subscribersMu.Lock()
	delete(rts.subscribers, s)
	rts.subscribersMu.Unlock()
}
