package main

import (
	"errors"
	"fmt"
	"git-bundle-server/internal/git"
	"log"
	"os"
	"time"
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

	repo := reporoot() + route
	web := webroot() + route

	mkdirErr := os.MkdirAll(web, os.ModePerm)
	if mkdirErr != nil {
		log.Fatal("Failed to create web directory: ", mkdirErr)
	}

	fmt.Printf("Cloning repository from %s\n", url)
	gitErr := git.GitCommand("clone", "--mirror", url, repo)

	if gitErr != nil {
		return gitErr
	}

	timestamp := time.Now().UTC().Unix()
	bundleFile := web + "/bundle-" + fmt.Sprint(timestamp) + ".bundle"
	fmt.Printf("Constructing base bundle file at %s\n", bundleFile)

	gitErr = git.GitCommand("-C", repo, "bundle", "create", bundleFile, "--all")
	if gitErr != nil {
		return gitErr
	}

	return nil
}
