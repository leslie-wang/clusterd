package integration_sqlite

import "os/exec"

func (suite *IntegrationTestSuite) killProcess(name string) {
	exec.Command("pkill", "-9", name).CombinedOutput()
}
