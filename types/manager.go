package types

import (
	"strings"
	"time"

	"github.com/leslie-wang/clusterd/common/model"
)

const (
	ClusterDBName  = "clusterd"
	ManagerPort    = 8088
	RunnerPort     = 8089
	UtilListenPort = 8090
)

const (
	ID = "id"
)

const (
	BaseURL         = "/mediaproc/v1"
	URLRecord       = BaseURL + "/record"
	URLPlay         = BaseURL + "/play"
	URLDownload     = BaseURL + "/dl"
	URLRunner       = "/cd/v1/runner"
	URLRunnerLogJob = URLRunner + "/log/job/"

	URLJob       = "/cd/v1/job"
	URLJobRunner = URLJob + "/runner/"
	URLJobLog    = URLJob + "/log/"
)

func MkIDURLByBase(base string) string {
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + "{" + ID + "}"
}

type JobCategory uint

const (
	CategoryRecord JobCategory = iota
)

type Job struct {
	ID           int         `json:"id"`
	RefID        int64       `json:"ref_id"`
	Category     JobCategory `json:"category"`
	Metadata     string      `json:"metadata"`
	RunningHost  *string     `json:"run_on,omitempty"`
	ExitCode     *int        `json:"exit_code,omitempty"`
	CreateTime   time.Time   `json:"create_time"`
	ScheduleTime *time.Time  `json:"schedule_time"`
	StartTime    *time.Time  `json:"start_time,omitempty"`
	EndTime      *time.Time  `json:"end_time,omitempty"`
	LastSeenTime *time.Time  `json:"last_seen_time,omitempty"`
}

type JobRecord struct {
	NotifyURL       string
	StorePath       string
	StartTime       *uint64
	EndTime         *uint64
	RecordStreams   []model.RecordInputStream
	Mp4FileDuration uint
	RecordTimeout   int64
}

type JobStatusType int

const (
	RecordJobStart JobStatusType = iota
	RecordJobEnd
	RecordJobException
	RecordMp4FileCreated
)

type JobStatus struct {
	ID          int           `json:"id"`
	Type        JobStatusType `json:"type"`
	ExitCode    int           `json:"exit_code"`
	Stdout      string        `json:"stdout"`
	Stderr      string        `json:"stderr"`
	Mp4Filename string        `json:"mp4_filename"`
	Size        uint64        `json:"size"`
	Duration    uint64        `json:"duration"`
}

type LiveRecordRule struct {
	*model.CreateLiveRecordRuleRequestParams
	ID         int64     `json:"id"`
	CreateTime time.Time `json:"create_time"`
}

type LiveRecordTemplate struct {
	*model.CreateLiveRecordTemplateRequestParams
	ID         int64     `json:"id"`
	CreateTime time.Time `json:"create_time"`
}

type LiveRecordTask struct {
	*model.CreateRecordTaskRequestParams
	ID         int64     `json:"id"`
	CreateTime time.Time `json:"create_time"`
}

// Config is configuration of DB
type Config struct {
	Driver string // mysql vs sqlite
	DBUser string // Username
	DBPass string // Password (requires User)
	Addr   string // Network address (requires Net)
	DBName string // Database name
}

type LiveCallbackEventType int

const (
	LiveCallbackEventTypePushStart    LiveCallbackEventType = 1
	LiveCallbackEventTypePushStop     LiveCallbackEventType = 0
	LiveCallbackEventTypeRecordFile   LiveCallbackEventType = 100
	LiveCallbackEventTypeException    LiveCallbackEventType = 321
	LiveCallbackEventTypeRecordStatus LiveCallbackEventType = 332
)

type LiveCallbackStreamEvent struct {
	EventType LiveCallbackEventType `json:"event_type"`

	Sign string `json:"sign"`
	T    int64  `json:"t"`

	AppID        int    `json:"appid"`
	App          string `json:"app"`
	AppName      string `json:"appname"`
	StreamID     string `json:"stream_id"`
	ChannelID    string `json:"channel_id"`
	EventTime    int64  `json:"event_time"`
	Sequence     string `json:"sequence"`
	Node         string `json:"node"`
	UserID       string `json:"user_ip"`
	StreamParam  string `json:"stream_param"`
	PushDuration string `json:"push_duration"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	SetID        int    `json:"set_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type LiveCallbackAbnormalDetail struct {
	Desc      string `json:"desc"`
	OccurTime string `json:"occur_time"`
}

type LiveCallbackAbnormalEvent struct {
	Type   int                          `json:"type"`
	Count  int                          `json:"count"`
	Detail []LiveCallbackAbnormalDetail `json:"detail"`
	DescCN string                       `json:"type_desc_cn"`
	DescEN string                       `json:"type_desc_en"`
}

type LiveCallbackStreamExceptionEvent struct {
	EventType LiveCallbackEventType `json:"event_type"`

	AppID          int                         `json:"appid"`
	StreamID       string                      `json:"stream_id"`
	DataTime       int                         `json:"data_time"`
	ReportInterval int                         `json:"report_interval"`
	AbnormalEvent  []LiveCallbackAbnormalEvent `json:"abnormal_event"`
}

type LiveCallbackRecordFileEvent struct {
	EventType LiveCallbackEventType `json:"event_type"`

	Sign string `json:"sign"`
	T    int64  `json:"t"`

	AppID          int    `json:"appid"`
	App            string `json:"app"`
	AppName        string `json:"appname"`
	StreamID       string `json:"stream_id"`
	ChannelID      string `json:"channel_id"`
	FileID         string `json:"file_id"`
	RecordFileID   string `json:"record_file_id"`
	FileFormat     string `json:"file_format"`
	TaskID         string `json:"task_id"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time"`
	StartTimeUsec  int    `json:"start_time_usec"`
	EndTimeUsec    int    `json:"end_time_usec"`
	Duration       int64  `json:"duration"`
	FileSize       uint64 `json:"file_size"`
	StreamParam    string `json:"stream_param"`
	VideoURL       string `json:"video_url"`
	MediaStartTime int64  `json:"media_start_time"`
	RecordBps      int    `json:"record_bps"`
	CallbackExt    string `json:"callback_ext"`
}

type LiveRecordStatusEvent string

const (
	LiveRecordStatusStartSucceeded = "record_start_succeeded"
	LiveRecordStatusStartFailed    = "record_start_failed"
	LiveRecordStatusPaused         = "record_paused"
	LiveRecordStatusResumed        = "record_resumed"
	LiveRecordStatusError          = "record_error"
	LiveRecordStatusEnded          = "record_ended"
	LiveRecordMp4FileCreated       = "record_mp4_created"
)

type LiveCallbackRecordStatusEvent struct {
	EventType LiveCallbackEventType `json:"event_type"`

	Sign string `json:"sign"`
	T    int64  `json:"t"`

	AppID        int                   `json:"appid"`
	AppName      string                `json:"appname"`
	Domain       string                `json:"domain"`
	EventTime    int64                 `json:"event_time"`
	RecordDetail string                `json:"record_detail"`
	RecordEvent  LiveRecordStatusEvent `json:"record_event"`
	DownloadURL  string                `json:"download_url"`
	Sequence     string                `json:"seq"`
	SessionID    string                `json:"session_id"`
	StreamID     string                `json:"stream_id"`
	Size         uint64                `json:"size"`
	Duration     uint64                `json:"duration"`
}
