package main

import (
	"context"
	"os"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type deleteCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewDeleteCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &deleteCmd{
		logger:    logger,
		container: container,
	}
}

func (deleteCmd) Name() string {
	return "delete"
}

func (deleteCmd) Description() string {
	return `
Remove the configuration for the given '<route>' and delete its repository
data.`
}

func (d *deleteCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(d.logger, "git-bundle-server delete <route>")
	route := parser.PositionalString("route", "the route to delete", true)
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, d.container)

	repo, err := repoProvider.CreateRepository(ctx, *route)
	if err != nil {
		return d.logger.Error(ctx, err)
	}

	err = repoProvider.RemoveRoute(ctx, *route)
	if err != nil {
		return d.logger.Error(ctx, err)
	}

	err = os.RemoveAll(repo.WebDir)
	if err != nil {
		return d.logger.Error(ctx, err)
	}

	err = os.RemoveAll(repo.RepoDir)
	if err != nil {
		return d.logger.Error(ctx, err)
	}

	return nil
}
