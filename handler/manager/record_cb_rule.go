package manager

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/leslie-wang/clusterd/common/model"
)

func (h *Handler) handleDescribeLiveCallbackRules() (*model.DescribeLiveCallbackRulesResponse, error) {
	list, err := h.recordDB.ListCallbackRules(context.Background())
	if err != nil {
		return nil, err
	}

	return &model.DescribeLiveCallbackRulesResponse{
		Response: &model.DescribeLiveCallbackRulesResponseParams{
			Rules: list,
		},
	}, nil
}

func (h *Handler) handleDeleteLiveCallbackRule(q url.Values) (*model.DeleteLiveCallbackRuleResponse, error) {
	domainName := q.Get(DomainName)
	if domainName == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	appName := q.Get(AppName)
	if appName == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	err := h.recordDB.RemoveCallbackRuleByDomainApp(domainName, appName)
	if err != nil {
		return nil, err
	}
	return &model.DeleteLiveCallbackRuleResponse{Response: &model.DeleteLiveCallbackRuleResponseParams{}}, nil
}

func (h *Handler) handleCreateLiveCallbackRule(q url.Values) (*model.CreateLiveCallbackRuleResponse, error) {
	r, err := h.parseLiveCallbackRule(q)
	if err != nil {
		return nil, err
	}

	_, err = h.recordDB.InsertCallbackRule(r)
	if err != nil {
		return nil, err
	}

	return &model.CreateLiveCallbackRuleResponse{Response: &model.CreateLiveCallbackRuleResponseParams{}}, nil
}

func (h *Handler) parseLiveCallbackRule(q url.Values) (*model.CreateLiveCallbackRuleRequestParams, error) {
	r := &model.CreateLiveCallbackRuleRequestParams{}
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
	return r, nil
}
