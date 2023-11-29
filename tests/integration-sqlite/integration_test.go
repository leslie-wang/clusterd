package integration_sqlite

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	testOriginServer *httptest.Server
	//simPlayer        *net.UDPConn
	//testProxyServer  *httptest.Server
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// setup one origin server because all tests need the same one.
	suite.testOriginServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("---- OS: serving %s\n", r.URL)
		f, err := os.Open("../../samples" + r.URL.Path)
		suite.Require().NoError(err)
		defer f.Close()

		_, err = io.Copy(w, f)
		suite.Require().NoError(err)
	}))
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	// close origin server at last
	suite.testOriginServer.Close()
}

/*
func run(prog string, args ...string) error {
	content, err := exec.Command(prog, args...).CombinedOutput()
	if err != nil {
		fmt.Println(string(content))
	}
	return err
}
*/

func (suite *IntegrationTestSuite) SetupTest() {
	// install manager and runner

	/*
		// setup udp receiver
		address := simPlayerIP + ":" + strconv.Itoa(simPlayerPort)

		// Resolve the UDP address
		udpAddr, err := net.ResolveUDPAddr("udp", address)
		suite.Require().NoError(err)

		// Create a UDP connection
		suite.simPlayer, err = net.ListenUDP("udp", udpAddr)
		suite.Require().NoError(err)

		// start proxy server
		suite.testProxyServer = httptest.NewServer(suite.handler.CreateRouter())
		suite.testProxyClient = hls2udp.NewClient(suite.testProxyServer.URL)

		fmt.Printf("test origin server: %s\n", suite.testOriginServer.URL)
		fmt.Printf("test proxy server: %s\n", suite.testProxyServer.URL)
	*/
}

func (suite *IntegrationTestSuite) TearDownTest() {
	/*
		suite.simPlayer.Close()
		suite.testProxyServer.Close()
	*/
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *IntegrationTestSuite) TestRecord() {
}

func TestIntegrationBasic(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
