package daemon

import (
	"fmt"
	"runtime"
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

func NewDaemonProvider() (DaemonProvider, error) {
	switch thisOs := runtime.GOOS; thisOs {
	case "linux":
		// Use systemd/systemctl
		return NewSystemdProvider(), nil
	case "darwin":
		// Use launchd/launchctl
		return NewLaunchdProvider(), nil
	default:
		return nil, fmt.Errorf("cannot configure daemon handler for OS '%s'", thisOs)
	}
}
