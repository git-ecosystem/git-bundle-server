package cmd

import (
	"context"
	"os/exec"

	"github.com/github/git-bundle-server/internal/log"
)

type CommandExecutor interface {
	Run(ctx context.Context, command string, args ...string) (int, error)
}

type commandExecutor struct {
	logger log.TraceLogger
}

func NewCommandExecutor(l log.TraceLogger) CommandExecutor {
	return &commandExecutor{
		logger: l,
	}
}

func (c *commandExecutor) Run(ctx context.Context, command string, args ...string) (int, error) {
	exe, err := exec.LookPath(command)
	if err != nil {
		return -1, c.logger.Errorf(ctx, "failed to find '%s' on the path: %w", command, err)
	}

	cmd := exec.Command(exe, args...)

	err = cmd.Start()
	if err != nil {
		return -1, c.logger.Errorf(ctx, "command failed to start: %w", err)
	}

	err = cmd.Wait()
	_, isExitError := err.(*exec.ExitError)

	// If the command succeeded, or ran to completion but returned a nonzero
	// exit code, return non-erroneous result
	if err == nil || isExitError {
		return cmd.ProcessState.ExitCode(), nil
	} else {
		return -1, c.logger.Error(ctx, err)
	}
}
