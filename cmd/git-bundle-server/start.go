package main

import (
	"context"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type startCmd struct {
	logger log.TraceLogger
}

func NewStartCommand(logger log.TraceLogger) argparse.Subcommand {
	return &startCmd{
		logger: logger,
	}
}

func (startCmd) Name() string {
	return "start"
}

func (startCmd) Description() string {
	return `
Start computing bundles and serving content for the repository at the
specified '<route>'.`
}

func (s *startCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(s.logger, "git-bundle-server start <route>")
	route := parser.PositionalString("route", "the route for which bundles should be generated")
	parser.Parse(ctx, args)

	// CreateRepository registers the route.
	repo, err := core.CreateRepository(*route)
	if err != nil {
		return s.logger.Error(ctx, err)
	}

	_, err = os.ReadDir(repo.RepoDir)
	if err != nil {
		return s.logger.Errorf(ctx, "route '%s' appears to have been deleted; use 'init' instead", *route)
	}

	// Make sure we have the global schedule running.
	SetCronSchedule()

	return nil
}
