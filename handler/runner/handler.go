package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/leslie-wang/clusterd/common/util"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
)

const (
	recordFilename = "index.m3u8"
	logFilename    = "ffmpeg.log"
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

	runningJobID int
	lock         *sync.Mutex

	cli *manager.Client
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) *Handler {
	h := &Handler{c: c, lock: &sync.Mutex{}}
	h.cli = manager.NewClient(c.MgrHost, c.MgrPort)
	return h
}

func (h *Handler) CreateRouter() *mux.Router {
	if h.r == nil {
		h.r = mux.NewRouter()
		h.r.HandleFunc(types.MkIDURLByBase(types.URLRunnerLogJob), h.jobLog).Methods(http.MethodGet)
	}
	return h.r
}

func (h *Handler) jobLog(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)[types.ID]

	dir := filepath.Join(h.c.Workdir, jobID)
	f, err := os.Open(filepath.Join(dir, logFilename))
	if err != nil {
		if os.IsNotExist(err) {
			util.WriteError(w, util.ErrNotExist)
		} else {
			util.WriteError(w, err)
		}
		return
	}
	defer f.Close()

	for {
		_, err := io.CopyN(w, f, 4096)
		if err == nil {
			continue
		}
		if err != io.EOF {
			util.WriteError(w, err)
			return
		}

		h.lock.Lock()
		if jobID != strconv.Itoa(h.runningJobID) {
			h.lock.Unlock()
			return
		}
		h.lock.Unlock()

		// still writing logs, so wait and retry
		time.Sleep(time.Second)
	}
}

func (h *Handler) Run(ctx context.Context) error {
	for {
		job, err := h.cli.AcquireJob(h.c.Name)
		if err != nil {
			log.Printf("Request job: %s", err)
		} else if job != nil {
			log.Printf("Run job: %v", job)
			h.lock.Lock()
			h.runningJobID = job.ID
			h.lock.Unlock()
			exitCode, err := h.runJob(ctx, job)
			if err != nil {
				log.Printf("Handle job %+v: %v", job, err)
			}
			h.lock.Lock()
			h.runningJobID = 0
			h.lock.Unlock()
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

	// sleep until start time
	var duration time.Duration
	if r.StartTime != nil {
		startTime := time.Unix(int64(*r.StartTime), 0)
		if startTime.Before(time.Now()) {
			return -1, fmt.Errorf("start time (%s) is earlier than now (%s)", startTime, time.Now())
		}
		time.Sleep(time.Until(startTime))
		duration = time.Duration(*r.EndTime-*r.StartTime) * time.Second
	} else {
		duration = time.Until(time.Unix(int64(*r.EndTime), 0))
	}

	runCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	dir := filepath.Join(h.c.Workdir, strconv.Itoa(j.ID))
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return -1, err
	}
	mediaFile := filepath.Join(dir, recordFilename)
	args := []string{"-i", r.SourceURL, "-c", "copy", "-hls_time", "10",
		"-hls_playlist_type", "vod", "-hls_segment_type", "fmp4", "-hls_segment_filename", "%d.m4s", mediaFile}
	fmt.Printf("---- ffmpeg %v\n", args)
	cmd := exec.CommandContext(runCtx, "ffmpeg", args...)
	cmd.Dir = dir

	logFile, err := os.Create(filepath.Join(dir, logFilename))
	if err != nil {
		return -1, err
	}
	defer logFile.Close()

	errChan := make(chan error)
	go func() {
		cmd.Stdout = logFile
		cmd.Stderr = logFile

		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%v: %s\n", args, err)
		} else if cmd.ProcessState == nil {
			err = fmt.Errorf("empty process state after run")
		}
		errChan <- err
	}()

	select {
	case err = <-errChan:
	case <-runCtx.Done():
		err = ctx.Err()
		if err == context.DeadlineExceeded {
			// recording is end now.
			err = nil
		}
	}

	return cmd.ProcessState.ExitCode(), err
}
