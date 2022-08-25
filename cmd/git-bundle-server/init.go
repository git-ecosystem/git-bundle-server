package main

import (
	"errors"
	"fmt"
	"git-bundle-server/internal/bundles"
	"git-bundle-server/internal/core"
	"git-bundle-server/internal/git"
)

type Init struct{}

func (Init) subcommand() string {
	return "init"
}

func (Init) run(args []string) error {
	if len(args) < 2 {
		// TODO: allow parsing <route> out of <url>
		return errors.New("usage: git-bundle-server init <url> <route>")
	}

	url := args[0]
	route := args[1]

	repo := core.GetRepository(route)

	fmt.Printf("Cloning repository from %s\n", url)
	gitErr := git.GitCommand("clone", "--mirror", url, repo.RepoDir)

	if gitErr != nil {
		return fmt.Errorf("failed to clone repository: %w", gitErr)
	}

	bundle := bundles.CreateInitialBundle(repo)
	fmt.Printf("Constructing base bundle file at %s\n", bundle.Filename)

	written, gitErr := git.CreateBundle(repo, bundle)
	if gitErr != nil {
		return fmt.Errorf("failed to create bundle: %w", gitErr)
	}
	if !written {
		return fmt.Errorf("refused to write empty bundle. Is the repo empty?")
	}

	list := bundles.SingletonList(bundle)
	listErr := bundles.WriteBundleList(list, repo)
	if listErr != nil {
		return fmt.Errorf("failed to write bundle list: %w", listErr)
	}

	return nil
}
