package main

import (
	"context"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/core"
)

type deleteCmd struct{}

func NewDeleteCommand() argparse.Subcommand {
	return &deleteCmd{}
}

func (deleteCmd) Name() string {
	return "delete"
}

func (deleteCmd) Description() string {
	return `
Remove the configuration for the given '<route>' and delete its repository
data.`
}

func (deleteCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser("git-bundle-server delete <route>")
	route := parser.PositionalString("route", "the route to delete")
	parser.Parse(ctx, args)

	repo, err := core.CreateRepository(*route)
	if err != nil {
		return err
	}

	err = core.RemoveRoute(*route)
	if err != nil {
		return err
	}

	err = os.RemoveAll(repo.WebDir)
	if err != nil {
		return err
	}

	err = os.RemoveAll(repo.RepoDir)
	if err != nil {
		return err
	}

	return nil
}
