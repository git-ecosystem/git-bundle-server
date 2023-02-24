package daemon

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/log"
)

const serviceTemplate string = `[Unit]
Description={{.Description}}

[Service]
Type=simple
ExecStart={{sq_escape .Program}}{{range .Arguments}} {{sq_escape .}}{{end}}
`

const SystemdUnitNotInstalledErrorCode int = 5

type systemd struct {
	logger     log.TraceLogger
	user       common.UserProvider
	cmdExec    cmd.CommandExecutor
	fileSystem common.FileSystem
}

func NewSystemdProvider(
	l log.TraceLogger,
	u common.UserProvider,
	c cmd.CommandExecutor,
	fs common.FileSystem,
) DaemonProvider {
	return &systemd{
		logger:     l,
		user:       u,
		cmdExec:    c,
		fileSystem: fs,
	}
}

func (s *systemd) reloadDaemon(ctx context.Context) error {
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "daemon-reload")
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	if exitCode != 0 {
		return s.logger.Errorf(ctx, "'systemctl --user daemon-reload' exited with status %d", exitCode)
	}

	return nil
}

func (s *systemd) Create(ctx context.Context, config *DaemonConfig, force bool) error {
	user, err := s.user.CurrentUser()
	if err != nil {
		return s.logger.Errorf(ctx, "could not get current user for systemd service: %w", err)
	}

	// Generate the configuration
	var newServiceUnit bytes.Buffer
	t, err := template.New(config.Label).Funcs(template.FuncMap{
		"sq_escape": func(str string) string {
			return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "\\'"))
		},
	}).Parse(serviceTemplate)
	if err != nil {
		return s.logger.Errorf(ctx, "unable to generate systemd configuration: %w", err)
	}
	t.Execute(&newServiceUnit, config)

	filename := filepath.Join(user.HomeDir, ".config", "systemd", "user", fmt.Sprintf("%s.service", config.Label))

	// Check whether the file exists
	fileExists, err := s.fileSystem.FileExists(filename)
	if err != nil {
		return s.logger.Errorf(ctx, "could not determine whether service unit '%s' exists: %w", config.Label, err)
	}

	if !force && fileExists {
		// File already exists and we aren't forcing a refresh, so we do nothing
		return nil
	}

	// Otherwise, write the new file
	err = s.fileSystem.WriteFile(filename, newServiceUnit.Bytes())
	if err != nil {
		return s.logger.Errorf(ctx, "unable to write service unit: %w", err)
	}

	// Reload the user-scoped service units after adding
	err = s.reloadDaemon(ctx)
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	return nil
}

func (s *systemd) Start(ctx context.Context, label string) error {
	// TODO: warn user if already running
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "start", label)
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	if exitCode != 0 {
		return s.logger.Errorf(ctx, "'systemctl stop' exited with status %d", exitCode)
	}

	return nil
}

func (s *systemd) Stop(ctx context.Context, label string) error {
	// TODO: warn user if already stopped
	exitCode, err := s.cmdExec.Run("systemctl", "--user", "stop", label)
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	if exitCode != 0 && exitCode != SystemdUnitNotInstalledErrorCode {
		return s.logger.Errorf(ctx, "'systemctl stop' exited with status %d", exitCode)
	}

	return nil
}

func (s *systemd) Remove(ctx context.Context, label string) error {
	user, err := s.user.CurrentUser()
	if err != nil {
		return s.logger.Errorf(ctx, "could not get current user for launchd service: %w", err)
	}
	filename := filepath.Join(user.HomeDir, ".config", "systemd", "user", fmt.Sprintf("%s.service", label))

	_, err = s.fileSystem.DeleteFile(filename)
	if err != nil {
		return s.logger.Errorf(ctx, "could not delete service unit: %w", err)
	}

	// Reload the user-scoped service units after removing
	err = s.reloadDaemon(ctx)
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	return nil
}
