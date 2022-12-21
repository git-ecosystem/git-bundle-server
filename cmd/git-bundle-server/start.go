package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/github/git-bundle-server/internal/core"
)

type Start struct{}

func (Start) subcommand() string {
	return "start"
}

func (Start) run(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: git-bundle-server start <route>")
	}

	route := args[0]

	// CreateRepository registers the route.
	repo, err := core.CreateRepository(route)
	if err != nil {
		return err
	}

	_, err = os.ReadDir(repo.RepoDir)
	if err != nil {
		return fmt.Errorf("route '%s' appears to have been deleted; use 'init' instead", route)
	}

	// Make sure we have the global schedule running.
	SetCronSchedule()

	return nil
}