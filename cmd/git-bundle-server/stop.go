package main

import (
	"context"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type stopCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewStopCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &stopCmd{
		logger:    logger,
		container: container,
	}
}

func (stopCmd) Name() string {
	return "stop"
}

func (stopCmd) Description() string {
	return `
Stop computing bundles or serving content for the repository at the
specified '<route>'.`
}

func (s *stopCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(s.logger, "git-bundle-server stop <route>")
	route := parser.PositionalString("route", "the route for which bundles should stop being generated", true)
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, s.container)

	err := repoProvider.RemoveRoute(ctx, *route)
	if err != nil {
		s.logger.Error(ctx, err)
	}

	return nil
}
