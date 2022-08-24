package main

import (
	"fmt"
	"git-bundle-server/internal/core"
	"os"
	"os/exec"
)

type UpdateAll struct{}

func (UpdateAll) subcommand() string {
	return "update-all"
}

func (UpdateAll) run(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get path to execuable: %w", err)
	}

	repos, err := core.GetRepositories()
	if err != nil {
		return err
	}

	for route := range repos {
		cmd := exec.Command(exe, "update", route)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("git command failed to start: %w", err)
		}

		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("git command returned a failure: %w", err)
		}
	}

	return nil
}
