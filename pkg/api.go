package rtsync

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
)

type CreateFileBody struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

type FileWithContent struct {
	repository.File
	Content []byte `json:"content"`
}

const (
	ErrInvalidFile     = "impossilbe to create file"
	ErrReadingFile     = "impossilbe to read file"
	ErrNotExistingFile = "not existing file"
)

func (rts *realTimeSyncServer) apiHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("GET /file", rts.listFilesHandler)
	router.HandleFunc("GET /file/{id}", rts.fetchFileHandler)
	router.HandleFunc("POST /file", rts.createFileHandler)
	router.HandleFunc("DELETE /file/{id}", rts.deleteFileHandler)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Cors(middleware.CorsOptions{}),
		middleware.IsAuthenticated(middleware.AuthOptions{SecretKey: rts.jwtSecret}),
	)

	routerWithStack := stack(router)
	return routerWithStack
}

func (rts *realTimeSyncServer) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.WorkspaceIDFromCtx(r.Context())

	files, err := rts.db.FetchFiles(r.Context(), workspaceID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(files); err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
}

func (rts *realTimeSyncServer) fetchFileHandler(w http.ResponseWriter, r *http.Request) {
	fileId, err := strconv.Atoi(r.PathValue("id"))

	if fileId == 0 || err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	file, err := rts.db.FetchFile(r.Context(), int64(fileId))
	if err != nil {
		http.Error(w, ErrNotExistingFile, http.StatusNotFound)
		return
	}

	workspaceID := middleware.WorkspaceIDFromCtx(r.Context())
	if file.WorkspaceID != workspaceID {
		http.Error(w, ErrNotExistingFile, http.StatusNotFound)
		return
	}

	fileContent, err := rts.storage.ReadObject(file.DiskPath)
	if err != nil {
		http.Error(w, ErrReadingFile, http.StatusInternalServerError)
		return
	}

	fileWithContent := FileWithContent{
		File:    file,
		Content: fileContent,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(fileWithContent); err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
}

func (rts *realTimeSyncServer) createFileHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	var data CreateFileBody
	if err = json.Unmarshal(body, &data); err != nil {
		http.Error(w, "error parsing JSON", http.StatusBadRequest)
		return
	}

	diskPath, err := rts.storage.CreateObject(data.Content)
	if err != nil {
		http.Error(w, ErrInvalidFile, http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(data.Content)
	workspaceID := middleware.WorkspaceIDFromCtx(r.Context())

	file, err := rts.db.AddFile(r.Context(), repository.AddFileParams{
		DiskPath:      diskPath,
		WorkspacePath: data.Path,
		MimeType:      mimeType,
		Hash:          filestorage.GenerateHash(data.Content),
		WorkspaceID:   workspaceID,
	})

	if err != nil {
		http.Error(w, ErrInvalidFile, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(file); err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
}

func (rts *realTimeSyncServer) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	fileId, err := strconv.Atoi(r.PathValue("id"))

	if fileId == 0 || err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	file, err := rts.db.FetchFile(r.Context(), int64(fileId))
	if err != nil {
		http.Error(w, ErrNotExistingFile, http.StatusNotFound)
		return
	}

	workspaceID := middleware.WorkspaceIDFromCtx(r.Context())
	if file.WorkspaceID != workspaceID {
		http.Error(w, ErrNotExistingFile, http.StatusNotFound)
		return
	}

	if err := rts.storage.DeleteObject(file.DiskPath); err != nil {
		http.Error(w, ErrNotExistingFile, http.StatusInternalServerError)
		return
	}

	err = rts.db.DeleteFile(r.Context(), int64(fileId))
	if err != nil {
		http.Error(w, ErrInvalidFile, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
