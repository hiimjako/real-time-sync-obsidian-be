package rtsync

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/internal/testutils"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// Test_createFileHandler tests the createFileHandler using mocked storage
func Test_createFileHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	workspaceID := int64(10)
	data := File{
		Path:    "/home/file",
		Content: []byte("here a new file!"),
	}

	virtualPath := "foo/bar"
	mockFileStorage.On("CreateObject", data.Path, data.Content).Return(virtualPath, nil)

	reqBody, err := json.Marshal(data)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, PathHttpApi+"/file", bytes.NewBuffer(reqBody))
	require.NoError(t, testutils.CreateAuthHeader(req, options.JWTSecret, workspaceID))

	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var resBody Response
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &resBody)
	assert.NoError(t, err)

	assert.Equal(t, 201, res.Code)
	assert.Equal(t, Response{Status: "success"}, resBody)

	files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, repository.FetchWorkspaceFilesRow{
		ID:          files[0].ID,
		Path:        "/home/file",
		VirtualPath: virtualPath,
		MimeType:    "text/plain; charset=utf-8",
		Hash:        filestorage.CalculateHash(data.Content),
		CreatedAt:   files[0].CreatedAt,
		UpdatedAt:   files[0].UpdatedAt,
	}, files[0])

	mockFileStorage.AssertCalled(t, "CreateObject", data.Path, data.Content)

	t.Cleanup(func() {
		server.Close()
	})
}
