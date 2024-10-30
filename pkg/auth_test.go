package rtsync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/migration"
	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func Test_fetchWorkspaceHandler(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, migration.Migrate(db))

	repo := repository.New(db)
	storageStub := filestorage.NewStorageStub()
	server := New(repo, storageStub)

	require.NoError(t, repo.AddWorkspace(context.Background(), repository.AddWorkspaceParams{
		Name:     "workspace1",
		Password: "strong_password",
	}))

	data := WorkspaceCredentials{
		Name:     "workspace1",
		Password: "strong_password",
	}
	reqBody, err := json.Marshal(data)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, PathHttpAuth+"/login", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var resBody Response
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &resBody)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, Response{Status: "success"}, resBody)

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})
}
