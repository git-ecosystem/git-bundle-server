package main

import (
	"context"
	"fmt"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/cmd"
	"github.com/git-ecosystem/git-bundle-server/internal/common"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
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
	fileSystem := utils.GetDependency[common.FileSystem](ctx, u.container)
	commandExecutor := utils.GetDependency[cmd.CommandExecutor](ctx, u.container)

	repos, err := repoProvider.GetRepositories(ctx)
	if err != nil {
		return u.logger.Error(ctx, err)
	}

	exe, err := fileSystem.GetLocalExecutable("git-bundle-server")
	if err != nil {
		return u.logger.Errorf(ctx, "failed to get path to execuable: %w", err)
	}

	subargs := []string{"update", ""}
	subargs = append(subargs, args...)

	for route := range repos {
		subargs[1] = route
		fmt.Printf("*** Updating %s ***\n", route)
		exitCode, err := commandExecutor.RunStdout(ctx, exe, subargs...)
		if err != nil {
			return u.logger.Error(ctx, err)
		} else if exitCode != 0 {
			return u.logger.Errorf(ctx, "git-bundle-server update exited with status %d", exitCode)
		}
		fmt.Print("\n")
	}

	return nil
}
