package rtsync

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	ts := httptest.NewServer(New())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	url := strings.Replace(ts.URL, "http", "ws", 1) + PathWebSocket
	sender, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)

	reciver, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)

	msg := DiffChunkMessage{
		FileId: "file-1",
		Chunks: []diff.DiffChunk{
			{
				Position: 1,
				Type:     diff.DiffAdd,
				Text:     "Hello!",
				Len:      6,
			},
		},
	}

	err = wsjson.Write(ctx, sender, msg)
	assert.NoError(t, err)
	go func() {
		// should not recive any message
		var recMsg DiffChunkMessage
		err = wsjson.Read(ctx, sender, &recMsg)
		assert.Error(t, err)
	}()

	var recMsg DiffChunkMessage
	err = wsjson.Read(ctx, reciver, &recMsg)
	assert.NoError(t, err)

	assert.Equal(t, msg, recMsg)

	t.Cleanup(func() {
		cancel()
		sender.Close(websocket.StatusNormalClosure, "")
		ts.Close()
	})
}
