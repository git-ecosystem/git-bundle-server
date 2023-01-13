package main

import (
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
)

type Stop struct{}

func (Stop) Name() string {
	return "stop"
}

func (Stop) Description() string {
	return `
Stop computing bundles or serving content for the repository at the
specified '<route>'.`
}

func (Stop) Run(args []string) error {
	parser := argparse.NewArgParser("git-bundle-server stop <route>")
	route := parser.PositionalString("route", "the route for which bundles should stop being generated")
	parser.Parse(args)

	return core.RemoveRoute(*route)
}
