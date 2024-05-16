package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/leslie-wang/clusterd/common/model"
	"github.com/leslie-wang/clusterd/types"
)

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

	tx, err := h.newTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	err = h.recordDB.RemoveRecordTask(tx, id)
	if err != nil {
		return nil, err
	}
	err = h.jobDB.CompleteAndArchiveWithTx(tx, id, nil)
	if err != nil {
		return nil, err
	}
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

	if task.SourceURL == nil {
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
		SourceURL: *task.SourceURL,
		NotifyURL: task.NotifyURL,
		StorePath: task.StorePath,
		StartTime: task.StartTime,
		EndTime:   task.EndTime,
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
	}

	err = h.jobDB.Insert(tx, job)
	if err != nil {
		return nil, err
	}

	tid := strconv.FormatInt(id, 10)
	playbackURL := fmt.Sprintf("%s%s/%d/index.m3u8", h.cfg.BaseURL, types.URLPlay, id)
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
