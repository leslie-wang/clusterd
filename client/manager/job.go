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
	url := c.makeURL(types.URLJobLog, strconv.Itoa(jobID))
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

func (c *Client) GetJob(jobID int) (*types.Job, error) {
	url := c.makeURL(types.URLJob, strconv.Itoa(jobID))
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

	job := &types.Job{}
	return job, json.NewDecoder(resp.Body).Decode(job)
}

func (c *Client) ReportJobStatus(status *types.JobStatus) error {
	content, err := json.Marshal(status)
	if err != nil {
		return err
	}
	url := c.makeURL(types.URLJob, strconv.Itoa(status.ID))
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

	return nil
}
