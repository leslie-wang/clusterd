package manager

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/types"
)

func (h *Handler) playback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars[types.ID]
	filename := vars["filename"]
	if filename == "" {
		filename = "index.m3u8"
	}

	http.ServeFile(w, r, filepath.Join(h.cfg.MediaDir, jobID, filename))
}
