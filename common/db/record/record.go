package record

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/leslie-wang/clusterd/common/model"
	"github.com/leslie-wang/clusterd/types"
	"strconv"
	"time"
)

const (
	insertRecordTemplate = "insert into record_templates (name, params, create_time) " +
		" values(?, ?, CURRENT_TIMESTAMP)"
	listRecordTemplates  = "select id, params, create_time from record_templates"
	getRecordTemplate    = "select id, params, create_time from record_templates where id=?"
	removeRecordTemplate = "delete from record_templates where id=?"

	insertRecordRule = "insert into record_rules (template_id, domain_name, app_name, stream_name, create_time)" +
		" values(?, ?, ?, ?, CURRENT_TIMESTAMP)"
	listRecordRules                   = "select template_id, domain_name, app_name, stream_name, create_time from record_rules"
	removeRecordRuleByDomainAppStream = "delete from record_rules where domain_name=? and app_name=? and stream_name=?"

	insertRecordTask = "insert into record_tasks (template_id, domain_name, app_name, stream_name, " +
		" stream_type, start_time, end_time, create_time) values(?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)"
	listRecordTasks = "select id, template_id, domain_name, app_name, stream_name, " +
		" start_time, end_time from record_tasks"
	removeRecordTask = "delete from record_tasks where id=?"
)

var (
	prepareRecordSQLs = []string{
		insertRecordRule,
		insertRecordTask,
		insertRecordTemplate,
		listRecordRules,
		listRecordTasks,
		listRecordTemplates,
		removeRecordRuleByDomainAppStream,
		removeRecordTask,
		removeRecordTemplate,
		getRecordTemplate,
	}
	prepareRecordStatements map[string]*sql.Stmt
)

// DB is interface to record database
type DB struct {
	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	rdb := &DB{db: db}
	return rdb
}

// Prepare prepares all statement
func (r *DB) Prepare() error {
	prepareRecordStatements = make(map[string]*sql.Stmt)
	for _, s := range prepareRecordSQLs {
		stmt, err := r.db.Prepare(s)
		if err != nil {
			return err
		}
		prepareRecordStatements[s] = stmt
	}
	return nil
}

func (r *DB) InsertRecordTemplate(t *model.CreateLiveRecordTemplateRequestParams) (int64, error) {
	s := prepareRecordStatements[insertRecordTemplate]
	content, err := json.Marshal(t)
	if err != nil {
		return 0, err
	}
	res, err := s.Exec(t.TemplateName, string(content))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *DB) GetRecordTemplateByID(id int64) (*types.LiveRecordTemplate, error) {
	s := prepareRecordStatements[getRecordTemplate]

	var (
		params string
		t      types.LiveRecordTemplate
		tmpl   model.CreateLiveRecordTemplateRequestParams
	)
	err := s.QueryRow(id).Scan(&t.ID, &params, &t.CreateTime)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(params)
	err = json.Unmarshal([]byte(params), &tmpl)
	if err != nil {
		return nil, err
	}

	t.CreateLiveRecordTemplateRequestParams = &tmpl
	return &t, nil
}

func (r *DB) ListRecordTemplates(ctx context.Context) ([]types.LiveRecordTemplate, error) {
	s := prepareRecordStatements[listRecordTemplates]

	rows, err := s.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var tmpls []types.LiveRecordTemplate
	for rows.Next() {
		var (
			params string
			t      types.LiveRecordTemplate
			tmpl   model.CreateLiveRecordTemplateRequestParams
		)
		err = rows.Scan(&t.ID, &params, &t.CreateTime)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(params), &tmpl)
		if err != nil {
			return nil, err
		}

		t.CreateLiveRecordTemplateRequestParams = &tmpl
		tmpls = append(tmpls, t)
	}

	return tmpls, nil
}

func (r *DB) RemoveRecordTemplate(id int64) error {
	s := prepareRecordStatements[removeRecordTemplate]
	_, err := s.Exec(id)
	return err
}

func (r *DB) InsertRecordRule(ru *model.CreateLiveRecordRuleRequestParams) (int64, error) {
	s := prepareRecordStatements[insertRecordRule]
	res, err := s.Exec(ru.TemplateId, ru.DomainName, ru.AppName, ru.StreamName)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *DB) ListRecordRules(ctx context.Context) ([]*model.RuleInfo, error) {
	s := prepareRecordStatements[listRecordRules]

	rows, err := s.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var rules []*model.RuleInfo
	for rows.Next() {
		ru := &model.RuleInfo{}
		err = rows.Scan(&ru.TemplateId, &ru.DomainName, &ru.AppName, &ru.StreamName, &ru.CreateTime)
		if err != nil {
			return nil, err
		}

		rules = append(rules, ru)
	}

	return rules, nil
}

func (r *DB) RemoveRecordRuleByDomainAppStream(domain, app, stream string) error {
	s := prepareRecordStatements[removeRecordRuleByDomainAppStream]
	_, err := s.Exec(domain, app, stream)
	return err
}

func (r *DB) InsertRecordTask(t *model.CreateRecordTaskRequestParams) (int64, error) {
	s := prepareRecordStatements[insertRecordTask]
	res, err := s.Exec(t.TemplateId, t.DomainName, t.AppName, t.StreamName, t.StreamType, t.StartTime, t.EndTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *DB) ListRecordTasks(ctx context.Context) ([]*model.RecordTask, error) {
	s := prepareRecordStatements[listRecordTasks]

	rows, err := s.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var tasks []*model.RecordTask
	for rows.Next() {
		t := &model.RecordTask{}
		var (
			id                 int64
			startTime, endTime *time.Time
		)
		err = rows.Scan(&id, &t.TemplateId, &t.DomainName, &t.AppName, &t.StreamName,
			&startTime, &endTime)

		if startTime != nil {
			st := uint64(startTime.Unix())
			t.StartTime = &st
		}

		if endTime != nil {
			et := uint64(endTime.Unix())
			t.EndTime = &et
		}
		idStr := strconv.FormatInt(id, 10)
		t.TaskId = &idStr

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *DB) RemoveRecordTask(id int64) error {
	s := prepareRecordStatements[removeRecordTask]
	_, err := s.Exec(id)
	return err
}
