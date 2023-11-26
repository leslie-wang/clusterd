package manager

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

func (c *Client) CreateJob(j *types.Job) error {
	url := c.makeURL(types.URLJob)
	content, err := json.Marshal(j)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(content))
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
	return json.NewDecoder(resp.Body).Decode(j)
}

func (c *Client) ListJobs() ([]types.Job, error) {
	url := c.makeURL(types.URLJob)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	jobs := []types.Job{}
	return jobs, json.NewDecoder(resp.Body).Decode(&jobs)
}

func (c *Client) DownloadLogFromManager(jobID int) (io.ReadCloser, error) {
	url := c.makeURL(types.MkIDURLByBase(types.URLJobLog), strconv.Itoa(jobID))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, util.MakeStatusError(resp.Body)
	}
	return resp.Body, nil
}
