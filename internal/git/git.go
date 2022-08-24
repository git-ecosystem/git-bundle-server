package git

import (
	"bytes"
	"fmt"
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

func CreateBundle(repo core.Repository, filename string) (bool, error) {
	err := GitCommand(
		"-C", repo.RepoDir, "bundle", "create",
		filename, "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CreateBundleFromRefs(repo core.Repository, filename string, refs map[string]string) error {
	refNames := []string{}

	for ref, oid := range refs {
		err := GitCommand("-C", repo.RepoDir, "branch", "-f", ref, oid)
		if err != nil {
			return fmt.Errorf("failed to create ref %s: %w", ref, err)
		}

		refNames = append(refNames, ref)
	}

	err := GitCommandWithStdin(
		refNames,
		"-C", repo.RepoDir, "bundle", "create",
		filename, "--stdin")
	if err != nil {
		return err
	}

	return nil
}

func CreateIncrementalBundle(repo core.Repository, filename string, prereqs []string) (bool, error) {
	err := GitCommandWithStdin(
		prereqs, "-C", repo.RepoDir, "bundle", "create",
		filename, "--stdin", "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
