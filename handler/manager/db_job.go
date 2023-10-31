package manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/leslie-wang/clusterd/types"
	"github.com/pkg/errors"
)

const (
	insertJob  = "insert into jobs (ref_id, cmd, create_time) values(?, ?, CURRENT_TIMESTAMP)"
	archiveJob = `insert into job_archives (id, ref_id, cmd, runner, create_time, start_time, exit_code, end_time) 
					select id, ref_id, cmd, runner, create_time, start_time, ?, CURRENT_TIMESTAMP from jobs where id=?`
	listJobs            = "select id, ref_id, cmd, runner, create_time, start_time, last_seen_time from jobs"
	getNotStartedJob    = "select id, cmd from jobs where start_time is null order by create_time limit 1"
	getNotFinishJobByID = "select id, ref_id, cmd, runner, create_time, start_time, last_seen_time from jobs where id=?"
	updateJobForRunner  = "update jobs set runner=?, start_time=CURRENT_TIMESTAMP, last_seen_time=CURRENT_TIMESTAMP where id=?"
	removeJob           = "delete from jobs where id=?"

	listActiveRunners = "select id, ref_id, cmd, runner, create_time, start_time, last_seen_time from jobs where runner is not null order by runner"
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

func (h *Handler) prepareJobDB() error {
	prepareJobStatements = make(map[string]*sql.Stmt)
	for _, s := range prepareJobSQLs {
		stmt, err := h.db.Prepare(s)
		if err != nil {
			return err
		}
		prepareJobStatements[s] = stmt
	}
	return nil
}

func (h *Handler) insertJobDB(j *types.Job) error {
	s := prepareJobStatements[insertJob]

	cmd, err := json.Marshal(j.Commands)
	if err != nil {
		return err
	}

	res, err := s.Exec(j.RefID, cmd)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	j.ID = int(id)
	return nil
}

func (h *Handler) listJobsDB() ([]types.Job, error) {
	s := prepareJobStatements[listJobs]

	rows, err := s.QueryContext(context.Background())
	if err != nil {
		return nil, err
	}

	jobs := []types.Job{}
	for rows.Next() {
		j := types.Job{}
		var (
			cmd string
		)
		err = rows.Scan(&j.ID, &j.RefID, &cmd, &j.RunningHost, &j.CreateTime, &j.StartTime, &j.LastSeenTime)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(cmd), &j.Commands)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (h *Handler) getUnStartedJobDB() (*types.Job, error) {
	cmd := ""
	job := &types.Job{}
	s := prepareJobStatements[listJobs]

	err := s.QueryRowContext(context.Background()).Scan(&job.ID, &cmd)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(cmd), &job.Commands)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (h *Handler) updateJobForRunnerDB(id int, runner string) error {
	s := prepareJobStatements[updateJobForRunner]
	_, err := s.Exec(runner, id)
	return err
}

func (h *Handler) listActiveRunnersDB() (map[string]types.Job, error) {
	s := prepareJobStatements[listActiveRunners]

	rows, err := s.QueryContext(context.Background())
	if err != nil {
		return nil, err
	}

	runners := map[string]types.Job{}
	for rows.Next() {
		var (
			cmd string
			j   types.Job
		)
		err = rows.Scan(&j.ID, &j.RefID, &cmd, &j.RunningHost, &j.CreateTime, &j.StartTime, &j.LastSeenTime)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(cmd), &j.Commands)
		if err != nil {
			return nil, err
		}
		if j.RunningHost == nil {
			return nil, errors.Errorf("Job %d has not assigned host while listing active runners", j.ID)
		}
		runners[*j.RunningHost] = j
	}

	return runners, nil
}

func (h *Handler) findAndUpdateJobTx(runner string) (*types.Job, error) {
	tx, err := h.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Rollback the transaction if an error occurs

	getStmt, err := tx.Prepare(getNotStartedJob)
	if err != nil {
		return nil, err
	}
	defer getStmt.Close()

	cmd := ""
	job := &types.Job{}

	err = getStmt.QueryRowContext(context.Background()).Scan(&job.ID, &cmd)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(cmd), &job.Commands)
	if err != nil {
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

func (h *Handler) archiveJobTx(id, exitCode int) error {
	tx, err := h.db.Begin()
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
