package main

import (
	"errors"

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

	// CreateRepository re-reigsters the route.
	_, err := core.CreateRepository(route)
	if err != nil {
		return err
	}

	// Make sure we have the global schedule running.
	SetCronSchedule()

	return nil
}
