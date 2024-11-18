package rtsync

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type subscriber struct {
	conn *websocket.Conn
	w    http.ResponseWriter
	r    *http.Request
	ctx  context.Context

	isConnected   atomic.Bool
	clientId      string
	chunkMsgQueue chan ChunkMessage
	eventMsgQueue chan EventMessage
	closeSlow     func()
	onMessage     func(ChunkMessage)
}

func NewSubscriber(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	onMessage func(ChunkMessage),
) (*subscriber, error) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"127.0.0.1", "obsidian.md"},
	})
	if err != nil {
		return nil, err
	}

	const subscriberMessageBuffer = 8
	s := &subscriber{
		conn:          c,
		w:             w,
		r:             r,
		ctx:           ctx,
		isConnected:   atomic.Bool{},
		chunkMsgQueue: make(chan ChunkMessage, subscriberMessageBuffer),
		eventMsgQueue: make(chan EventMessage, subscriberMessageBuffer),
		clientId:      uuid.New().String(),
		closeSlow: func() {
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
		onMessage: onMessage,
	}

	s.isConnected.Store(true)

	return s, nil
}

func (s *subscriber) IsOpen() bool {
	return s.isConnected.Load()
}

func (s *subscriber) Close() error {
	s.isConnected.Store(false)
	return s.conn.CloseNow()
}

func (s *subscriber) Listen() {
	// on ws message
	go func() {
		for {
			if !s.IsOpen() {
				return
			}

			data, err := s.ReadMessage()
			if err != nil {
				log.Println(err)
			}

			s.onMessage(data)
		}
	}()

	// on internal queue event
	go func() {
		for {
			select {
			case chunkMsg := <-s.chunkMsgQueue:
				if chunkMsg.SenderId == s.clientId {
					continue
				}

				err := s.WriteMessage(chunkMsg, time.Second*1)
				if err != nil {
					log.Println("error writing message to client", err)
				}
			case eventMsg := <-s.eventMsgQueue:
				if eventMsg.SenderId == s.clientId {
					continue
				}

				err := s.WriteMessage(eventMsg, time.Second*1)
				if err != nil {
					log.Println("error writing message to client", err)
				}
			case <-s.ctx.Done():
				s.Close()
				return
			case <-s.r.Context().Done():
				s.Close()
				return
			}
		}
	}()

	<-s.ctx.Done()
}

func (s *subscriber) ReadMessage() (ChunkMessage, error) {
	var data ChunkMessage

	err := wsjson.Read(s.ctx, s.conn, &data)
	if err != nil {
		if websocket.CloseStatus(err) != -1 || strings.Contains(err.Error(), "EOF") {
			s.Close()
			return data, fmt.Errorf("client %s disconnected", s.clientId)
		}

		return data, err
	}

	if data.FileId <= 0 {
		return data, fmt.Errorf("missing fileId")
	}

	data.SenderId = s.clientId

	return data, nil
}

func (s *subscriber) WriteMessage(msg any, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, s.conn, msg)
}
