package main

import (
	"errors"

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
	if len(args) < 1 {
		return errors.New("usage: git-bundle-server stop <route>")
	}

	route := args[0]

	return core.RemoveRoute(route)
}
