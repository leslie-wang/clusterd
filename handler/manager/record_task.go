package manager

import (
	"context"
	"errors"
	"github.com/leslie-wang/clusterd/common/model"
	"net/url"
	"strconv"
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

	err = h.recordDB.RemoveRecordTask(id)
	if err != nil {
		return nil, err
	}
	return &model.DeleteLiveRecordRuleResponse{Response: &model.DeleteLiveRecordRuleResponseParams{}}, nil
}

func (h *Handler) handleCreateRecordTask(q url.Values) (*model.CreateRecordTaskResponse, error) {
	r, err := h.parseRecordTask(q)
	if err != nil {
		return nil, err
	}

	id, err := h.recordDB.InsertRecordTask(r)
	if err != nil {
		return nil, err
	}

	tid := strconv.FormatInt(id, 10)
	return &model.CreateRecordTaskResponse{Response: &model.CreateRecordTaskResponseParams{
		TaskId: &tid,
	}}, nil
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
