package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/leslie-wang/clusterd/common/model"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/handler/manager"
	"github.com/leslie-wang/clusterd/types"
)

func (c *Client) CreateRecordTask(domain, app, stream, url string, start, end *uint64) (*string, error) {
	task := &types.LiveRecordTask{}
	task.CreateRecordTaskRequestParams = &model.CreateRecordTaskRequestParams{
		DomainName:    &domain,
		AppName:       &app,
		StreamName:    &stream,
		StartTime:     start,
		RecordStreams: []model.RecordInputStream{{SourceURL: url}},
		EndTime:       end,
	}
	createRecordURL := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action: manager.ActionCreateRecordTask,
	}
	if start != nil || end != nil {
		if start == nil || end == nil {
			return nil, errors.New("start time and end time need be set or unset at the same time")
		}
	}

	createRecordURL = c.addQuery(createRecordURL, query)

	content, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(content)
	req, err := http.NewRequest(http.MethodPost, createRecordURL, buf)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, util.MakeStatusError(resp.Body)
	}
	createRecordResp := &model.CreateRecordTaskResponse{}
	return createRecordResp.Response.TaskId, json.NewDecoder(resp.Body).Decode(createRecordResp)
}

func (c *Client) CancelRecordTask(id string) error {
	cancelRecordURL := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action: manager.ActionDeleteRecordTask,
		manager.TaskID: id,
	}

	cancelRecordURL = c.addQuery(cancelRecordURL, query)
	fmt.Println(cancelRecordURL)

	req, err := http.NewRequest(http.MethodPost, cancelRecordURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return util.MakeStatusError(resp.Body)
	}
	return nil
}

func (c *Client) ListLiveCallbackTemplates() ([]*model.CallBackTemplateInfo, error) {
	u := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action: manager.ActionDescribeLiveCallbackTemplates,
	}

	u = c.addQuery(u, query)

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, util.MakeStatusError(resp.Body)
	}

	templates := model.DescribeLiveCallbackTemplatesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&templates)
	if err != nil {
		return nil, err
	}
	return templates.Response.Templates, nil
}

func (c *Client) CreateLiveCallbackTemplate(template *model.CallBackTemplateInfo) (int64, error) {
	u := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action: manager.ActionCreateLiveCallbackTemplate,
	}

	u = c.addQuery(u, query)

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(template)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, u, buf)
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, util.MakeStatusError(resp.Body)
	}

	tresp := model.CreateLiveCallbackTemplateResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tresp)
	if err != nil {
		return 0, err
	}
	return *tresp.Response.TemplateId, nil
}

func (c *Client) ListLiveCallbackRules() ([]*model.CallBackRuleInfo, error) {
	u := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action: manager.ActionDescribeLiveCallbackRules,
	}

	u = c.addQuery(u, query)

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, util.MakeStatusError(resp.Body)
	}

	templates := model.DescribeLiveCallbackRulesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&templates)
	if err != nil {
		return nil, err
	}
	return templates.Response.Rules, nil
}

func (c *Client) CreateLiveCallbackRule(templateID int, domainName, appName string) error {
	u := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action:     manager.ActionCreateLiveCallbackRule,
		manager.TemplateID: strconv.Itoa(templateID),
		manager.DomainName: domainName,
		manager.AppName:    appName,
	}

	u = c.addQuery(u, query)
	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return util.MakeStatusError(resp.Body)
	}

	return nil
}
