package record

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/leslie-wang/clusterd/common/model"
	"github.com/leslie-wang/clusterd/types"
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
		" stream_type, start_time, end_time, source_url, store_path, create_time) " +
		" values(?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)"
	listRecordTasks = "select id, template_id, domain_name, app_name, stream_name, " +
		" start_time, end_time from record_tasks"
	removeRecordTask = "delete from record_tasks where id=?"
	getRecordTask    = "select template_id, domain_name, app_name, stream_name, start_time, end_time from record_tasks where id=?"

	insertCallbackTemplate = "insert into record_cb_templates (name, description, callback_key, begin_url, end_url," +
		" record_url, record_status_url, porn_censorship_url, stream_mix_url, push_exception_url, audio_audit_url," +
		" snapshot_url, create_time) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)"
	listCallbackTemplates = "select id, name, description, callback_key, begin_url, end_url," +
		" record_url, record_status_url, porn_censorship_url, stream_mix_url, push_exception_url, audio_audit_url," +
		" snapshot_url from record_cb_templates"
	getCallbackTemplate = "select name, description, callback_key, begin_url, end_url," +
		" record_url, record_status_url, porn_censorship_url, stream_mix_url, push_exception_url, audio_audit_url," +
		" snapshot_url from record_cb_templates where id=?"
	removeCallbackTemplate = "delete from record_cb_templates where id=?"

	insertCallbackRule = "insert into record_cb_rules (template_id, domain_name, app_name, create_time)" +
		" values(?, ?, ?, CURRENT_TIMESTAMP)"
	listCallbackRules                   = "select template_id, domain_name, app_name, create_time from record_cb_rules"
	removeCallbackRuleByDomainAppStream = "delete from record_cb_rules where domain_name=? and app_name=?"
	getCallbackRuleByDomainAndApp       = "select name, description, callback_key, begin_url, end_url, record_url," +
		" record_status_url, porn_censorship_url, stream_mix_url, push_exception_url, audio_audit_url, snapshot_url" +
		" from record_cb_templates inner join record_cb_rules as r where r.domain_name=? and r.app_name=?"
	getCallbackRuleByRecordTaskID = "select cb.id, name, description, callback_key, begin_url, end_url, record_url," +
		" record_status_url, porn_censorship_url, stream_mix_url, push_exception_url, audio_audit_url, snapshot_url" +
		" from record_cb_templates as cb inner join record_cb_rules as r inner join record_tasks as rt" +
		" on r.domain_name=rt.domain_name and r.app_name=rt.app_name" +
		" where rt.id=?"
)

var (
	prepareRecordSQLs = []string{
		insertRecordRule,
		insertRecordTask,
		insertRecordTemplate,
		insertCallbackRule,
		insertCallbackTemplate,
		listRecordRules,
		listRecordTasks,
		listRecordTemplates,
		listCallbackRules,
		listCallbackTemplates,
		removeRecordRuleByDomainAppStream,
		removeRecordTask,
		removeRecordTemplate,
		removeCallbackRuleByDomainAppStream,
		removeCallbackTemplate,
		getRecordTask,
		getRecordTemplate,
		getCallbackTemplate,
		getCallbackRuleByDomainAndApp,
		getCallbackRuleByRecordTaskID,
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
		return nil, err
	}
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

func (r *DB) InsertRecordTask(tx *sql.Tx, t *types.LiveRecordTask) (int64, error) {
	res, err := tx.Exec(insertRecordTask, t.TemplateId, t.DomainName, t.AppName, t.StreamName, t.StreamType, t.StartTime, t.EndTime,
		t.RecordStreams[0].SourceURL, t.StorePath)
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

func (r *DB) RemoveRecordTask(tx *sql.Tx, id int64) error {
	_, err := tx.Exec(removeRecordTask, id)
	return err
}

func (r *DB) InsertCallbackTemplate(t *model.CreateLiveCallbackTemplateRequestParams) (int64, error) {
	s := prepareRecordStatements[insertCallbackTemplate]
	res, err := s.Exec(t.TemplateName, t.Description, t.CallbackKey, t.StreamBeginNotifyUrl, t.StreamEndNotifyUrl,
		t.RecordNotifyUrl, t.RecordStatusNotifyUrl, t.PornCensorshipNotifyUrl, t.StreamMixNotifyUrl,
		t.PushExceptionNotifyUrl, t.AudioAuditNotifyUrl, t.SnapshotNotifyUrl)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *DB) GetCallbackTemplateByID(id int64) (*model.CallBackTemplateInfo, error) {
	s := prepareRecordStatements[getCallbackTemplate]

	var (
		t model.CallBackTemplateInfo
	)
	err := s.QueryRow(id).Scan(&t.TemplateId, &t.TemplateName, &t.Description, &t.CallbackKey,
		&t.StreamBeginNotifyUrl, &t.StreamEndNotifyUrl, &t.RecordNotifyUrl, &t.RecordStatusNotifyUrl,
		&t.PornCensorshipNotifyUrl, &t.StreamMixNotifyUrl, &t.PushExceptionNotifyUrl, &t.AudioAuditNotifyUrl,
		&t.SnapshotNotifyUrl)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *DB) ListCallbackTemplates(ctx context.Context) ([]*model.CallBackTemplateInfo, error) {
	s := prepareRecordStatements[listCallbackTemplates]

	rows, err := s.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var tmpls []*model.CallBackTemplateInfo
	for rows.Next() {
		t := &model.CallBackTemplateInfo{}
		err = rows.Scan(&t.TemplateId, &t.TemplateName, &t.Description, &t.CallbackKey,
			&t.StreamBeginNotifyUrl, &t.StreamEndNotifyUrl, &t.RecordNotifyUrl, &t.RecordStatusNotifyUrl,
			&t.PornCensorshipNotifyUrl, &t.StreamMixNotifyUrl, &t.PushExceptionNotifyUrl, &t.AudioAuditNotifyUrl,
			&t.SnapshotNotifyUrl)
		if err != nil {
			return nil, err
		}

		tmpls = append(tmpls, t)
	}
	return tmpls, nil
}

func (r *DB) RemoveCallbackTemplate(id int64) error {
	s := prepareRecordStatements[removeCallbackTemplate]
	_, err := s.Exec(id)
	return err
}

func (r *DB) InsertCallbackRule(ru *model.CreateLiveCallbackRuleRequestParams) (int64, error) {
	s := prepareRecordStatements[insertCallbackRule]
	res, err := s.Exec(ru.TemplateId, ru.DomainName, ru.AppName)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *DB) ListCallbackRules(ctx context.Context) ([]*model.CallBackRuleInfo, error) {
	s := prepareRecordStatements[listCallbackRules]

	rows, err := s.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var rules []*model.CallBackRuleInfo
	for rows.Next() {
		ru := &model.CallBackRuleInfo{}
		err = rows.Scan(&ru.TemplateId, &ru.DomainName, &ru.AppName, &ru.CreateTime)
		if err != nil {
			return nil, err
		}

		rules = append(rules, ru)
	}

	return rules, nil
}

func (r *DB) RemoveCallbackRuleByDomainApp(domain, app string) error {
	s := prepareRecordStatements[removeCallbackRuleByDomainAppStream]
	_, err := s.Exec(domain, app)
	return err
}

func (r *DB) GetCallbackRuleByRecordTaskID(id int64) (*model.CallBackTemplateInfo, error) {
	s := prepareRecordStatements[getCallbackRuleByRecordTaskID]
	var t model.CallBackTemplateInfo
	err := s.QueryRow(id).Scan(&t.TemplateId, &t.TemplateName, &t.Description, &t.CallbackKey,
		&t.StreamBeginNotifyUrl, &t.StreamEndNotifyUrl, &t.RecordNotifyUrl, &t.RecordStatusNotifyUrl,
		&t.PornCensorshipNotifyUrl, &t.StreamMixNotifyUrl, &t.PushExceptionNotifyUrl, &t.AudioAuditNotifyUrl,
		&t.SnapshotNotifyUrl)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}
