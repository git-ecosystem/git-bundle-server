package main

import (
	"context"
	"os"
	"os/exec"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type updateAllCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewUpdateAllCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &updateAllCmd{
		logger:    logger,
		container: container,
	}
}

func (updateAllCmd) Name() string {
	return "update-all"
}

func (updateAllCmd) Description() string {
	return `
For every configured route, run 'git-bundle-server update <options> <route>'.`
}

func (u *updateAllCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(u.logger, "git-bundle-server update-all")
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, u.container)

	repos, err := repoProvider.GetRepositories(ctx)
	if err != nil {
		return u.logger.Error(ctx, err)
	}

	exe, err := os.Executable()
	if err != nil {
		return u.logger.Errorf(ctx, "failed to get path to execuable: %w", err)
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
			return u.logger.Errorf(ctx, "git command failed to start: %w", err)
		}

		err = cmd.Wait()
		if err != nil {
			return u.logger.Errorf(ctx, "git command returned a failure: %w", err)
		}
	}

	return nil
}
