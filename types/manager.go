package types

import (
	"github.com/leslie-wang/clusterd/common/model"
	"time"
)

const (
	ClusterDBName = "clusterd"
	ManagerPort   = 8088
)

const (
	ID = "id"
)

const (
	BaseURL     = "/mediaproc/v1"
	URLRecord   = BaseURL + "/record"
	URLRunner   = "/cd/v1/runner"
	URLRunnerID = URLRunner + "/{" + ID + "}"
	URLJob      = "/cd/v1/job"
	URLJobID    = URLJob + "/{" + ID + "}"
)

type Job struct {
	ID           int        `json:"id"`
	RefID        string     `json:"ref_id"`
	Commands     []string   `json:"commands"`
	RunningHost  *string    `json:"run_on,omitempty"`
	CreateTime   time.Time  `json:"create_time"`
	StartTime    *time.Time `json:"Start_time,omitempty"`
	LastSeenTime *time.Time `json:"last_seen_time,omitempty"`
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
