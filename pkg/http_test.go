package rtsync

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func Test_createFileHandler(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	storageStub := filestorage.NewStorageStub()
	server := New(repository.New(db), storageStub)

	data := File{
		ID:   1,
		Path: "/home/file",
	}
	reqBody, err := json.Marshal(data)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, PathHttpApi+"/file", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var resBody Response
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &resBody)
	assert.NoError(t, err)

	assert.Equal(t, 201, res.Code)
	assert.Equal(t, Response{Status: "success"}, resBody)

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})
}
