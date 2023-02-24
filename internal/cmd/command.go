package cmd

import (
	"fmt"
	"os/exec"
)

type CommandExecutor interface {
	Run(command string, args ...string) (int, error)
}

type commandExecutor struct{}

func NewCommandExecutor() CommandExecutor {
	return &commandExecutor{}
}

func (c *commandExecutor) Run(command string, args ...string) (int, error) {
	exe, err := exec.LookPath(command)
	if err != nil {
		return -1, fmt.Errorf("failed to find '%s' on the path: %w", command, err)
	}

	cmd := exec.Command(exe, args...)

	err = cmd.Start()
	if err != nil {
		return -1, fmt.Errorf("command failed to start: %w", err)
	}

	err = cmd.Wait()
	_, isExitError := err.(*exec.ExitError)

	// If the command succeeded, or ran to completion but returned a nonzero
	// exit code, return non-erroneous result
	if err == nil || isExitError {
		return cmd.ProcessState.ExitCode(), nil
	} else {
		return -1, err
	}
}
