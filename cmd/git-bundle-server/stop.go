package main

import (
	"errors"

	"github.com/github/git-bundle-server/internal/core"
)

type Stop struct{}

func (Stop) subcommand() string {
	return "stop"
}

func (Stop) run(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: git-bundle-server stop <route>")
	}

	route := args[0]

	return core.RemoveRoute(route)
}
