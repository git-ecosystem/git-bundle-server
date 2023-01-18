package main

import (
	"errors"
	"fmt"

	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/core"
)

type Update struct{}

func (Update) Name() string {
	return "update"
}

func (Update) Description() string {
	return `
For the repository in the current directory (or the one specified by
'<route>'), fetch the latest content from the remote, create a new set of
bundles, and update the bundle list.`
}

func (Update) Run(args []string) error {
	if len(args) != 1 {
		// TODO: allow parsing <route> out of <url>
		return errors.New("usage: git-bundle-server update <route>")
	}

	route := args[0]
	repo, err := core.CreateRepository(route)
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
