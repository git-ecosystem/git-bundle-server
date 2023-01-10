package daemon

import (
	"fmt"
)

type launchd struct{}

func NewLaunchdProvider() DaemonProvider {
	return &launchd{}
}

func (l *launchd) Create(config *DaemonConfig, force bool) error {
	return fmt.Errorf("not implemented")
}

func (l *launchd) Start(label string) error {
	return fmt.Errorf("not implemented")
}

func (l *launchd) Stop(label string) error {
	return fmt.Errorf("not implemented")
}
