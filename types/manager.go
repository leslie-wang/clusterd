package types

import "time"

const (
	ClusterDBName = "clusterd"
	ManagerPort   = 8088
)

const (
	ID = "id"
)

const (
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
