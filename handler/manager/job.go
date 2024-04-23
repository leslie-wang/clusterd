package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
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
	jobID, err := strconv.Atoi(mux.Vars(r)[types.ID])
	if err != nil {
		util.WriteError(w, err)
		return
	}
	if err != nil {
		util.WriteError(w, err)
		return
	}

	job, err := h.jobDB.Get(jobID)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	cb, err := h.recordDB.GetCallbackRuleByRecordTaskID(job.RefID)
	if err != nil {
		log.Printf("WARN: retrieve job %d's callback info: %s", jobID, err)
	}

	status := &types.JobStatus{}
	err = json.NewDecoder(r.Body).Decode(status)
	if err != nil {
		util.WriteError(w, err)
		return
	}

	callbackURL := h.cfg.NotifyURL
	if cb != nil && cb.RecordStatusNotifyUrl != nil {
		callbackURL = *cb.RecordNotifyUrl
	}

	switch status.Type {
	case types.RecordJobStart:
		go notify(callbackURL, &types.LiveCallbackRecordStatusEvent{
			SessionID:   strconv.Itoa(jobID),
			RecordEvent: types.LiveRecordStatusStartSucceeded,
		})
	case types.RecordJobEnd:
		go notify(callbackURL, &types.LiveCallbackRecordStatusEvent{
			SessionID:   strconv.Itoa(jobID),
			RecordEvent: types.LiveRecordStatusEnded,
		})
		err = h.jobDB.CompleteAndArchive(int64(jobID), &status.ExitCode)
		if err != nil {
			util.WriteError(w, err)
			return
		}

		// TODO: save stdout and stderr
		util.WriteBody(w, status)
	case types.RecordJobException:
		go notify(callbackURL, &types.LiveCallbackRecordStatusEvent{
			SessionID:   strconv.Itoa(jobID),
			RecordEvent: types.LiveRecordStatusError,
		})
		err = h.jobDB.CompleteAndArchive(int64(jobID), &status.ExitCode)
		if err != nil {
			util.WriteError(w, err)
			return
		}

		// TODO: save stdout and stderr
		util.WriteBody(w, status)
	}
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	jobID, err := strconv.Atoi(mux.Vars(r)[types.ID])
	if err != nil {
		util.WriteError(w, err)
		return
	}

	job, err := h.jobDB.Get(jobID)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteBody(w, job)
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
