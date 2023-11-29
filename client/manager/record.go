package manager

import (
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

func (c *Client) CreateRecordTask(url string, start, end *int64) (*string, error) {
	createRecordURL := c.makeURL(types.URLRecord)
	query := map[string]string{
		manager.Action:     manager.ActionCreateRecordTask,
		manager.DomainName: url,
	}
	if start != nil || end != nil {
		if start == nil || end == nil {
			return nil, errors.New("start time and end time need be set or unset at the same time")
		}
		query[manager.StartTime] = strconv.FormatInt(*start, 10)
		query[manager.EndTime] = strconv.FormatInt(*end, 10)
	}

	createRecordURL = c.addQuery(createRecordURL, query)
	fmt.Println(createRecordURL)

	req, err := http.NewRequest(http.MethodPost, createRecordURL, nil)
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
