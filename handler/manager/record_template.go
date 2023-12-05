package manager

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/leslie-wang/clusterd/common/model"
)

func (h *Handler) handleGetLiveRecordTemplate(q url.Values) (*model.DescribeLiveRecordTemplateResponse, error) {
	val := q.Get(TemplateID)
	if val == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	item, err := h.recordDB.GetRecordTemplateByID(id)
	if err != nil {
		return nil, err
	}
	return &model.DescribeLiveRecordTemplateResponse{
		Response: &model.DescribeLiveRecordTemplateResponseParams{
			Template: &model.RecordTemplateInfo{
				TemplateId:      &item.ID,
				TemplateName:    item.TemplateName,
				Description:     item.Description,
				FlvParam:        item.FlvParam,
				HlsParam:        item.HlsParam,
				Mp4Param:        item.Mp4Param,
				AacParam:        item.AacParam,
				IsDelayLive:     item.IsDelayLive,
				HlsSpecialParam: item.HlsSpecialParam,
				Mp3Param:        item.Mp3Param,
				RemoveWatermark: item.RemoveWatermark,
				FlvSpecialParam: item.FlvSpecialParam,
			},
		},
	}, nil
}

func (h *Handler) handleListLiveRecordTemplates() (*model.DescribeLiveRecordTemplatesResponse, error) {
	list, err := h.recordDB.ListRecordTemplates(context.Background())
	if err != nil {
		return nil, err
	}

	resp := &model.DescribeLiveRecordTemplatesResponse{
		Response: &model.DescribeLiveRecordTemplatesResponseParams{
			Templates: []*model.RecordTemplateInfo{},
		},
	}
	for _, item := range list {
		resp.Response.Templates = append(resp.Response.Templates, &model.RecordTemplateInfo{
			TemplateId:      &item.ID,
			TemplateName:    item.TemplateName,
			Description:     item.Description,
			FlvParam:        item.FlvParam,
			HlsParam:        item.HlsParam,
			Mp4Param:        item.Mp4Param,
			AacParam:        item.AacParam,
			IsDelayLive:     item.IsDelayLive,
			HlsSpecialParam: item.HlsSpecialParam,
			Mp3Param:        item.Mp3Param,
			RemoveWatermark: item.RemoveWatermark,
			FlvSpecialParam: item.FlvSpecialParam,
		})
	}
	return resp, nil
}

func (h *Handler) handleDeleteLiveRecordTemplate(q url.Values) (*model.DeleteLiveRecordTemplateResponse, error) {
	val := q.Get(TemplateID)
	if val == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	err = h.recordDB.RemoveRecordTemplate(id)
	if err != nil {
		return nil, err
	}
	return &model.DeleteLiveRecordTemplateResponse{Response: &model.DeleteLiveRecordTemplateResponseParams{}}, nil
}

func (h *Handler) handleCreateLiveRecordTemplate(q url.Values) (*model.CreateLiveRecordTemplateResponse, error) {
	t, err := h.parseLiveRecordTemplate(q)
	if err != nil {
		return nil, err
	}

	id, err := h.recordDB.InsertRecordTemplate(t)
	if err != nil {
		return nil, err
	}

	return &model.CreateLiveRecordTemplateResponse{
		Response: &model.CreateLiveRecordTemplateResponseParams{
			TemplateId: &id,
		},
	}, nil
}

func (h *Handler) parseLiveRecordTemplate(q url.Values) (*model.CreateLiveRecordTemplateRequestParams, error) {
	t := &model.CreateLiveRecordTemplateRequestParams{}
	val := q.Get(TemplateName)
	if val == "" {
		return nil, errors.New(model.INVALIDPARAMETERVALUE)
	}
	t.TemplateName = &val

	val = q.Get(Description)
	if val != "" {
		t.Description = &val
	}

	ok := hasPrefixInQueryKeys(q, FlvParam)
	if ok {
		param, err := h.parseRecordParam(q, FlvParam+".")
		if err != nil {
			return nil, err
		}
		t.FlvParam = param
	}

	ok = hasPrefixInQueryKeys(q, HlsParam)
	if ok {
		param, err := h.parseRecordParam(q, HlsParam+".")
		if err != nil {
			return nil, err
		}
		t.HlsParam = param
	}

	ok = hasPrefixInQueryKeys(q, Mp4Param)
	if ok {
		param, err := h.parseRecordParam(q, Mp4Param+".")
		if err != nil {
			return nil, err
		}
		t.Mp4Param = param
	}

	ok = hasPrefixInQueryKeys(q, AacParam)
	if ok {
		param, err := h.parseRecordParam(q, AacParam+".")
		if err != nil {
			return nil, err
		}
		t.AacParam = param
	}

	val = q.Get(IsDelayLive)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		t.IsDelayLive = &data
	}

	val = q.Get(HlsSpecialParamFlowContinueDuration)
	if val != "" {
		data, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		t.HlsSpecialParam = &model.HlsSpecialParam{FlowContinueDuration: &data}
	}

	ok = hasPrefixInQueryKeys(q, Mp3Param)
	if ok {
		param, err := h.parseRecordParam(q, Mp3Param+".")
		if err != nil {
			return nil, err
		}
		t.Mp3Param = param
	}

	val = q.Get(RemoveWatermark)
	if val != "" {
		rw, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		t.RemoveWatermark = &rw
	}

	val = q.Get(FlvSpecialParamUploadInRecording)
	if val != "" {
		rw, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		t.FlvSpecialParam = &model.FlvSpecialParam{UploadInRecording: &rw}
	}

	return t, nil
}

func (h *Handler) parseRecordParam(q url.Values, prefix string) (*model.RecordParam, error) {
	rp := &model.RecordParam{}
	val := q.Get(prefix + RecordInterval)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		rp.RecordInterval = &data
	}

	val = q.Get(prefix + StorageTime)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		rp.StorageTime = &data
	}

	val = q.Get(prefix + Enable)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		rp.Enable = &data
	}

	val = q.Get(prefix + VodSubAppId)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		rp.VodSubAppId = &data
	}

	vodFileName := q.Get(prefix + VodFileName)
	if vodFileName != "" {
		rp.VodFileName = &vodFileName
	}

	procedure := q.Get(prefix + Procedure)
	if procedure != "" {
		rp.Procedure = &procedure
	}

	smode := q.Get(prefix + StorageMode)
	if smode != "" {
		rp.StorageMode = &smode
	}

	val = q.Get(prefix + ClassId)
	if val != "" {
		data, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		rp.ClassId = &data
	}

	return rp, nil
}
