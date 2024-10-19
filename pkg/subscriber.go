package rtsync

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

	clientId  string
	msgs      chan InternalMessage
	closeSlow func()
}

func NewSubscriber(w http.ResponseWriter, r *http.Request) (*subscriber, error) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, err
	}

	const subscriberMessageBuffer = 8
	s := &subscriber{
		conn:     c,
		w:        w,
		r:        r,
		ctx:      r.Context(),
		msgs:     make(chan InternalMessage, subscriberMessageBuffer),
		clientId: uuid.New().String(),
		closeSlow: func() {
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}

	return s, nil
}

func (s *subscriber) IsOpen() bool {
	return s.ctx.Err() == nil
}

func (s *subscriber) Close() error {
	return s.conn.CloseNow()
}

func (s *subscriber) ReadMessage() (DiffChunkMessage, error) {
	var data DiffChunkMessage

	err := wsjson.Read(s.ctx, s.conn, &data)
	if err != nil {
		if websocket.CloseStatus(err) != -1 || strings.Contains(err.Error(), "EOF") {
			return data, fmt.Errorf("client %s disconnected", s.clientId)
		}

		return data, err
	}

	fileId := data.FileId
	if fileId == "" {
		return data, fmt.Errorf("missing fileId")
	}

	return data, nil
}

func (s *subscriber) WriteMessage(msg DiffChunkMessage, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	return wsjson.Write(ctx, s.conn, msg)
}
