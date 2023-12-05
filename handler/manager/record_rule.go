package manager

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/leslie-wang/clusterd/common/model"
)

func (h *Handler) handleListLiveRecordRules() (*model.DescribeLiveRecordRulesResponse, error) {
	list, err := h.recordDB.ListRecordRules(context.Background())
	if err != nil {
		return nil, err
	}

	return &model.DescribeLiveRecordRulesResponse{
		Response: &model.DescribeLiveRecordRulesResponseParams{
			Rules: list,
		},
	}, nil
}

func (h *Handler) handleDeleteLiveRecordRule(q url.Values) (*model.DeleteLiveRecordRuleResponse, error) {
	domainName := q.Get(DomainName)
	if domainName == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	appName := q.Get(AppName)
	if appName == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	streamName := q.Get(StreamName)
	if streamName == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}
	err := h.recordDB.RemoveRecordRuleByDomainAppStream(domainName, appName, streamName)
	if err != nil {
		return nil, err
	}
	return &model.DeleteLiveRecordRuleResponse{Response: &model.DeleteLiveRecordRuleResponseParams{}}, nil
}

func (h *Handler) handleCreateLiveRecordRule(q url.Values) (*model.CreateLiveRecordRuleResponse, error) {
	r, err := h.parseLiveRecordRule(q)
	if err != nil {
		return nil, err
	}

	_, err = h.recordDB.InsertRecordRule(r)
	if err != nil {
		return nil, err
	}

	return &model.CreateLiveRecordRuleResponse{Response: &model.CreateLiveRecordRuleResponseParams{}}, nil
}

func (h *Handler) parseLiveRecordRule(q url.Values) (*model.CreateLiveRecordRuleRequestParams, error) {
	r := &model.CreateLiveRecordRuleRequestParams{}
	val := q.Get(TemplateID)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		r.TemplateId = &data
	}

	domainName := q.Get(DomainName)
	r.DomainName = &domainName

	appName := q.Get(AppName)
	r.AppName = &appName

	streamName := q.Get(StreamName)
	r.StreamName = &streamName
	return r, nil
}
