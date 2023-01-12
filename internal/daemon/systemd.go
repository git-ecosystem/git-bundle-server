package daemon

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/github/git-bundle-server/internal/common"
)

const serviceTemplate string = `[Unit]
Description={{.Description}}

[Service]
Type=simple
ExecStart={{.Program}}
`

type systemd struct {
	user       common.UserProvider
	cmdExec    common.CommandExecutor
	fileSystem common.FileSystem
}

func NewSystemdProvider(
	u common.UserProvider,
	c common.CommandExecutor,
	fs common.FileSystem,
) DaemonProvider {
	return &systemd{
		user:       u,
		cmdExec:    c,
		fileSystem: fs,
	}
}

func (s *systemd) Create(config *DaemonConfig, force bool) error {
	user, err := s.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for systemd service: %w", err)
	}

	// Generate the configuration
	var newServiceUnit bytes.Buffer
	t, err := template.New(config.Label).Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("unable to generate systemd configuration: %w", err)
	}
	t.Execute(&newServiceUnit, config)

	filename := filepath.Join(user.HomeDir, ".config", "systemd", "user", fmt.Sprintf("%s.service", config.Label))

	// Check whether the file exists
	fileExists, err := s.fileSystem.FileExists(filename)
	if err != nil {
		return fmt.Errorf("could not determine whether service unit '%s' exists: %w", config.Label, err)
	}

	if !force && fileExists {
		// File already exists and we aren't forcing a refresh, so we do nothing
		return nil
	}

	// Otherwise, write the new file
	err = s.fileSystem.WriteFile(filename, newServiceUnit.Bytes())
	if err != nil {
		return fmt.Errorf("unable to write service unit: %w", err)
	}

	// Reload the user-scoped service units
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "daemon-reload")
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'systemctl --user daemon-reload' exited with status %d", exitCode)
	}

	return nil
}

func (s *systemd) Start(label string) error {
	// TODO: warn user if already running
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "start", label)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'systemctl stop' exited with status %d", exitCode)
	}

	return nil
}

func (s *systemd) Stop(label string) error {
	// TODO: warn user if already stopped
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "stop", label)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'systemctl stop' exited with status %d", exitCode)
	}

	return nil
}
