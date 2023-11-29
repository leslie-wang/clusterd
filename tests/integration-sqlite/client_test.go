package integration_sqlite

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/leslie-wang/clusterd/types"
)

func (suite *IntegrationTestSuite) listJobs() []types.Job {
	f, err := os.CreateTemp("", "")
	suite.Require().NoError(err)
	filename := f.Name()
	err = f.Close()
	suite.Require().NoError(err)

	defer os.Remove(filename)

	content, err := exec.CommandContext(suite.globalCtx, "cd-util", "--mgr-host", "localhost", "job", "queue",
		"--retry-count", "10", "--output", filename).CombinedOutput()
	suite.Require().NoError(err, string(content))

	f, err = os.Open(filename)
	suite.Require().NoError(err)
	defer f.Close()

	var jobs []types.Job
	suite.Require().NoError(json.NewDecoder(f).Decode(&jobs))

	return jobs
}

func (suite *IntegrationTestSuite) getJob(id int) *types.Job {
	f, err := os.CreateTemp("", "")
	suite.Require().NoError(err)
	filename := f.Name()
	err = f.Close()
	suite.Require().NoError(err)

	defer os.Remove(filename)

	content, err := exec.CommandContext(suite.globalCtx, "cd-util", "--mgr-host", "localhost", "job", "get",
		"--output", filename, strconv.Itoa(id)).CombinedOutput()
	suite.Require().NoError(err, string(content))

	f, err = os.Open(filename)
	suite.Require().NoError(err)
	defer f.Close()

	job := &types.Job{}
	suite.Require().NoError(json.NewDecoder(f).Decode(job))

	return job
}

func (suite *IntegrationTestSuite) createRecord(u, start, end string) int {
	f, err := os.CreateTemp("", "")
	suite.Require().NoError(err)
	filename := f.Name()
	err = f.Close()
	suite.Require().NoError(err)

	defer os.Remove(filename)
	args := []string{"--mgr-host", "localhost", "record", "create", "--output", filename, u}
	if start == "" {
		args = append(args, end)
	} else {
		args = append(args, start, end)
	}

	content, err := exec.CommandContext(suite.globalCtx, "cd-util", args...).CombinedOutput()
	suite.Require().NoError(err, string(content))

	f, err = os.Open(filename)
	suite.Require().NoError(err)
	defer f.Close()

	content, err = io.ReadAll(f)
	suite.Require().NoError(err)

	id, err := strconv.Atoi(string(content))
	suite.Require().NoError(err)
	return id
}
