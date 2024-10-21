package rtsync

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
)

type File struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

type Response struct {
	Status string `json:"status"`
}

func (rts *realTimeSyncServer) apiHandler() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("POST /file", rts.createFileHandler)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Cors(middleware.CorsOptions{}),
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := Response{
		Status: "success",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
}
