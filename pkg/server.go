package rtsync

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"golang.org/x/time/rate"
)

const (
	ApiV1Prefix = "/api/v1"

	PathWebSocket = ApiV1Prefix + "/sync"
	PathFile      = ApiV1Prefix + "/file/{id}"
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
	ctx    context.Context
	cancel context.CancelFunc

	publishLimiter *rate.Limiter
	serveMux       http.ServeMux
	subscribersMu  sync.Mutex
	subscribers    map[*subscriber]struct{}
	files          map[string]string
	storageQueue   chan DiffChunkMessage
	storage        filestorage.Storage
}

func New(s filestorage.Storage) *realTimeSyncServer {
	ctx, cancel := context.WithCancel(context.Background())
	rts := &realTimeSyncServer{
		ctx:    ctx,
		cancel: cancel,

		publishLimiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 8),
		subscribers:    make(map[*subscriber]struct{}),
		files:          make(map[string]string),
		storageQueue:   make(chan DiffChunkMessage, 128),
		storage:        s,
	}

	rts.serveMux.HandleFunc(PathWebSocket, rts.subscribeHandler)
	rts.serveMux.HandleFunc(PathFile, rts.fileHandler)
	go rts.persistChunks()

	return rts
}

func (rts *realTimeSyncServer) Close() error {
	if rts.ctx.Err() != nil {
		rts.cancel()
	}
	return nil
}

func (rts *realTimeSyncServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rts.serveMux.ServeHTTP(w, r)
}

func (rts *realTimeSyncServer) persistChunks() {
	for {
		select {
		case data := <-rts.storageQueue:
			for _, d := range data.Chunks {
				err := rts.storage.PersistChunk(data.FileId, d)
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
