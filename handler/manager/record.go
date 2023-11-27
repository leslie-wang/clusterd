package manager

import (
	"net/http"

	"github.com/leslie-wang/clusterd/common/util"
)

const (
	Action = "Action"

	ActionCreateLiveRecordTemplate    = "CreateLiveRecordTemplate"
	ActionDescribeLiveRecordTemplate  = "DescribeLiveRecordTemplate"
	ActionDescribeLiveRecordTemplates = "DescribeLiveRecordTemplates"
	ActionDeleteLiveRecordTemplate    = "DeleteLiveRecordTemplate"
	ActionModifyLiveRecordTemplate    = "ModifyLiveRecordTemplate"

	ActionCreateLiveRecordRule    = "CreateLiveRecordRule"
	ActionDeleteLiveRecordRule    = "DeleteLiveRecordRule"
	ActionDescribeLiveRecordRules = "DescribeLiveRecordRules"

	ActionDescribeRecordTask = "DescribeRecordTask"
	ActionCreateRecordTask   = "CreateRecordTask"
	ActionDeleteRecordTask   = "DeleteRecordTask"
	ActionStopRecordTask     = "StopRecordTask"
)

const (
	TemplateID      = "TemplateId"
	TemplateName    = "TemplateName"
	Description     = "Description"
	FlvParam        = "FlvParam"
	HlsParam        = "HlsParam"
	Mp4Param        = "Mp4Param"
	AacParam        = "AacParam"
	IsDelayLive     = "IsDelayLive"
	HlsSpecialParam = "HlsSpecialParam"
	Mp3Param        = "Mp3Param"
	RemoveWatermark = "RemoveWatermark"
	FlvSpecialParam = "FlvSpecialParam"
)

// RecordParam string
const (
	RecordInterval = "RecordInterval"
	StorageTime    = "StorageTime"
	Enable         = "Enable"
	VodSubAppId    = "VodSubAppId"
	VodFileName    = "VodFileName"
	Procedure      = "Procedure"
	StorageMode    = "StorageMode"
	ClassId        = "ClassId"
)

// misc string
const (
	HlsSpecialParamFlowContinueDuration = "HlsSpecialParam.FlowContinueDuration"
	FlvSpecialParamUploadInRecording    = "FlvSpecialParam.UploadInRecording"
)

const (
	DomainName = "DomainName"
	AppName    = "AppName"
	StreamName = "StreamName"
)

const (
	TaskID     = "TaskId"
	EndTime    = "EndTime"
	StartTime  = "StartTime"
	StreamType = "StreamType"
)

/*
func mkRequestIDByID(id int64) string {
	return strconv.FormatInt(id, 10)
}
*/

func (h *Handler) record(w http.ResponseWriter, r *http.Request) {
	var (
		resp interface{}
		err  error
	)
	q := r.URL.Query()
	switch q.Get(Action) {
	case ActionCreateLiveRecordTemplate:
		resp, err = h.handleCreateLiveRecordTemplate(q)
	case ActionDescribeLiveRecordTemplate:
		resp, err = h.handleGetLiveRecordTemplate(q)
	case ActionDescribeLiveRecordTemplates:
		resp, err = h.handleListLiveRecordTemplates()
	case ActionDeleteLiveRecordTemplate:
		resp, err = h.handleDeleteLiveRecordTemplate(q)

	case ActionCreateLiveRecordRule:
		resp, err = h.handleCreateLiveRecordRule(q)
	case ActionDeleteLiveRecordRule:
		resp, err = h.handleDeleteLiveRecordRule(q)
	case ActionDescribeLiveRecordRules:
		resp, err = h.handleListLiveRecordRules()

	case ActionDescribeRecordTask:
		resp, err = h.handleListRecordTasks()
	case ActionCreateRecordTask:
		resp, err = h.handleCreateRecordTask(q)
	case ActionDeleteRecordTask:
		resp, err = h.handleDeleteRecordTask(q)
	case ActionStopRecordTask:
	default:
		err = util.ErrNotSupportedAPI
	}
	if err != nil {
		util.WriteError(w, err)
	} else {
		util.WriteBody(w, resp)
	}
}
