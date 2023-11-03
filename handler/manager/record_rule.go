package manager

import (
	"context"
	"errors"
	"github.com/leslie-wang/clusterd/common/model"
	"net/url"
	"strconv"
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

	appNames, ok := q[AppName]
	if !ok || len(appNames) == 0 {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	streamNames, ok := q[StreamName]
	if !ok || len(streamNames) == 0 {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}
	err := h.recordDB.RemoveRecordRuleByDomainAppStream(domainName, appNames[0], streamNames[0])
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
