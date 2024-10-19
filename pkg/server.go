package rtsync

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
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
	subscriberMessageBuffer int
	publishLimiter          *rate.Limiter
	serveMux                http.ServeMux
	subscribersMu           sync.Mutex
	subscribers             map[*subscriber]struct{}
	files                   map[string]string
}

func New() *realTimeSyncServer {
	rts := &realTimeSyncServer{
		subscriberMessageBuffer: 8,
		publishLimiter:          rate.NewLimiter(rate.Every(100*time.Millisecond), 8),
		subscribers:             make(map[*subscriber]struct{}),
		files:                   make(map[string]string),
	}

	rts.serveMux.HandleFunc(PathWebSocket, rts.subscribeHandler)

	return rts
}

type subscriber struct {
	msgs      chan InternalMessage
	closeSlow func()
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
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	s := &subscriber{
		msgs: make(chan InternalMessage, rts.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}
	rts.addSubscriber(s)
	defer rts.deleteSubscriber(s)

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	//nolint:errcheck
	defer c.CloseNow()

	clientId := uuid.New().String()

	go func() {
		for {
			if r.Context().Err() != nil {
				return
			}

			var data DiffChunkMessage
			err := wsjson.Read(r.Context(), c, &data)
			if websocket.CloseStatus(err) != -1 {
				log.Println("Client disconnected", err)
				return
			}

			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					log.Println("Client disconnected", err)
					return
				}
				log.Println("Error reading message:", err)
				continue
			}

			fileId := data.FileId
			if fileId == "" {
				log.Println("Missing fileId", err)
				continue
			}

			localCopy := rts.files[fileId]
			for _, d := range data.Chunks {
				localCopy = diff.ApplyDiff(localCopy, d)
			}
			diffs := diff.ComputeDiff(rts.files[fileId], localCopy)
			rts.files[fileId] = localCopy

			rts.publish(InternalMessage{
				SenderId: clientId,
				Message: DiffChunkMessage{
					FileId: fileId,
					Chunks: diffs,
				},
			})
		}
	}()

	for {
		select {
		case msg := <-s.msgs:
			if msg.SenderId == clientId {
				continue
			}

			err := writeTimeout(r.Context(), time.Second*1, c, msg.Message)
			if err != nil {
				log.Println("error writing message to client", err)
			}
		case <-r.Context().Done():
			return r.Context().Err()
		}
	}
}

// publish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (rts *realTimeSyncServer) publish(msg InternalMessage) {
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

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg DiffChunkMessage) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, c, msg)
}
