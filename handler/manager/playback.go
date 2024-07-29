package manager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/common/hls"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

const (
	defaultIndexFile = "index.m3u8"
	defaultInitFile  = "init.mp4"
)

func (h *Handler) mkPlaybackURL(id int) string {
	return fmt.Sprintf("%s%s/%d/%s", h.cfg.BaseURL, types.URLPlay, id, defaultIndexFile)
}

func (h *Handler) mkDownloadURL(id int, filename string) string {
	if filename == "" {
		return fmt.Sprintf("%s%s/%d", h.cfg.BaseURL, types.URLDownload, id)
	}
	return fmt.Sprintf("%s%s/%d/%s", h.cfg.BaseURL, types.URLDownload, id, filename)
}

func (h *Handler) playback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars[types.ID]
	filename := vars["filename"]
	if filename == "" {
		filename = defaultIndexFile
	}

	http.ServeFile(w, r, filepath.Join(h.cfg.MediaDir, jobID, filename))
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars[types.ID]

	dir := filepath.Join(h.cfg.MediaDir, jobID)
	filename := vars["filename"]
	if filename == "" {
		filename = defaultIndexFile
	} else {
		filename = strings.Trim(filename, filepath.Ext(filename)) + ".m3u8"
	}
	mediaPL, err := hls.ParseMediaPlaylist(filepath.Join(dir, filename))
	if err != nil {
		util.WriteError(w, err)
		return
	}

	initFile := defaultInitFile
	if mediaPL.Map != nil && mediaPL.Map.URI != "" {
		initFile = mediaPL.Map.URI
	}

	f, err := os.Open(filepath.Join(dir, initFile))
	if err != nil {
		util.WriteError(w, err)
		return
	}
	defer f.Close()
	files := []*os.File{f}

	stat, err := f.Stat()
	if err != nil {
		util.WriteError(w, err)
		return
	}

	contentLength := stat.Size()

	for _, seg := range mediaPL.Segments {
		f, err := os.Open(filepath.Join(dir, seg.URI))
		if err != nil {
			util.WriteError(w, err)
			return
		}
		defer f.Close()
		files = append(files, f)

		stat, err = f.Stat()
		if err != nil {
			util.WriteError(w, err)
			return
		}
		contentLength += stat.Size()
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Content-Type", "video/mp4")

	for i, f := range files {
		_, err = io.Copy(w, f)
		if err != nil {
			h.logger.Warnf("Download recorded content %s's %d segmenbt: %s", jobID, i+1, err)
			return
		}
	}
}
