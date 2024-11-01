package rtsync

import (
	"context"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/internal/testutils"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
)

// Test_createFileHandler tests the createFileHandler using mocked storage
func Test_createFileHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	t.Cleanup(func() { server.Close() })

	workspaceID := int64(10)
	data := File{
		Path:    "/home/file",
		Content: []byte("here a new file!"),
	}

	virtualPath := "foo/bar"
	mockFileStorage.On("CreateObject", data.Path, data.Content).Return(virtualPath, nil)

	res, body := testutils.DoRequest[repository.File](
		t,
		server,
		PathHttpApi+"/file",
		data,
		testutils.WithAuthHeader(options.JWTSecret, workspaceID),
	)

	// check response
	assert.Equal(t, 201, res.Code)
	assert.Equal(t, repository.File{
		ID:          1,
		Path:        data.Path,
		VirtualPath: virtualPath,
		MimeType:    "text/plain; charset=utf-8",
		Hash:        filestorage.GenerateHash(data.Content),
		CreatedAt:   body.CreatedAt,
		UpdatedAt:   body.UpdatedAt,
		WorkspaceID: workspaceID,
	}, body)

	// check db
	files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, repository.FetchWorkspaceFilesRow{
		ID:          1,
		Path:        data.Path,
		VirtualPath: virtualPath,
		MimeType:    "text/plain; charset=utf-8",
		Hash:        filestorage.GenerateHash(data.Content),
		CreatedAt:   files[0].CreatedAt,
		UpdatedAt:   files[0].UpdatedAt,
	}, files[0])

	// check mock assertions
	mockFileStorage.AssertCalled(t, "CreateObject", data.Path, data.Content)
}
