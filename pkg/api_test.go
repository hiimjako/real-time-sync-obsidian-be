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

// Test_listFilesHandler tests the listFileHandler using mocked storage
func Test_listFilesHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	t.Cleanup(func() { server.Close() })

	workspaceID := int64(10)
	filesToInsert := []struct {
		file        CreateFileBody
		workspaceID int64
	}{
		{
			file:        CreateFileBody{Path: "/home/file/1", Content: []byte("here a new file!")},
			workspaceID: workspaceID,
		},
		{
			file:        CreateFileBody{Path: "/home/file/2", Content: []byte("here a new file 2!")},
			workspaceID: workspaceID,
		},
		{
			file:        CreateFileBody{Path: "/home/file/3", Content: []byte("here a new file 3!")},
			workspaceID: 123,
		},
	}

	for _, f := range filesToInsert {
		mockFileStorage.On("CreateObject", f.file.Content).Return(f.file.Path, nil)

		res, _ := testutils.DoRequest[repository.File](
			t,
			server,
			http.MethodPost,
			PathHttpApi+"/file",
			f.file,
			testutils.WithAuthHeader(options.JWTSecret, f.workspaceID),
		)
		assert.Equal(t, http.StatusCreated, res.Code)
	}

	// fetch files
	res, body := testutils.DoRequest[[]repository.File](
		t,
		server,
		http.MethodGet,
		PathHttpApi+"/file",
		nil,
		testutils.WithAuthHeader(options.JWTSecret, workspaceID),
	)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Len(t, body, 2)

	// check mock assertions
	mockFileStorage.AssertNumberOfCalls(t, "CreateObject", len(filesToInsert))
}

// Test_fetchFileHandler tests the fetchFileHandler using mocked storage
func Test_fetchFileHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	t.Cleanup(func() { server.Close() })

	workspaceID := int64(10)
	filesToInsert := []struct {
		file        CreateFileBody
		workspaceID int64
	}{
		{
			file:        CreateFileBody{Path: "/home/file/1", Content: []byte("here a new file!")},
			workspaceID: 123,
		},
		{
			file:        CreateFileBody{Path: "/home/file/2", Content: []byte("here a new file 2!")},
			workspaceID: workspaceID,
		},
	}

	for _, f := range filesToInsert {
		mockFileStorage.On("CreateObject", f.file.Content).Return(f.file.Path, nil)

		res, _ := testutils.DoRequest[repository.File](
			t,
			server,
			http.MethodPost,
			PathHttpApi+"/file",
			f.file,
			testutils.WithAuthHeader(options.JWTSecret, f.workspaceID),
		)
		assert.Equal(t, http.StatusCreated, res.Code)
	}

	// file of other workspace
	res, _ := testutils.DoRequest[string](
		t,
		server,
		http.MethodGet,
		PathHttpApi+"/file/1",
		nil,
		testutils.WithAuthHeader(options.JWTSecret, workspaceID),
	)
	assert.Equal(t, http.StatusNotFound, res.Code)

	// fetch file
	mockFileStorage.On("ReadObject", filesToInsert[1].file.Path).Return(filesToInsert[1].file.Content, nil)

	res, body := testutils.DoRequest[FileWithContent](
		t,
		server,
		http.MethodGet,
		PathHttpApi+"/file/2",
		nil,
		testutils.WithAuthHeader(options.JWTSecret, workspaceID),
	)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, FileWithContent{
		File: repository.File{
			ID:            2,
			DiskPath:      "/home/file/2",
			WorkspacePath: "/home/file/2",
			MimeType:      "text/plain; charset=utf-8",
			Hash:          body.Hash,
			CreatedAt:     body.CreatedAt,
			UpdatedAt:     body.UpdatedAt,
			WorkspaceID:   workspaceID,
		},
		Content: "here a new file 2!",
	}, body)

	// check mock assertions
	mockFileStorage.AssertNumberOfCalls(t, "CreateObject", len(filesToInsert))
	mockFileStorage.AssertCalled(t, "ReadObject", "/home/file/2")
}

// Test_createFileHandler tests the createFileHandler using mocked storage
func Test_createFileHandler(t *testing.T) {
	mockFileStorage := new(filestorage.MockFileStorage)
	db := testutils.CreateDB(t)
	repo := repository.New(db)
	options := Options{JWTSecret: []byte("secret")}
	server := New(repo, mockFileStorage, options)

	t.Cleanup(func() { server.Close() })

	t.Run("should create a file", func(t *testing.T) {
		workspaceID := int64(10)
		data := CreateFileBody{
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
	})

	t.Run("should not insert duplicate paths", func(t *testing.T) {
		workspaceID := int64(10)
		data := CreateFileBody{
			Path:    "/home/file",
			Content: []byte("here a new file!"),
		}

		diskPath := "/foo/bar"
		mockFileStorage.On("CreateObject", data.Content).Return(diskPath, nil)

		res, body := testutils.DoRequest[string](
			t,
			server,
			http.MethodPost,
			PathHttpApi+"/file",
			data,
			testutils.WithAuthHeader(options.JWTSecret, workspaceID),
		)

		// check response
		assert.Equal(t, http.StatusConflict, res.Code)
		assert.Equal(t, ErrDuplicateFile, body)

		// check db
		files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
		assert.NoError(t, err)
		assert.Len(t, files, 1)

		// check mock assertions
		mockFileStorage.AssertCalled(t, "CreateObject", data.Content)
	})
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
		data := CreateFileBody{
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
		data := CreateFileBody{
			Path:    "/home/file/2",
			Content: []byte("here a new file!"),
		}

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
		assert.Equal(t, http.StatusNotFound, res.Code)
		assert.Equal(t, ErrNotExistingFile, deleteBody)

		// check db
		files, err := repo.FetchWorkspaceFiles(context.Background(), workspaceID)
		assert.NoError(t, err)
		assert.Len(t, files, 1)

		// check mock assertions
		mockFileStorage.AssertNotCalled(t, "DeleteObject")
	})
}
