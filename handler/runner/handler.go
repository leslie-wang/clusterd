package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/bluenviron/gohlslib/pkg/playlist"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/common"
	"github.com/leslie-wang/clusterd/common/hls"
	"github.com/leslie-wang/clusterd/common/logger"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
	"github.com/pkg/errors"
)

const (
	recordFilename    = "index.m3u8"
	logStdoutFilename = "record-%d_out.log"
	logStderrFilename = "record-%d_err.log"
)

// Config is configuration for the handler
type Config struct {
	MgrHost  string
	MgrPort  uint
	Name     string
	Workdir  string
	Interval time.Duration

	LogDir       string
	MaxLogSize   int
	MaxLogBackup int
}

// Handler is structure for recorder API
type Handler struct {
	c Config
	r *mux.Router

	runningJobID int
	lock         *sync.Mutex

	logger *logger.Logger

	reportChan chan types.JobStatus
	cli        *manager.Client
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) (*Handler, error) {
	err := os.MkdirAll(c.LogDir, 0644)
	if err != nil {
		return nil, err
	}

	h := &Handler{c: c, lock: &sync.Mutex{}, reportChan: make(chan types.JobStatus),
		logger: logger.New(c.MaxLogSize, c.MaxLogBackup, filepath.Join(c.LogDir, "cd-runner.log")),
	}
	h.cli = manager.NewClient(c.MgrHost, c.MgrPort)

	go h.reportLoop()

	return h, nil
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
			h.logger.Infof("Request job: %s", err)
		} else if job != nil {
			h.logger.Infof("Run job: %v", job)
			h.lock.Lock()
			h.runningJobID = job.ID
			h.lock.Unlock()
			status, err := h.runJob(ctx, job)
			if err != nil {
				h.logger.Infof("Handle job %+v: %v", job, err)
			}
			if status == nil {
				goto wait
			}
			h.lock.Lock()
			h.runningJobID = 0
			h.lock.Unlock()
			err = h.cli.ReportJobStatus(status)
			if err != nil {
				h.logger.Infof("Report job %+v: %v", job, err)
			}
		} else {
			count++
			if count > int(5*time.Minute/h.c.Interval) {
				h.logger.Infof("No jobs in 5 minutes, sleep")
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
		r := &types.JobRecord{}
		err := json.Unmarshal([]byte(j.Metadata), r)
		if err != nil {
			return &types.JobStatus{
				ID:       j.ID,
				ExitCode: -1,
				Stdout:   err.Error(),
			}, err
		}

		if len(r.RecordStreams) == 0 || r.RecordStreams[0].SourceURL == "" {
			return &types.JobStatus{
				ID:       j.ID,
				ExitCode: -1,
			}, errors.New("Record source URL is empty")
		}

		runCtx, cancel := context.WithCancel(ctx)
		go func() {
			//pull status, and cancel job if it is deleted
			for {
				currentJob, err := h.cli.GetJob(j.ID)
				if err != nil {
					h.logger.Infof("get job ID failure: %s\n", err)
					goto sleep
				}
				if currentJob.EndTime == nil {
					// not finished, sleep
					goto sleep
				}

				cancel()
				return
			sleep:
				after := time.After(5 * time.Second)
				select {
				case <-after:
				case <-runCtx.Done():
					return
				}
			}
		}()
		return h.runRecordJob(runCtx, j.ID, r)
	}
	return nil, fmt.Errorf("unknown job category: %v", j)
}

const sdpTemplate = `SDP:
v=0
o=- 0 0 IN IP4 %s
s=No Name
c=IN IP4 %s
t=0 0
a=tool:libavformat 57.29.101
m=video %s RTP/AVP 96
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1
m=audio %s RTP/AVP 97
a=rtpmap:97 opus/48000/2
`

func (h *Handler) runRecordJob(ctx context.Context, id int, r *types.JobRecord) (*types.JobStatus, error) {
	var runCtx context.Context
	if r.EndTime != nil {
		var (
			duration time.Duration
			cancel   context.CancelFunc
		)
		if r.StartTime != nil {
			startTime := time.Unix(int64(*r.StartTime), 0)
			if startTime.Before(time.Now()) {
				h.logger.Warnf("start time (%s) is earlier than now (%s)", startTime, time.Now())
			} else {
				time.Sleep(time.Until(startTime))
			}
			duration = time.Duration(*r.EndTime-*r.StartTime) * time.Second
		} else {
			duration = time.Until(time.Unix(int64(*r.EndTime), 0))
		}
		runCtx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	} else {
		// no stop time until it is stopped by api
		runCtx = context.Background()
	}

	go h.addReport(types.JobStatus{ID: id, Type: types.RecordJobStart})

	storePath := r.StorePath
	if storePath == "" {
		storePath = h.c.Workdir
	}
	dir := common.MkStoragePath(storePath, strconv.Itoa(id))
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return &types.JobStatus{
			ID:       id,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	masterIndexFilename := filepath.Join(dir, recordFilename)

	var args []string
	sourceURL := r.RecordStreams[0].SourceURL
	if r.RecordTimeout > 0 {
		args = []string{"-rw_timeout", fmt.Sprintf("%d", r.RecordTimeout)}
	}

	if len(r.RecordStreams) > 1 {
		vu, err := url.Parse(r.RecordStreams[0].SourceURL)
		if err != nil {
			return nil, err
		}
		va, err := url.Parse(r.RecordStreams[1].SourceURL)
		if err != nil {
			return nil, err
		}
		sdp := fmt.Sprintf(sdpTemplate, vu.Hostname(), vu.Hostname(), vu.Port(), va.Port())
		fmt.Println(sdp)
		sdpFile, err := os.CreateTemp("", "sdp")
		if err != nil {
			return nil, err
		}
		sourceURL = sdpFile.Name()

		_, err = sdpFile.Write([]byte(sdp))
		if err != nil {
			return nil, err
		}
		err = sdpFile.Close()
		if err != nil {
			return nil, err
		}
		args = append(args, "-protocol_whitelist", "file,udp,rtp", "-i", sourceURL, "-vcodec", "copy",
			"-acodec", "aac", "-bsf:a", "aac_adtstoasc", "-hls_time", fmt.Sprintf("%d", r.HlsSegmentDuration),
			"-hls_playlist_type", "event", "-hls_segment_type", "fmp4", "-hls_segment_filename", "%d.m4s", masterIndexFilename)
	} else {
		args = append(args, "-i", sourceURL, "-c", "copy", "-bsf:a", "aac_adtstoasc",
			"-hls_time", fmt.Sprintf("%d", r.HlsSegmentDuration),
			"-hls_playlist_type", "event", "-hls_segment_type", "fmp4", "-hls_segment_filename", "%d.m4s", masterIndexFilename)
	}

	h.logger.Infof("record started: ffmpeg %v\n", args)
	cmd := exec.CommandContext(runCtx, "ffmpeg", args...)
	cmd.Dir = dir

	logoutFilename := filepath.Join(h.c.LogDir, fmt.Sprintf(logStdoutFilename, id))
	logoutFile, err := os.Create(logoutFilename)
	if err != nil {
		return &types.JobStatus{
			ID:       id,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	defer logoutFile.Close()

	logerrFilename := filepath.Join(h.c.LogDir, fmt.Sprintf(logStderrFilename, id))
	logerrFile, err := os.Create(logerrFilename)
	if err != nil {
		return &types.JobStatus{
			ID:       id,
			ExitCode: -1,
			Stdout:   err.Error(),
		}, err
	}
	defer logerrFile.Close()

	// start count record
	go h.generateIntermittentDownloadIndexFile(runCtx, r, id, dir, masterIndexFilename)

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
		err = runCtx.Err()
		if err == context.DeadlineExceeded {
			// recording is end now.
			h.logger.Warnf("recording is exceeding deadline")
			err = nil
		}
	}

	var duration, size uint64
	mediaPL, err := hls.ParseMediaPlaylist(masterIndexFilename)
	if err != nil {
		h.logger.Warnf("parse master index file %s: %s", masterIndexFilename, err)
	} else {
		duration = hls.CalculateDuration(mediaPL)
		size = hls.CalculateFileSize(dir, mediaPL, h.logger)
	}

	exitCode := cmd.ProcessState.ExitCode()
	if err == nil && exitCode == 0 {
		h.logger.Infof("recording finished")

		return &types.JobStatus{
			ID:       id,
			Type:     types.RecordJobEnd,
			Size:     size,
			Duration: duration,
		}, nil
	} else {
		h.logger.Infof("record exitcode: %d, err: %s", cmd.ProcessState.ExitCode(), err)
	}
	sout, err := os.ReadFile(logoutFilename)
	if err != nil {
		h.logger.Warnf("read stdout log file %s: %s", logoutFilename, err)
	}
	serr, err := os.ReadFile(logerrFilename)
	if err != nil {
		h.logger.Warnf("read stderr log file %s: %s", logerrFilename, err)
	}

	return &types.JobStatus{
		ID:       id,
		Type:     types.RecordJobException,
		ExitCode: exitCode,
		Stdout:   string(sout),
		Stderr:   string(serr),
		Size:     size,
		Duration: duration,
	}, nil
}

func (h *Handler) generateIntermittentDownloadIndexFile(ctx context.Context, r *types.JobRecord,
	id int, dir, masterIndexFilename string) {
	if r.Mp4FileDuration <= 0 {
		h.logger.Warnf("invalid mp4 record file duration: %ds", r.Mp4FileDuration)
		return
	}
	index := 0
	for {
		index++
		start := time.Now()
		after := time.After(time.Duration(r.Mp4FileDuration) * time.Second)
		select {
		case <-after:
		case <-ctx.Done():
			return
		}
		// create a new m3u8 file
		dlIndexFilename := fmt.Sprintf("dl%d.m3u8", index)
		lastIndexFilename := fmt.Sprintf("dl%d.m3u8", index-1)
		var (
			err     error
			content []byte
			mediaPL *playlist.Media
		)
		if index == 1 {
			fname := filepath.Join(dir, dlIndexFilename)
			f, err := os.Create(fname)
			if err != nil {
				h.logger.Warnf("create %s: %s", dlIndexFilename, err)
				continue
			}

			content, err = os.ReadFile(masterIndexFilename)
			if err != nil {
				h.logger.Warnf("read %s: %s", masterIndexFilename, err)
			} else {
				_, err = f.Write(content)
				if err != nil {
					h.logger.Warnf("copy %s into %s: %s", masterIndexFilename, dlIndexFilename, err)
				}
			}
			err = f.Close()
			if err != nil {
				h.logger.Warnf("close %s: %s", dlIndexFilename, err)
				continue
			}

			mediaPL, err = hls.ParseMediaPlaylist(fname)
			if err != nil {
				h.logger.Warnf("parse %s: %s", fname, err)
				continue
			}
		} else {
			// copy last media file
			mediaPL, err = h.trimSegments(masterIndexFilename, filepath.Join(dir, lastIndexFilename))
			if err != nil {
				h.logger.Warnf("trim segment for %s: %s", dlIndexFilename, err)
				continue
			}
			content, err = mediaPL.Marshal()
			if err != nil {
				h.logger.Warnf("marshal media playlist %s: %s", dlIndexFilename, content)
				continue
			}
			err = os.WriteFile(filepath.Join(dir, dlIndexFilename), content, 0755)
			if err != nil {
				h.logger.Warnf("write media playlist %s: %s", dlIndexFilename, content)
			}
		}

		h.logger.Infof("Generated mp4 recording index file %s", filepath.Join(dir, dlIndexFilename))

		timestampIndexFilename := fmt.Sprintf("%d-%d-%d-%d-%d-%d-%d", id,
			start.Year(), start.Month(), start.Day(),
			start.Hour(), start.Minute(), start.Second())

		err = os.Symlink(filepath.Join(dir, dlIndexFilename), filepath.Join(dir, timestampIndexFilename)+".m3u8")
		if err != nil {
			h.logger.Warnf("link file %s.m3u8: %s", timestampIndexFilename, err)
		}

		duration := hls.CalculateDuration(mediaPL)
		size := hls.CalculateFileSize(dir, mediaPL, h.logger)
		err = h.cli.ReportJobStatus(&types.JobStatus{
			ID:          id,
			Type:        types.RecordMp4FileCreated,
			Mp4Filename: timestampIndexFilename + ".mp4",
			Duration:    duration,
			Size:        size,
		})
		if err != nil {
			h.logger.Warnf("report mp4 file %s creation: %s", dlIndexFilename, content)
		}
	}
}

func (h *Handler) trimSegments(currentPl, lastPl string) (*playlist.Media, error) {
	currMediaPL, err := hls.ParseMediaPlaylist(currentPl)
	if err != nil {
		return nil, err
	}

	lastMediaPl, err := hls.ParseMediaPlaylist(lastPl)
	if err != nil {
		return nil, err
	}

	// find the last media segment uri
	lastSegmentURI := lastMediaPl.Segments[len(lastMediaPl.Segments)-1].URI

	startIndex := 0
	for i, s := range currMediaPL.Segments {
		if s.URI == lastSegmentURI {
			startIndex = i + 1
			break
		}
	}
	h.logger.Infof("last download segment: %s, new download segments %s -> %s", lastSegmentURI,
		currMediaPL.Segments[startIndex].URI,
		currMediaPL.Segments[len(currMediaPL.Segments)-1].URI)

	if startIndex == 0 {
		return nil, errors.Errorf("unable to find last segment URI '%s' in current playlist", lastSegmentURI)
	}

	lastMediaPl.Segments = currMediaPL.Segments[startIndex:]
	return lastMediaPl, nil
}
