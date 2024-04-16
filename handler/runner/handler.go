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
	recordFilename    = "index.m3u8"
	logStdoutFilename = "record_out.log"
	logStderrFilename = "record_err.log"
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

	reportChan chan types.JobStatus
	cli        *manager.Client
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) *Handler {
	h := &Handler{c: c, lock: &sync.Mutex{}, reportChan: make(chan types.JobStatus)}
	h.cli = manager.NewClient(c.MgrHost, c.MgrPort)

	go h.reportLoop()

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
	f, err := os.Open(filepath.Join(dir, logStdoutFilename))
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
	count := 0
	for {
		job, err := h.cli.AcquireJob(h.c.Name)
		if err != nil {
			log.Printf("Request job: %s", err)
		} else if job != nil {
			log.Printf("Run job: %v", job)
			h.lock.Lock()
			h.runningJobID = job.ID
			h.lock.Unlock()
			status, err := h.runJob(ctx, job)
			if err != nil {
				log.Printf("Handle job %+v: %v", job, err)
			}
			if status == nil {
				goto wait
			}
			h.lock.Lock()
			h.runningJobID = 0
			h.lock.Unlock()
			err = h.cli.ReportJobStatus(status)
			if err != nil {
				log.Printf("Report job %+v: %v", job, err)
			}
		} else {
			count++
			if count > int(5*time.Minute/h.c.Interval) {
				log.Printf("No jobs in 5 minutes, sleep")
				count = 0
			}
		}

	wait:
		after := time.After(h.c.Interval)
		select {
		case <-after:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *Handler) runJob(ctx context.Context, j *types.Job) (*types.JobStatus, error) {
	if j.Category == types.CategoryRecord {
		runCtx, cancel := context.WithCancel(ctx)
		go func() {
			//pull status, and cancel job if it is deleted
			for {
				currentJob, err := h.cli.GetJob(j.ID)
				if err != nil {
					log.Printf("get job ID failure: %s\n", err)
					goto sleep
				}
				if currentJob.EndTime == nil {
					// not finished, sleep
					goto sleep
				}
				if currentJob.ExitCode == nil {
					cancel()
				}
				return
			sleep:
				time.Sleep(5 * time.Second)
			}
		}()
		return h.runRecordJob(runCtx, j)
	}
	return nil, fmt.Errorf("unknown job category: %v", j)
}

func (h *Handler) runRecordJob(ctx context.Context, j *types.Job) (*types.JobStatus, error) {
	r := &types.JobRecord{}
	err := json.Unmarshal([]byte(j.Metadata), r)
	if err != nil {
		return &types.JobStatus{
			ID:       j.ID,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}

	// sleep until start time
	var duration time.Duration
	if r.StartTime != nil {
		startTime := time.Unix(int64(*r.StartTime), 0)
		if startTime.Before(time.Now()) {
			log.Printf("start time (%s) is earlier than now (%s)", startTime, time.Now())
		} else {
			time.Sleep(time.Until(startTime))
		}
		duration = time.Duration(*r.EndTime-*r.StartTime) * time.Second
	} else {
		duration = time.Until(time.Unix(int64(*r.EndTime), 0))
	}

	go h.addReport(types.JobStatus{ID: j.ID, Type: types.RecordJobStart})

	runCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	storePath := r.StorePath
	if storePath == "" {
		storePath = h.c.Workdir
	}
	dir := filepath.Join(storePath, strconv.Itoa(j.ID))
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return &types.JobStatus{
			ID:       j.ID,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	mediaFile := filepath.Join(dir, recordFilename)
	args := []string{"-i", r.SourceURL, "-c", "copy", "-hls_time", "10",
		"-hls_playlist_type", "vod", "-hls_segment_type", "fmp4", "-hls_segment_filename", "%d.m4s", mediaFile}
	fmt.Printf("record started: ffmpeg %v\n", args)
	cmd := exec.CommandContext(runCtx, "ffmpeg", args...)
	cmd.Dir = dir

	logoutFilename := filepath.Join(dir, logStdoutFilename)
	logoutFile, err := os.Create(logoutFilename)
	if err != nil {
		return &types.JobStatus{
			ID:       j.ID,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	defer logoutFile.Close()

	logerrFilename := filepath.Join(dir, logStderrFilename)
	logerrFile, err := os.Create(logerrFilename)
	if err != nil {
		return &types.JobStatus{
			ID:       j.ID,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	defer logerrFile.Close()

	errChan := make(chan error)
	go func() {
		cmd.Stdout = logoutFile
		cmd.Stderr = logerrFile

		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%v: %s", args, err)
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

	exitCode := cmd.ProcessState.ExitCode()
	if err == nil && exitCode == 0 {
		log.Printf("recording finished")
		return &types.JobStatus{ID: j.ID, Type: types.RecordJobEnd}, nil
	}
	sout, err := os.ReadFile(logoutFilename)
	if err != nil {
		log.Printf("WARN: read stdout log file %s: %s", logoutFilename, err)
	}
	serr, err := os.ReadFile(logerrFilename)
	if err != nil {
		log.Printf("WARN: read stderr log file %s: %s", logerrFilename, err)
	}

	return &types.JobStatus{
		ID:       j.ID,
		Type:     types.RecordJobException,
		ExitCode: exitCode,
		Stdout:   string(sout),
		Stderr:   string(serr),
	}, nil
}
