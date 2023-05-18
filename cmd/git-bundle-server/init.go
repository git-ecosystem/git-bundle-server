package main

import (
	"context"
	"fmt"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/bundles"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/git-ecosystem/git-bundle-server/internal/git"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
)

type initCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewInitCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &initCmd{
		logger:    logger,
		container: container,
	}
}

func (initCmd) Name() string {
	return "init"
}

func (initCmd) Description() string {
	return `
Initialize a repository by cloning a bare repo from '<url>', whose bundles
should be hosted at '<route>'.`
}

func (i *initCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(i.logger, "git-bundle-server init <url> [<route>]")
	url := parser.PositionalString("url", "the URL of a repository to clone", true)
	route := parser.PositionalString("route", "the route to host the specified repo", false)
	parser.Parse(ctx, args)

	// Set route value, if needed
	if *route == "" {
		var ok bool
		*route, ok = core.GetRouteFromUrl(*url)
		if !ok {
			parser.Usage(ctx, "Cannot parse route from url '%s'; please specify an explicit route.", *url)
		}
	}

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, i.container)
	bundleProvider := utils.GetDependency[bundles.BundleProvider](ctx, i.container)
	gitHelper := utils.GetDependency[git.GitHelper](ctx, i.container)

	// First, check whether route already exists (enabled or not). If it does,
	// exit with an error.
	allRepos, err := repoProvider.ReadRepositoryStorage(ctx)
	if err != nil {
		return i.logger.Errorf(ctx, "could not load existing routes: %w", err)
	} else if _, ok := allRepos[*route]; ok {
		return i.logger.Errorf(ctx, "route '%s' already exists; if you want to "+
			"overwrite the route, delete it with 'git-bundle-server delete' "+
			"then re-run this command", *route)
	}

	// Create the new route
	repo, err := repoProvider.CreateRepository(ctx, *route)
	if err != nil {
		return i.logger.Error(ctx, err)
	}

	fmt.Printf("Cloning repository from %s\n", *url)
	err = gitHelper.CloneBareRepo(ctx, *url, repo.RepoDir)
	if err != nil {
		return i.logger.Errorf(ctx, "failed to clone repository: %w", err)
	}

	bundle := bundleProvider.CreateInitialBundle(ctx, repo)
	fmt.Printf("Constructing base bundle file at %s\n", bundle.Filename)

	written, gitErr := gitHelper.CreateBundle(ctx, repo.RepoDir, bundle.Filename)
	if gitErr != nil {
		return i.logger.Errorf(ctx, "failed to create bundle: %w", gitErr)
	}
	if !written {
		return i.logger.Errorf(ctx, "refused to write empty bundle. Is the repo empty?")
	}

	list := bundleProvider.CreateSingletonList(ctx, bundle)
	listErr := bundleProvider.WriteBundleList(ctx, list, repo)
	if listErr != nil {
		return i.logger.Errorf(ctx, "failed to write bundle list: %w", listErr)
	}

	cron := utils.GetDependency[utils.CronHelper](ctx, i.container)
	cron.SetCronSchedule(ctx)

	return nil
}
