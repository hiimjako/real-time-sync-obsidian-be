package rtsync

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/diff"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func Test_wsHandler(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	storageStub := filestorage.NewStorageStub()
	handler := New(repository.New(db), storageStub, Options{JWTSecret: []byte("secret")})
	ts := httptest.NewServer(handler)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	url := strings.Replace(ts.URL, "http", "ws", 1) + PathWebSocket
	//nolint:bodyclose
	sender, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)

	//nolint:bodyclose
	reciver, _, err := websocket.Dial(ctx, url, nil)
	require.NoError(t, err)

	msg := DiffChunkMessage{
		FileId: "file-1",
		Chunks: []diff.DiffChunk{
			{
				Position: 0,
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
	assert.Equal(t, map[string]string{"file-1": "Hello!"}, storageStub.Files())

	t.Cleanup(func() {
		cancel()
		sender.Close(websocket.StatusNormalClosure, "")
		reciver.Close(websocket.StatusNormalClosure, "")
		ts.Close()
		handler.Close()
	})
}
