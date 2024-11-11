package rtsync

import (
	"context"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/internal/testutils"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)

	file, err := repo.AddFile(context.Background(), repository.AddFileParams{
		DiskPath:      "disk_path",
		WorkspacePath: "workspace_path",
		MimeType:      "text/plain; charset=utf-8",
		Hash:          "123",
		WorkspaceID:   1,
	})
	require.NoError(t, err)

	fileContent := []byte("hello world!")
	mockFileStorage.On("CreateObject", fileContent).Return(file.DiskPath, nil)
	mockFileStorage.On("ReadObject", file.DiskPath).Return(fileContent, nil)

	_, err = mockFileStorage.CreateObject(fileContent)
	require.NoError(t, err)

	var server *realTimeSyncServer
	require.NotPanics(t, func() {
		options := Options{JWTSecret: []byte("secret")}
		server = New(repo, mockFileStorage, options)
	})
	t.Cleanup(func() { server.Close() })

	assert.Len(t, server.files, 1)
	assert.Equal(t, string(fileContent), server.files[file.ID])
}
