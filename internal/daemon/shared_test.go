package daemon_test

import (
	"github.com/git-ecosystem/git-bundle-server/internal/daemon"
)

/*********************************************/
/***************** Constants *****************/
/*********************************************/

var basicDaemonConfig = daemon.DaemonConfig{
	Label:       "com.example.testdaemon",
	Description: "Test service",
	Program:     "/usr/local/bin/test/git-bundle-web-server",
}
