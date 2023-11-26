package job

import (
	"context"
	"database/sql"
	"github.com/leslie-wang/clusterd/types"
	"github.com/pkg/errors"
	"time"
)

const (
	insertJob  = "insert into jobs (ref_id, category, metadata, create_time, schedule_time) values(?, ?, ?, CURRENT_TIMESTAMP, ?)"
	archiveJob = `insert into job_archives (id, ref_id, category, metadata, runner, exit_code, create_time, start_time, end_time) 
					select id, ref_id, category, metadata, runner, ?, create_time, start_time, CURRENT_TIMESTAMP from jobs where id=?`
	listJobs            = "select id, ref_id, category, metadata, runner, create_time, schedule_time, start_time, last_seen_time from jobs"
	getNotStartedJob    = "select id, category, metadata, schedule_time from jobs where start_time is null and schedule_time < ? order by create_time limit 1"
	getNotFinishJobByID = "select id, ref_id, category, metadata, runner, create_time, start_time, last_seen_time from jobs where id=?"
	updateJobForRunner  = "update jobs set runner=?, start_time=CURRENT_TIMESTAMP, last_seen_time=CURRENT_TIMESTAMP where id=?"
	removeJob           = "delete from jobs where id=?"

	listActiveRunners = "select id, ref_id, category, metadata, runner, create_time, start_time, last_seen_time from jobs where runner is not null order by runner"
)

var (
	prepareJobSQLs = []string{
		insertJob,
		listJobs,
		getNotStartedJob,
		getNotFinishJobByID,
		updateJobForRunner,
		listActiveRunners,
		archiveJob,
		removeJob,
	}
	prepareJobStatements map[string]*sql.Stmt
)

// DB is interface to job database
type DB struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	rdb := &DB{db: db}
	return rdb
}

func (j *DB) Prepare() error {
	prepareJobStatements = make(map[string]*sql.Stmt)
	for _, s := range prepareJobSQLs {
		stmt, err := j.db.Prepare(s)
		if err != nil {
			return err
		}
		prepareJobStatements[s] = stmt
	}
	return nil
}

func (j *DB) Insert(tx *sql.Tx, job *types.Job) error {
	res, err := tx.Exec(insertJob, job.RefID, job.Category, job.Metadata, job.ScheduleTime)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	job.ID = int(id)
	return nil
}

func (j *DB) List() ([]types.Job, error) {
	s := prepareJobStatements[listJobs]

	rows, err := s.QueryContext(context.Background())
	if err != nil {
		return nil, err
	}

	jobs := []types.Job{}
	for rows.Next() {
		job := types.Job{}
		err = rows.Scan(&job.ID, &job.RefID, &job.Category, &job.Metadata, &job.RunningHost,
			&job.CreateTime, &job.ScheduleTime, &job.StartTime, &job.LastSeenTime)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (j *DB) GetUnStartedJob() (*types.Job, error) {
	job := &types.Job{}
	s := prepareJobStatements[getNotStartedJob]

	err := s.QueryRowContext(context.Background()).Scan(&job.ID, &job.Category, &job.Metadata)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	return job, err
}

func (j *DB) ListActiveRunners() (map[string]types.Job, error) {
	s := prepareJobStatements[listActiveRunners]

	rows, err := s.QueryContext(context.Background())
	if err != nil {
		return nil, err
	}

	runners := map[string]types.Job{}
	for rows.Next() {
		var job types.Job
		err = rows.Scan(&job.ID, &job.RefID, &job.Category, &job.Metadata, &job.RunningHost, &job.CreateTime,
			&job.StartTime, &job.LastSeenTime)
		if err != nil {
			return nil, err
		}

		if job.RunningHost == nil {
			return nil, errors.Errorf("Job %d has not assigned host while listing active runners", job.ID)
		}
		runners[*job.RunningHost] = job
	}

	return runners, nil
}

func (j *DB) Acquire(runner string, scheduleTime time.Time) (*types.Job, error) {
	tx, err := j.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Rollback the transaction if an error occurs

	getStmt, err := tx.Prepare(getNotStartedJob)
	if err != nil {
		return nil, err
	}
	defer getStmt.Close()

	job := &types.Job{}

	err = getStmt.QueryRowContext(context.Background(), scheduleTime).Scan(&job.ID, &job.Category, &job.Metadata, &job.ScheduleTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	updateStmt, err := tx.Prepare(updateJobForRunner)
	if err != nil {
		return nil, err
	}
	defer updateStmt.Close()

	_, err = updateStmt.Exec(runner, job.ID)
	if err != nil {
		return nil, err
	}

	return job, tx.Commit()
}

func (j *DB) CompleteAndArchive(id, exitCode int) error {
	tx, err := j.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback the transaction if an error occurs

	archStmt, err := tx.Prepare(archiveJob)
	if err != nil {
		return err
	}
	defer archStmt.Close()

	_, err = archStmt.Exec(exitCode, id)
	if err != nil {
		return err
	}

	removeStmt, err := tx.Prepare(removeJob)
	if err != nil {
		return err
	}
	defer removeStmt.Close()

	_, err = removeStmt.Exec(id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
