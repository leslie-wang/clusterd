package manager

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request) {
	job := &types.Job{}
	err := json.NewDecoder(r.Body).Decode(job)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	job.CreateTime = time.Now()
	err = h.insertJobDB(job)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	util.WriteBody(w, job)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.listJobsDB()
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

	err = h.archiveJobTx(job.ID, job.ExitCode)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	// TODO: save stdout and stderr
	util.WriteBody(w, job)
	return
}
