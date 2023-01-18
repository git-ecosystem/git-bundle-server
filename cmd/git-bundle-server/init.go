package main

import (
	"errors"
	"fmt"

	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/git"
)

type Init struct{}

func (Init) Name() string {
	return "init"
}

func (Init) Description() string {
	return `
Initialize a repository by cloning a bare repo from '<url>', whose bundles
should be hosted at '<route>'.`
}

func (Init) Run(args []string) error {
	if len(args) < 2 {
		// TODO: allow parsing <route> out of <url>
		return errors.New("usage: git-bundle-server init <url> <route>")
	}

	url := args[0]
	route := args[1]

	repo, err := core.CreateRepository(route)
	if err != nil {
		return err
	}

	fmt.Printf("Cloning repository from %s\n", url)
	gitErr := git.GitCommand("clone", "--bare", url, repo.RepoDir)

	if gitErr != nil {
		return fmt.Errorf("failed to clone repository: %w", gitErr)
	}

	gitErr = git.GitCommand("-C", repo.RepoDir, "config", "remote.origin.fetch", "+refs/heads/*:refs/heads/*")
	if gitErr != nil {
		return fmt.Errorf("failed to configure refspec: %w", gitErr)
	}

	gitErr = git.GitCommand("-C", repo.RepoDir, "fetch", "origin")
	if gitErr != nil {
		return fmt.Errorf("failed to fetch latest refs: %w", gitErr)
	}

	bundle := bundles.CreateInitialBundle(repo)
	fmt.Printf("Constructing base bundle file at %s\n", bundle.Filename)

	written, gitErr := git.CreateBundle(repo.RepoDir, bundle.Filename)
	if gitErr != nil {
		return fmt.Errorf("failed to create bundle: %w", gitErr)
	}
	if !written {
		return fmt.Errorf("refused to write empty bundle. Is the repo empty?")
	}

	list := bundles.CreateSingletonList(bundle)
	listErr := bundles.WriteBundleList(list, repo)
	if listErr != nil {
		return fmt.Errorf("failed to write bundle list: %w", listErr)
	}

	SetCronSchedule()

	return nil
}
