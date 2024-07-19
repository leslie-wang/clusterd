package manager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/leslie-wang/clusterd/common/model"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

var recordSuccess = 0

func (h *Handler) handleListRecordTasks() (*model.DescribeRecordTaskResponse, error) {
	list, err := h.recordDB.ListRecordTasks(context.Background())
	if err != nil {
		return nil, err
	}

	return &model.DescribeRecordTaskResponse{
		Response: &model.DescribeRecordTaskResponseParams{
			TaskList: list,
		},
	}, nil
}

func (h *Handler) handleDeleteRecordTask(q url.Values) (*model.DeleteLiveRecordRuleResponse, error) {
	tid := q.Get(TaskID)
	if tid == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	id, err := strconv.ParseInt(tid, 10, 64)
	if err != nil {
		return nil, err
	}

	job, err := h.jobDB.Get(int(id))
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, util.ErrNotExist
	}

	callbackURL := h.getCallbackURL(job)

	tx, err := h.newTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	err = h.recordDB.RemoveRecordTask(tx, id)
	if err != nil {
		return nil, err
	}

	err = h.jobDB.CompleteAndArchiveWithTx(tx, id, &recordSuccess)
	if err != nil {
		return nil, err
	}

	go notify(callbackURL, tid, &types.LiveCallbackRecordStatusEvent{
		SessionID:   tid,
		RecordEvent: types.LiveRecordStatusEnded,
		DownloadURL: h.mkDownloadURL(int(id), ""),
	})

	return &model.DeleteLiveRecordRuleResponse{Response: &model.DeleteLiveRecordRuleResponseParams{}}, tx.Commit()
}

func (h *Handler) handleCreateRecordTask(q url.Values, request io.ReadCloser) (*model.CreateRecordTaskResponse, error) {
	defer request.Close()

	var err error
	task := &types.LiveRecordTask{}
	if h.cfg.ParamQuery {
		r, err := h.parseRecordTask(q)
		if err != nil {
			return nil, err
		}
		if r.EndTime == nil {
			return nil, errors.New("EndTime can not be empty")
		}
		task.CreateRecordTaskRequestParams = r
	} else {
		task.CreateRecordTaskRequestParams = &model.CreateRecordTaskRequestParams{}
		err = json.NewDecoder(request).Decode(task)
		if err != nil {
			return nil, err
		}
	}

	if len(task.RecordStreams) == 0 || task.RecordStreams[0].SourceURL == "" {
		return nil, errors.New("SourceURL can not be empty")
	}

	tx, err := h.newTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if task.DomainName == nil {
		return nil, errors.New("DomainName can not be empty")
	}

	id, err := h.recordDB.InsertRecordTask(tx, task)
	if err != nil {
		return nil, err
	}

	record := &types.JobRecord{
		RecordStreams:   task.RecordStreams,
		NotifyURL:       task.NotifyURL,
		StorePath:       task.StorePath,
		EndTime:         task.EndTime,
		Mp4FileDuration: task.Mp4FileDuration,
	}
	if task.StartTime != nil && *task.StartTime != 0 {
		record.StartTime = task.StartTime
	}
	content, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	job := &types.Job{
		RefID:    id,
		Category: types.CategoryRecord,
		Metadata: string(content),
	}

	if task.StartTime != nil {
		st := time.Unix(int64(*task.StartTime), 0)
		job.ScheduleTime = &st
	} else {
		now := time.Now()
		job.ScheduleTime = &now
	}

	err = h.jobDB.Insert(tx, job)
	if err != nil {
		return nil, err
	}

	tid := strconv.FormatInt(id, 10)
	playbackURL := h.mkPlaybackURL(int(id))
	return &model.CreateRecordTaskResponse{Response: &model.CreateRecordTaskResponseParams{
		TaskId:      &tid,
		PlaybackURL: &playbackURL,
	}}, tx.Commit()
}

func (h *Handler) parseRecordTask(q url.Values) (*model.CreateRecordTaskRequestParams, error) {
	r := &model.CreateRecordTaskRequestParams{}

	domainName := q.Get(DomainName)
	r.DomainName = &domainName

	appName := q.Get(AppName)
	r.AppName = &appName

	streamName := q.Get(StreamName)
	r.StreamName = &streamName

	val := q.Get(StartTime)
	if val != "" {
		data, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		r.StartTime = &data
	}

	val = q.Get(EndTime)
	if val != "" {
		data, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		r.EndTime = &data
	}

	val = q.Get(StreamType)
	if val != "" {
		data, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		r.StreamType = &data
	}

	val = q.Get(TemplateID)
	if val != "" {
		data, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		r.TemplateId = &data
	}

	return r, nil
}
