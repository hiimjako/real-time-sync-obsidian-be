package rtsync

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"golang.org/x/time/rate"
)

const (
	ApiV1Prefix = "/api/v1"

	PathWebSocket = ApiV1Prefix + "/sync"
)

type InternalMessage struct {
	SenderId string
	Message  DiffChunkMessage
}

type DiffChunkMessage struct {
	FileId string
	Chunks []diff.DiffChunk
}

type realTimeSyncServer struct {
	publishLimiter *rate.Limiter
	serveMux       http.ServeMux
	subscribersMu  sync.Mutex
	subscribers    map[*subscriber]struct{}
	files          map[string]string
}

func New() *realTimeSyncServer {
	rts := &realTimeSyncServer{
		publishLimiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 8),
		subscribers:    make(map[*subscriber]struct{}),
		files:          make(map[string]string),
	}

	rts.serveMux.HandleFunc(PathWebSocket, rts.subscribeHandler)

	return rts
}

func (rts *realTimeSyncServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rts.serveMux.ServeHTTP(w, r)
}

func (rts *realTimeSyncServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
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

	rts.addSubscriber(s)
	defer rts.deleteSubscriber(s)

	go func() {
		for {
			if !s.IsOpen() {
				log.Printf("client %s disconnected\n", s.clientId)
				return
			}

			data, err := s.ReadMessage()
			if err != nil {
				log.Println(err)
			}

			localCopy := rts.files[data.FileId]
			for _, d := range data.Chunks {
				localCopy = diff.ApplyDiff(localCopy, d)
			}
			diffs := diff.ComputeDiff(rts.files[data.FileId], localCopy)
			rts.files[data.FileId] = localCopy

			rts.broadcastPublish(InternalMessage{
				SenderId: s.clientId,
				Message: DiffChunkMessage{
					FileId: data.FileId,
					Chunks: diffs,
				},
			})
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
		case <-r.Context().Done():
			return r.Context().Err()
		}
	}
}

// broadcastPublish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (rts *realTimeSyncServer) broadcastPublish(msg InternalMessage) {
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
