package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

func (c *Client) RegisterRunner(name string) (*types.Job, error) {
	url := c.makeURL(types.URLRunner, name)
	fmt.Println(url)
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
