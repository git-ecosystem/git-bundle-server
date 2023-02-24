package daemon

import (
	"context"
	"fmt"
	"runtime"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/log"
)

type DaemonConfig struct {
	Label       string
	Description string
	Program     string
	Arguments   []string
}

type DaemonProvider interface {
	Create(ctx context.Context, config *DaemonConfig, force bool) error

	Start(ctx context.Context, label string) error

	Stop(ctx context.Context, label string) error

	Remove(ctx context.Context, label string) error
}

func NewDaemonProvider(
	l log.TraceLogger,
	u common.UserProvider,
	c cmd.CommandExecutor,
	fs common.FileSystem,
) (DaemonProvider, error) {
	switch thisOs := runtime.GOOS; thisOs {
	case "linux":
		// Use systemd/systemctl
		return NewSystemdProvider(l, u, c, fs), nil
	case "darwin":
		// Use launchd/launchctl
		return NewLaunchdProvider(l, u, c, fs), nil
	default:
		return nil, fmt.Errorf("cannot configure daemon handler for OS '%s'", thisOs)
	}
}
