package main

import (
	"context"
	"fmt"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/core"
)

type updateCmd struct{}

func NewUpdateCommand() argparse.Subcommand {
	return &updateAllCmd{}
}

func (updateCmd) Name() string {
	return "update"
}

func (updateCmd) Description() string {
	return `
For the repository in the current directory (or the one specified by
'<route>'), fetch the latest content from the remote, create a new set of
bundles, and update the bundle list.`
}

func (updateCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser("git-bundle-server update <route>")
	route := parser.PositionalString("route", "the route to update")
	parser.Parse(ctx, args)

	repo, err := core.CreateRepository(*route)
	if err != nil {
		return err
	}

	list, err := bundles.GetBundleList(repo)
	if err != nil {
		return fmt.Errorf("failed to load bundle list: %w", err)
	}

	fmt.Printf("Creating new incremental bundle\n")
	bundle, err := bundles.CreateIncrementalBundle(repo, list)
	if err != nil {
		return err
	}

	// Nothing new!
	if bundle == nil {
		return nil
	}

	list.Bundles[bundle.CreationToken] = *bundle

	fmt.Printf("Collapsing bundle list\n")
	err = bundles.CollapseList(repo, list)
	if err != nil {
		return err
	}

	fmt.Printf("Writing updated bundle list\n")
	listErr := bundles.WriteBundleList(list, repo)
	if listErr != nil {
		return fmt.Errorf("failed to write bundle list: %w", listErr)
	}

	return nil
}
