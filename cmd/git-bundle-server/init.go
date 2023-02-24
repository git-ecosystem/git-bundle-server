package main

import (
	"context"
	"fmt"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/git"
	"github.com/github/git-bundle-server/internal/log"
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
	parser := argparse.NewArgParser(i.logger, "git-bundle-server init <url> <route>")
	url := parser.PositionalString("url", "the URL of a repository to clone")
	// TODO: allow parsing <route> out of <url>
	route := parser.PositionalString("route", "the route to host the specified repo")
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, i.container)
	bundleProvider := utils.GetDependency[bundles.BundleProvider](ctx, i.container)

	repo, err := repoProvider.CreateRepository(ctx, *route)
	if err != nil {
		return i.logger.Error(ctx, err)
	}

	fmt.Printf("Cloning repository from %s\n", *url)
	gitErr := git.GitCommand("clone", "--bare", *url, repo.RepoDir)

	if gitErr != nil {
		return i.logger.Errorf(ctx, "failed to clone repository: %w", gitErr)
	}

	gitErr = git.GitCommand("-C", repo.RepoDir, "config", "remote.origin.fetch", "+refs/heads/*:refs/heads/*")
	if gitErr != nil {
		return i.logger.Errorf(ctx, "failed to configure refspec: %w", gitErr)
	}

	gitErr = git.GitCommand("-C", repo.RepoDir, "fetch", "origin")
	if gitErr != nil {
		return i.logger.Errorf(ctx, "failed to fetch latest refs: %w", gitErr)
	}

	bundle := bundleProvider.CreateInitialBundle(ctx, repo)
	fmt.Printf("Constructing base bundle file at %s\n", bundle.Filename)

	written, gitErr := git.CreateBundle(repo.RepoDir, bundle.Filename)
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
