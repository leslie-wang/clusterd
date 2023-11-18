package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
)

const (
	recordFilename = "index.m3u8"
)

// Config is configuration for the handler
type Config struct {
	MgrHost  string
	MgrPort  uint
	Name     string
	Workdir  string
	Interval time.Duration
}

// Handler is structure for recorder API
type Handler struct {
	c Config
	r *mux.Router

	cli *manager.Client
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) *Handler {
	h := &Handler{c: c}
	h.cli = manager.NewClient(c.MgrHost, c.MgrPort)
	return h
}

func (h *Handler) CreateRouter() *mux.Router {
	if h.r == nil {
		h.r = mux.NewRouter()
		//h.r.HandleFunc(types.RecordStartURL, h.start).Methods(http.MethodPost)
	}
	return h.r
}

func (h *Handler) Run(ctx context.Context) error {
	for {
		job, err := h.cli.AcquireJob(h.c.Name)
		if err != nil {
			log.Printf("Request job: %s", err)
		} else if job != nil {
			log.Printf("Run job: %v", job)
			exitCode, err := h.runJob(ctx, job)
			if err != nil {
				log.Printf("Handle job %+v: %v", job, err)
			}
			err = h.cli.ReportJob(job.ID, exitCode)
			if err != nil {
				log.Printf("Report job %+v: %v", job, err)
			}
		} else {
			log.Printf("No jobs, sleep %s", h.c.Interval)
		}
		after := time.After(h.c.Interval)
		select {
		case <-after:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *Handler) runJob(ctx context.Context, j *types.Job) (int, error) {
	/*
		if len(j.Commands) == 0 {
			return errors.Errorf("Empty commands in job %d", j.ID)
		}
		name := j.Commands[0]
		var arg []string
		if len(j.Commands) > 1 {
			arg = j.Commands[1:]
		}
		content, err := exec.Command(name, arg...).CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "%s", string(content))
		}

	*/
	if j.Category == types.CategoryRecord {
		return h.runRecordJob(ctx, j)
	}
	return -1, fmt.Errorf("unknown job category: %v", j)
}

func (h *Handler) runRecordJob(ctx context.Context, j *types.Job) (int, error) {
	r := &types.JobRecord{}
	err := json.Unmarshal([]byte(j.Metadata), r)
	if err != nil {
		return -1, err
	}

	dir := filepath.Join(h.c.Workdir, strconv.Itoa(j.ID))
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return -1, err
	}
	filename := filepath.Join(dir, recordFilename)
	args := []string{"-re", "-i", r.SourceURL, "-c", "copy",
		"-hls_playlist_type", "vod", "-hls_segment_type", "fmp4", "-hls_segment_filename", "%d.m4s", filename}
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = dir
	content, err := cmd.CombinedOutput()
	if err != nil {
		return -1, fmt.Errorf("%v: %s\n%s", args, err, string(content))
	}
	return 0, nil
}
