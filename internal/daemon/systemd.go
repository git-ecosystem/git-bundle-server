package daemon

import (
	"fmt"
)

type systemd struct{}

func NewSystemdProvider() DaemonProvider {
	return &systemd{}
}

func (s *systemd) Create(config *DaemonConfig, force bool) error {
	return fmt.Errorf("not implemented")
}

func (s *systemd) Start(label string) error {
	return fmt.Errorf("not implemented")
}

func (s *systemd) Stop(label string) error {
	return fmt.Errorf("not implemented")
}
