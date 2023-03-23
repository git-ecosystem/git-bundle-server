package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/git"
	"github.com/github/git-bundle-server/internal/log"
)

type listCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewListCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &listCmd{
		logger:    logger,
		container: container,
	}
}

func (listCmd) Name() string {
	return "list"
}

func (listCmd) Description() string {
	return `
List the routes registered to the bundle server.`
}

func (l *listCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(l.logger, "git-bundle-server list [--name-only]")
	nameOnly := parser.Bool("name-only", false, "print only the names of configured routes")
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, l.container)
	gitHelper := utils.GetDependency[git.GitHelper](ctx, l.container)

	repos, err := repoProvider.GetRepositories(ctx)
	if err != nil {
		return l.logger.Error(ctx, err)
	}

	for _, repo := range repos {
		info := []string{repo.Route}
		if !*nameOnly {
			remote, err := gitHelper.GetRemoteUrl(ctx, repo.RepoDir)
			if err != nil {
				return l.logger.Error(ctx, err)
			}
			info = append(info, remote)
		}

		// Join with space & tab to ensure each element of the info array is
		// separated by at least two spaces (for better readability).
		fmt.Println(strings.Join(info, " \t"))
	}

	return nil
}
