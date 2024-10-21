package rtsync

import (
	"net/http"
)

type File struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

func (rts *realTimeSyncServer) fileHandler(w http.ResponseWriter, r *http.Request) {

}
