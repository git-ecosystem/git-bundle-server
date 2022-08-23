package git

import (
	"bytes"
	"fmt"
	"git-bundle-server/internal/bundles"
	"git-bundle-server/internal/core"
	"os"
	"os/exec"
	"strings"
)

func GitCommand(args ...string) error {
	git, lookErr := exec.LookPath("git")

	if lookErr != nil {
		return lookErr
	}

	cmd := exec.Command(git, args...)
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

	return err
}

func GitCommandWithStdin(stdinLines []string, args ...string) error {
	git, lookErr := exec.LookPath("git")

	if lookErr != nil {
		return lookErr
	}

	buffer := bytes.Buffer{}
	for line := range stdinLines {
		buffer.Write([]byte(stdinLines[line] + "\n"))
	}

	cmd := exec.Command(git, args...)

	cmd.Stdin = &buffer

	errorBuffer := bytes.Buffer{}
	cmd.Stderr = &errorBuffer
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("git command failed to start: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("git command returned a failure: %w\nstderr: %s", err, errorBuffer.String())
	}

	return err
}

func CreateBundle(repo core.Repository, bundle bundles.Bundle) (bool, error) {
	err := GitCommand(
		"-C", repo.RepoDir, "bundle", "create",
		bundle.Filename, "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CreateIncrementalBundle(repo core.Repository, bundle bundles.Bundle, list bundles.BundleList) (bool, error) {
	lines, err := bundles.GetAllPrereqsForIncrementalBundle(list)
	if err != nil {
		return false, err
	}

	for _, line := range lines {
		fmt.Printf("Sending prereq: %s\n", line)
	}

	err = GitCommandWithStdin(
		lines, "-C", repo.RepoDir, "bundle", "create",
		bundle.Filename, "--stdin", "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
