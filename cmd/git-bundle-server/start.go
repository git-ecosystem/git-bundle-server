package main

import (
	"fmt"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
)

type Start struct{}

func (Start) Name() string {
	return "start"
}

func (Start) Description() string {
	return `
Start computing bundles and serving content for the repository at the
specified '<route>'.`
}

func (Start) Run(args []string) error {
	parser := argparse.NewArgParser("git-bundle-server start <route>")
	route := parser.PositionalString("route", "the route for which bundles should be generated")
	parser.Parse(args)

	// CreateRepository registers the route.
	repo, err := core.CreateRepository(*route)
	if err != nil {
		return err
	}

	_, err = os.ReadDir(repo.RepoDir)
	if err != nil {
		return fmt.Errorf("route '%s' appears to have been deleted; use 'init' instead", *route)
	}

	// Make sure we have the global schedule running.
	SetCronSchedule()

	return nil
}
