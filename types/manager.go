package types

import (
	"strings"
	"time"

	"github.com/leslie-wang/clusterd/common/model"
)

const (
	ClusterDBName = "clusterd"
	ManagerPort   = 8088
	RunnerPort    = 8089
)

const (
	ID = "id"
)

const (
	BaseURL         = "/mediaproc/v1"
	URLRecord       = BaseURL + "/record"
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
	SourceURL string
	StartTime *uint64
	EndTime   *uint64
}

type JobResult struct {
	ID       int    `json:"id"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type LiveRecordRule struct {
	*model.CreateLiveRecordRuleRequestParams
	ID         int64     `json:"ID"`
	CreateTime time.Time `json:"CreateTime"`
}

type LiveRecordTemplate struct {
	*model.CreateLiveRecordTemplateRequestParams
	ID         int64     `json:"ID"`
	CreateTime time.Time `json:"CreateTime"`
}

type LiveRecordTask struct {
	*model.CreateRecordTaskRequestParams
	ID         int64     `json:"ID"`
	CreateTime time.Time `json:"CreateTime"`
}

// Config is configuration of DB
type Config struct {
	Driver string // mysql vs sqlite
	DBUser string // Username
	DBPass string // Password (requires User)
	Addr   string // Network address (requires Net)
	DBName string // Database name
}
