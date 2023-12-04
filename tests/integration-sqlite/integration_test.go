package integration_sqlite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/leslie-wang/clusterd/common/db"
	"github.com/leslie-wang/clusterd/types"
	"github.com/stretchr/testify/suite"
)

var (
	dbscheduleDir = filepath.Join("..", "..", "migrations", "sqlite")
	mediaDir      = filepath.Join("..", "test-media", "apple", "basic-stream-osx-ios4-3")
	mediaFile     = "gear1_mix.mp4"
)

type IntegrationTestSuite struct {
	sqliteDBFile string
	globalCtx    context.Context
	globalCancel context.CancelFunc
	suite.Suite
	testOriginServer *httptest.Server
	//simPlayer        *net.UDPConn
	//testProxyServer  *httptest.Server
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.globalCtx, suite.globalCancel = context.WithCancel(context.Background())

	// start manager and runner
	// create sqlite database file
	f, err := os.CreateTemp("", types.ClusterDBName)
	suite.Require().NoError(err)
	suite.sqliteDBFile = f.Name()
	err = f.Close()
	suite.Require().NoError(err)

	fmt.Printf("---- testing sqlite db: %s\n", suite.sqliteDBFile)

	cmd := exec.Command("sqlite3", suite.sqliteDBFile)
	buf := &bytes.Buffer{}
	for _, n := range []string{"0.sql"} {
		schema, err := os.ReadFile(filepath.Join(dbscheduleDir, n))
		suite.Require().NoError(err)
		_, err = buf.Write(schema)
		suite.Require().NoError(err)
	}
	cmd.Stdin = buf

	content, err := cmd.CombinedOutput()
	suite.Require().NoError(err, string(content))

	go func() {
		cmd := exec.CommandContext(suite.globalCtx, "cd-manager", "--db-host", db.Sqlite+"://"+suite.sqliteDBFile,
			"--schedule-interval", "2s")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}()

	go func() {
		cmd := exec.CommandContext(suite.globalCtx, "cd-runner", "--mgr-host", "localhost")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}()

	content, err = exec.CommandContext(suite.globalCtx, "ls", "-lh", mediaDir).CombinedOutput()
	suite.Require().NoError(err, string(content))
	fmt.Printf("---- origin server has below contents:" + string(content))

	// setup one origin server because all tests need the same one.
	suite.testOriginServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(mediaDir + r.URL.Path)
		if err != nil {
			fmt.Printf("---- OS: serving [%s] from %s: %s\n", r.URL.Path, mediaDir, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()

		fmt.Printf("---- OS: serving [%s] from %s\n", r.URL.Path, mediaDir)
		_, err = f.Seek(0, 0)
		suite.Require().NoError(err)

		// add pacing to prevent ffmpeg quit too fast
		size, err := io.Copy(w, f)
		if err != nil {
			fmt.Printf("---- OS: served error: %s\n", err)
		} else {
			fmt.Printf("---- OS: served %d byte\n", size)
		}
	}))
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	// close origin server at last
	suite.testOriginServer.Close()

	suite.killProcess("cd-manager")
	suite.killProcess("cd-runner")

	suite.globalCancel()
}

func (suite *IntegrationTestSuite) SetupTest() {
}

func (suite *IntegrationTestSuite) TearDownTest() {
}

// All methods that begin with "Test" are run as tests within a suite.
func (suite *IntegrationTestSuite) TestRecordNow() {
	// make sure manager is started, and no jobs yet
	jobs := suite.listJobs()
	suite.Require().Equal(0, len(jobs))

	u := suite.testOriginServer.URL + "/" + mediaFile

	// assume the job will finish in 10 second
	id := suite.createRecord(u, "", "10s")

	jobs = suite.listJobs()
	suite.Require().Equal(1, len(jobs))
	suite.Require().Equal(id, jobs[0].ID)
	suite.Require().Equal(types.CategoryRecord, jobs[0].Category)

	// wait 10 seconds, and runner should have finished the job
	time.Sleep(10 * time.Second)

	jobs = suite.listJobs()
	suite.Require().Equal(0, len(jobs))

	job := suite.getJob(id)
	suite.Require().Equal(types.CategoryRecord, job.Category)
	suite.Require().NotNil(job.RunningHost)
	suite.Require().NotNil(job.StartTime)
	suite.Require().NotNil(job.ExitCode)

	suite.Require().Equal(0, *job.ExitCode)

	name, err := os.Hostname()
	suite.Require().NoError(err)
	suite.Require().Equal(name, *job.RunningHost)
}

func (suite *IntegrationTestSuite) TestRecordNonExistAsset() {
	jobs := suite.listJobs()
	suite.Require().Equal(0, len(jobs))

	u := suite.testOriginServer.URL + "/not-exist"

	// assume the job will finish in 10 second
	id := suite.createRecord(u, "", "10s")

	jobs = suite.listJobs()
	suite.Require().Equal(1, len(jobs))
	suite.Require().Equal(id, jobs[0].ID)
	suite.Require().Equal(types.CategoryRecord, jobs[0].Category)

	// wait 10 seconds, and runner should have finished the job
	time.Sleep(10 * time.Second)

	jobs = suite.listJobs()
	suite.Require().Equal(0, len(jobs))

	job := suite.getJob(id)
	suite.Require().Equal(types.CategoryRecord, job.Category)
	suite.Require().NotNil(job.RunningHost)
	suite.Require().NotNil(job.StartTime)
	suite.Require().NotNil(job.ExitCode)

	suite.Require().Equal(1, *job.ExitCode)

	name, err := os.Hostname()
	suite.Require().NoError(err)
	suite.Require().Equal(name, *job.RunningHost)
}

func (suite *IntegrationTestSuite) TestRecordDelete() {
	jobs := suite.listJobs()
	suite.Require().Equal(0, len(jobs))

	u := suite.testOriginServer.URL + "/not-exist"

	// assume the job will finish in 10 second
	startTime := time.Now().Add(time.Hour)
	id := suite.createRecord(u, strconv.Itoa(int(startTime.Unix())), "10s")

	jobs = suite.listJobs()
	suite.Require().Equal(1, len(jobs))
	suite.Require().Equal(id, jobs[0].ID)
	suite.Require().Equal(types.CategoryRecord, jobs[0].Category)

	// wait 10 seconds, and runner should still not pickup
	time.Sleep(10 * time.Second)
	jobs = suite.listJobs()
	suite.Require().Equal(1, len(jobs))

	suite.cancelJob(id)

	jobs = suite.listJobs()
	suite.Require().Equal(0, len(jobs))
}

func TestIntegrationBasic(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
