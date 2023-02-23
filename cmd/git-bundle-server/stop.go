package main

import (
	"context"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type stopCmd struct {
	logger log.TraceLogger
}

func NewStopCommand(logger log.TraceLogger) argparse.Subcommand {
	return &stopCmd{
		logger: logger,
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
	route := parser.PositionalString("route", "the route for which bundles should stop being generated")
	parser.Parse(ctx, args)

	err := core.RemoveRoute(*route)
	if err != nil {
		s.logger.Error(ctx, err)
	}

	return nil
}
