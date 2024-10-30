package rtsync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

	t.Run("existing user", func(t *testing.T) {
		data := WorkspaceCredentials{
			Name:     "workspace1",
			Password: "strong_password",
		}
		code, res, _ := sendRequest(t, server, data)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, Response{Status: "success"}, res)
	})

	t.Run("wrong password", func(t *testing.T) {
		data := WorkspaceCredentials{
			Name:     "workspace1",
			Password: "invalid_pass",
		}
		code, _, errStr := sendRequest(t, server, data)
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, ErrIncorrectPassword, errStr)
	})

	t.Run("missing user", func(t *testing.T) {
		data := WorkspaceCredentials{
			Name:     "workspace2",
			Password: "random",
		}

		code, _, errStr := sendRequest(t, server, data)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Equal(t, ErrWorkspaceNotFound, errStr)
	})

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})
}

func sendRequest(t *testing.T, server *realTimeSyncServer, data WorkspaceCredentials) (int, Response, string) {
	reqBody, err := json.Marshal(data)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, PathHttpAuth+"/login", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	if res.Code == http.StatusOK {
		var resBody Response
		err = json.Unmarshal(body, &resBody)
		assert.NoError(t, err)
		return res.Code, resBody, ""
	}

	return res.Code, Response{}, strings.Trim(string(body), "\n")
}
