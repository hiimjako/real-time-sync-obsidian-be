package rtsync

import (
	"net/http"

	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
)

type File struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

func (rts *realTimeSyncServer) apiHandler(w http.ResponseWriter, r *http.Request) {
	const apiPrefix = "/api"

	router := http.NewServeMux()
	router.HandleFunc("POST /file", rts.fileHandler)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.Cors(middleware.CorsOptions{}),
	)

	routerWithStack := stack(router)
	routerWithStripPrefix := http.StripPrefix(apiPrefix, routerWithStack)

	api := http.NewServeMux()
	api.Handle(apiPrefix+"/", routerWithStripPrefix)
}

func (rts *realTimeSyncServer) fileHandler(w http.ResponseWriter, r *http.Request) {

}
