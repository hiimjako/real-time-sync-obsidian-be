package rtsync

import (
	"context"
	"net/http"
	"strconv"
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

	diskPath := "/foo/bar"
	mockFileStorage.On("CreateObject", data.Content).Return(diskPath, nil)

	res, body := testutils.DoRequest[repository.File](
		t,
		server,
		http.MethodPost,
		PathHttpApi+"/file",
		data,
		testutils.WithAuthHeader(options.JWTSecret, workspaceID),
	)

	// check response
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, repository.File{
		ID:            1,
		DiskPath:      diskPath,
		WorkspacePath: data.Path,
		MimeType:      "text/plain; charset=utf-8",
		Hash:          filestorage.GenerateHash(data.Content),
		CreatedAt:     body.CreatedAt,
		UpdatedAt:     body.UpdatedAt,
		WorkspaceID:   workspaceID,
	}, body)

	// check db
	files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, repository.File{
		ID:            1,
		DiskPath:      diskPath,
		WorkspacePath: data.Path,
		MimeType:      "text/plain; charset=utf-8",
		Hash:          filestorage.GenerateHash(data.Content),
		CreatedAt:     files[0].CreatedAt,
		UpdatedAt:     files[0].UpdatedAt,
		WorkspaceID:   workspaceID,
	}, files[0])

	// check mock assertions
	mockFileStorage.AssertCalled(t, "CreateObject", data.Content)
}

// Test_deleteFileHandler tests the deleteFileHandler using mocked storage
func Test_deleteFileHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	t.Cleanup(func() { server.Close() })

	t.Run("successfully delete a file", func(t *testing.T) {
		workspaceID := int64(10)
		data := File{
			Path:    "/home/file",
			Content: []byte("here a new file!"),
		}

		diskPath := "/foo/bar"
		mockFileStorage.On("CreateObject", data.Content).Return(diskPath, nil)
		mockFileStorage.On("DeleteObject", diskPath).Return(nil)

		// creating file
		res, createBody := testutils.DoRequest[repository.File](
			t,
			server,
			http.MethodPost,
			PathHttpApi+"/file",
			data,
			testutils.WithAuthHeader(options.JWTSecret, workspaceID),
		)
		assert.Equal(t, http.StatusCreated, res.Code)

		// deleting a file
		res, deleteBody := testutils.DoRequest[string](
			t,
			server,
			http.MethodDelete,
			PathHttpApi+"/file/"+strconv.Itoa(int(createBody.ID)),
			data,
			testutils.WithAuthHeader(options.JWTSecret, workspaceID),
		)
		assert.Equal(t, http.StatusNoContent, res.Code)
		assert.Equal(t, "", deleteBody)

		// check db
		files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
		assert.NoError(t, err)
		assert.Len(t, files, 0)

		// check mock assertions
		mockFileStorage.AssertCalled(t, "CreateObject", data.Content)
		mockFileStorage.AssertCalled(t, "DeleteObject", diskPath)
	})

	t.Run("unauthorize to delete a file of other workspace", func(t *testing.T) {
		workspaceID := int64(10)
		data := File{
			Path:    "/home/file/2",
			Content: []byte("here a new file!"),
		}

		diskPath := "/foo/bar/2"
		mockFileStorage.On("CreateObject", data.Content).Return(diskPath, nil)

		// creating file
		res, createBody := testutils.DoRequest[repository.File](
			t,
			server,
			http.MethodPost,
			PathHttpApi+"/file",
			data,
			testutils.WithAuthHeader(options.JWTSecret, workspaceID),
		)
		assert.Equal(t, http.StatusCreated, res.Code)

		// deleting a file
		anotherWorkspaceID := int64(20)
		res, deleteBody := testutils.DoRequest[string](
			t,
			server,
			http.MethodDelete,
			PathHttpApi+"/file/"+strconv.Itoa(int(createBody.ID)),
			data,
			testutils.WithAuthHeader(options.JWTSecret, anotherWorkspaceID),
		)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Equal(t, ErrNotExistingFile, deleteBody)

		// check db
		files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
		assert.NoError(t, err)
		assert.Len(t, files, 1)

		// check mock assertions
		mockFileStorage.AssertCalled(t, "CreateObject", data.Content)
		mockFileStorage.AssertNotCalled(t, "DeleteObject")
	})
}
