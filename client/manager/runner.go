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

func (c *Client) AcquireJob(name string) (*types.Job, error) {
	url := c.makeURL(types.URLJobRunner, name)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, nil
	}

	job := &types.Job{}
	return job, json.Unmarshal(content, job)
}

func (c *Client) ReportJob(id, exitCode int) error {
	url := c.makeURL(types.URLJob, strconv.Itoa(id))

	report := types.JobResult{
		ID:       id,
		ExitCode: exitCode,
	}
	content, err := json.Marshal(report)
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

	return nil
}

func (c *Client) ListRunners() (map[string]types.Job, error) {
	url := c.makeURL(types.URLRunner)
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

	runners := map[string]types.Job{}
	return runners, json.NewDecoder(resp.Body).Decode(&runners)
}

func (c *Client) DownloadLogFromRunner(jobID int, writer io.Writer) error {
	url := c.makeURL(types.MkIDURLByBase(types.URLRunnerLogJob), strconv.Itoa(jobID))
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
	_, err = io.Copy(writer, resp.Body)
	return err
}
