package daemon

import (
	"fmt"
	"runtime"

	"github.com/github/git-bundle-server/internal/common"
)

type DaemonConfig struct {
	Label   string
	Program string
}

type DaemonProvider interface {
	Create(config *DaemonConfig, force bool) error

	Start(label string) error

	Stop(label string) error
}

func NewDaemonProvider(
	u common.UserProvider,
	c common.CommandExecutor,
	fs common.FileSystem,
) (DaemonProvider, error) {
	switch thisOs := runtime.GOOS; thisOs {
	case "linux":
		// Use systemd/systemctl
		return NewSystemdProvider(), nil
	case "darwin":
		// Use launchd/launchctl
		return NewLaunchdProvider(u, c, fs), nil
	default:
		return nil, fmt.Errorf("cannot configure daemon handler for OS '%s'", thisOs)
	}
}
