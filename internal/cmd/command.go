package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/github/git-bundle-server/internal/log"
)

type CommandExecutor interface {
	RunStdout(ctx context.Context, command string, args ...string) (int, error)
	RunQuiet(ctx context.Context, command string, args ...string) (int, error)
	Run(ctx context.Context, command string, args []string, settings ...Setting) (int, error)
}

type commandExecutor struct {
	logger log.TraceLogger
}

func NewCommandExecutor(l log.TraceLogger) CommandExecutor {
	return &commandExecutor{
		logger: l,
	}
}

func (c *commandExecutor) buildCmd(ctx context.Context, command string, args ...string) (*exec.Cmd, error) {
	exe, err := exec.LookPath(command)
	if err != nil {
		return nil, c.logger.Errorf(ctx, "failed to find '%s' on the path: %w", command, err)
	}

	cmd := exec.Command(exe, args...)

	return cmd, nil
}

func (c *commandExecutor) applyOptions(ctx context.Context, cmd *exec.Cmd, settings []Setting) {
	for _, setting := range settings {
		switch setting.Key {
		case stdinKey:
			cmd.Stdin = setting.Value.(io.Reader)
		case stdoutKey:
			cmd.Stdout = setting.Value.(io.Writer)
		case stderrKey:
			cmd.Stderr = setting.Value.(io.Writer)
		case envKey:
			env, ok := setting.Value.([]string)
			if !ok {
				panic("incorrect env setting type")
			}
			cmd.Env = append(cmd.Env, env...)
		default:
			panic("invalid cmdSettingKey")
		}
	}
}

func (c *commandExecutor) runCmd(ctx context.Context, cmd *exec.Cmd) (int, error) {
	childReady, childExit := c.logger.ChildProcess(ctx, cmd)
	err := cmd.Start()
	childReady(err)
	if err != nil {
		return -1, c.logger.Errorf(ctx, "command failed to start: %w", err)
	}

	err = cmd.Wait()
	childExit()
	_, isExitError := err.(*exec.ExitError)

	// If the command succeeded, or ran to completion but returned a nonzero
	// exit code, return non-erroneous result
	if err == nil || isExitError {
		return cmd.ProcessState.ExitCode(), nil
	} else {
		return -1, err
	}
}

func (c *commandExecutor) RunStdout(ctx context.Context, command string, args ...string) (int, error) {
	return c.Run(ctx, command, args, Stdout(os.Stdout), Stderr(os.Stderr))
}

func (c *commandExecutor) RunQuiet(ctx context.Context, command string, args ...string) (int, error) {
	return c.Run(ctx, command, args)
}

func (c *commandExecutor) Run(ctx context.Context, command string, args []string, settings ...Setting) (int, error) {
	cmd, err := c.buildCmd(ctx, command, args...)
	if err != nil {
		return -1, err
	}

	c.applyOptions(ctx, cmd, settings)

	return c.runCmd(ctx, cmd)
}
