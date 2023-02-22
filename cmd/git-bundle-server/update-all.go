package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/core"
)

type updateAllCmd struct{}

func NewUpdateAllCommand() argparse.Subcommand {
	return &updateAllCmd{}
}

func (updateAllCmd) Name() string {
	return "update-all"
}

func (updateAllCmd) Description() string {
	return `
For every configured route, run 'git-bundle-server update <options> <route>'.`
}

func (updateAllCmd) Run(ctx context.Context, args []string) error {
	user, err := common.NewUserProvider().CurrentUser()
	if err != nil {
		return err
	}
	fs := common.NewFileSystem()

	parser := argparse.NewArgParser("git-bundle-server update-all")
	parser.Parse(ctx, args)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get path to execuable: %w", err)
	}

	repos, err := core.GetRepositories(user, fs)
	if err != nil {
		return err
	}

	subargs := []string{"update", ""}
	subargs = append(subargs, args...)

	for route := range repos {
		subargs[1] = route
		cmd := exec.Command(exe, subargs...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("git command failed to start: %w", err)
		}

		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("git command returned a failure: %w", err)
		}
	}

	return nil
}
