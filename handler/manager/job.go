package manager

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
	"net/http"
	"time"
)

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobDB.List()
	if err != nil {
		util.WriteError(w, err)
		return
	}

	util.WriteBody(w, jobs)
}

func (h *Handler) reportJob(w http.ResponseWriter, r *http.Request) {
	job := &types.JobResult{}
	err := json.NewDecoder(r.Body).Decode(job)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	err = h.jobDB.CompleteAndArchive(job.ID, job.ExitCode)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	// TODO: save stdout and stderr
	util.WriteBody(w, job)
	return
}

func (h *Handler) acquireJob(w http.ResponseWriter, r *http.Request) {
	runner := mux.Vars(r)[types.ID]

	job, err := h.jobDB.Acquire(runner, time.Now().Add(h.cfg.ScheduleInterval))
	if err != nil {
		util.WriteError(w, err)
		return
	}

	if job != nil {
		util.WriteBody(w, job)
	}
}
