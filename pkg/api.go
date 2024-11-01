package rtsync

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/repository"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/filestorage"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
)

type File struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

const (
	ErrInvalidFile = "impossilbe to create file"
)

func (rts *realTimeSyncServer) apiHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("POST /file", rts.createFileHandler)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Cors(middleware.CorsOptions{}),
		middleware.IsAuthenticated(middleware.AuthOptions{SecretKey: rts.jwtSecret}),
	)

	routerWithStack := stack(router)
	return routerWithStack
}

func (rts *realTimeSyncServer) createFileHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	var data File
	if err = json.Unmarshal(body, &data); err != nil {
		http.Error(w, "error parsing JSON", http.StatusBadRequest)
		return
	}

	virtualPath, err := rts.storage.CreateObject(data.Content)
	if err != nil {
		http.Error(w, ErrInvalidFile, http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(data.Content)
	workspaceID := middleware.WorkspaceIDFromCtx(r.Context())

	file, err := rts.db.AddFile(r.Context(), repository.AddFileParams{
		Path:        data.Path,
		VirtualPath: virtualPath,
		MimeType:    mimeType,
		Hash:        filestorage.GenerateHash(data.Content),
		WorkspaceID: workspaceID,
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
