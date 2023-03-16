package main

import (
	"context"
	"fmt"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/log"
)

type updateCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewUpdateCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &updateCmd{
		logger:    logger,
		container: container,
	}
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

func (u *updateCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(u.logger, "git-bundle-server update <route>")
	route := parser.PositionalString("route", "the route to update", true)
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, u.container)
	bundleProvider := utils.GetDependency[bundles.BundleProvider](ctx, u.container)

	repo, err := repoProvider.CreateRepository(ctx, *route)
	if err != nil {
		return u.logger.Error(ctx, err)
	}

	list, err := bundleProvider.GetBundleList(ctx, repo)
	if err != nil {
		return u.logger.Errorf(ctx, "failed to load bundle list: %w", err)
	}

	fmt.Printf("Creating new incremental bundle\n")
	bundle, err := bundleProvider.CreateIncrementalBundle(ctx, repo, list)
	if err != nil {
		return u.logger.Error(ctx, err)
	}

	// Nothing new!
	if bundle == nil {
		return nil
	}

	list.Bundles[bundle.CreationToken] = *bundle

	fmt.Printf("Collapsing bundle list\n")
	err = bundleProvider.CollapseList(ctx, repo, list)
	if err != nil {
		return u.logger.Error(ctx, err)
	}

	fmt.Printf("Writing updated bundle list\n")
	listErr := bundleProvider.WriteBundleList(ctx, list, repo)
	if listErr != nil {
		return u.logger.Errorf(ctx, "failed to write bundle list: %w", listErr)
	}

	return nil
}
