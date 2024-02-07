package manager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"

	"github.com/leslie-wang/clusterd/common/model"
)

func (h *Handler) handleDescribeLiveCallbackTemplate(q url.Values) (*model.DescribeLiveCallbackTemplateResponse, error) {
	val := q.Get(TemplateID)
	if val == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	item, err := h.recordDB.GetCallbackTemplateByID(id)
	if err != nil {
		return nil, err
	}
	return &model.DescribeLiveCallbackTemplateResponse{
		Response: &model.DescribeLiveCallbackTemplateResponseParams{
			Template: item,
		},
	}, nil
}

func (h *Handler) handleDescribeLiveCallbackTemplates() (*model.DescribeLiveCallbackTemplatesResponse, error) {
	list, err := h.recordDB.ListCallbackTemplates(context.Background())
	if err != nil {
		return nil, err
	}

	resp := &model.DescribeLiveCallbackTemplatesResponse{
		Response: &model.DescribeLiveCallbackTemplatesResponseParams{
			Templates: []*model.CallBackTemplateInfo{},
		},
	}
	resp.Response.Templates = list
	return resp, nil
}

func (h *Handler) handleDeleteLiveCallbackTemplate(q url.Values) (*model.DeleteLiveCallbackTemplateResponse, error) {
	val := q.Get(TemplateID)
	if val == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	err = h.recordDB.RemoveCallbackTemplate(id)
	if err != nil {
		return nil, err
	}
	return &model.DeleteLiveCallbackTemplateResponse{Response: &model.DeleteLiveCallbackTemplateResponseParams{}}, nil
}

func (h *Handler) handleCreateLiveCallbackTemplate(q url.Values, request io.ReadCloser) (*model.CreateLiveCallbackTemplateResponse, error) {
	defer request.Close()

	var (
		t   *model.CreateLiveCallbackTemplateRequestParams
		err error
	)
	if h.cfg.ParamQuery {
		t, err = h.parseLiveCallbackTemplate(q)
		if err != nil {
			return nil, err
		}
	} else {
		t = &model.CreateLiveCallbackTemplateRequestParams{}
		err = json.NewDecoder(request).Decode(t)
		if err != nil {
			return nil, err
		}
	}

	id, err := h.recordDB.InsertCallbackTemplate(t)
	if err != nil {
		return nil, err
	}

	return &model.CreateLiveCallbackTemplateResponse{
		Response: &model.CreateLiveCallbackTemplateResponseParams{
			TemplateId: &id,
		},
	}, nil
}

func (h *Handler) parseLiveCallbackTemplate(q url.Values) (*model.CreateLiveCallbackTemplateRequestParams, error) {
	t := &model.CreateLiveCallbackTemplateRequestParams{}
	name := q.Get(TemplateName)
	if name == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}
	t.TemplateName = &name

	desc := q.Get(Description)
	if desc != "" {
		t.Description = &desc
	}

	beginURL := q.Get(StreamBeginNotifyUrl)
	if beginURL != "" {
		t.StreamBeginNotifyUrl = &beginURL
	}

	endURL := q.Get(StreamEndNotifyUrl)
	if endURL != "" {
		t.StreamEndNotifyUrl = &endURL
	}

	recordURL := q.Get(RecordNotifyUrl)
	if recordURL != "" {
		t.RecordNotifyUrl = &recordURL
	}

	recordStatusURL := q.Get(RecordStatusNotifyUrl)
	if recordStatusURL != "" {
		t.RecordStatusNotifyUrl = &recordStatusURL
	}

	snapshotURL := q.Get(SnapshotNotifyUrl)
	if snapshotURL != "" {
		t.SnapshotNotifyUrl = &snapshotURL
	}

	pornCensorURL := q.Get(PornCensorshipNotifyUrl)
	if pornCensorURL != "" {
		t.PornCensorshipNotifyUrl = &pornCensorURL
	}

	key := q.Get(CallbackKey)
	if key != "" {
		t.CallbackKey = &key
	}

	mixURL := q.Get(StreamMixNotifyUrl)
	if mixURL != "" {
		t.StreamMixNotifyUrl = &mixURL
	}

	pushExceptionURL := q.Get(PushExceptionNotifyUrl)
	if pushExceptionURL != "" {
		t.PushExceptionNotifyUrl = &pushExceptionURL
	}

	audioAuditURL := q.Get(AudioAuditNotifyUrl)
	if audioAuditURL != "" {
		t.AudioAuditNotifyUrl = &audioAuditURL
	}

	return t, nil
}
