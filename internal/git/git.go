package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetExecCommand(args []string) (*exec.Cmd, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("failed to find 'git' on the path: %w", err)
	}

	cmd := exec.Command(git, args...)
	cmd.Env = append(cmd.Env, "LC_CTYPE=C")

	return cmd, nil
}

func GitCommand(args ...string) error {
	cmd, err := GetExecCommand(args)
	if err != nil {
		return err
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Start()
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
	cmd, err := GetExecCommand(args)
	if err != nil {
		return err
	}

	buffer := bytes.Buffer{}
	for line := range stdinLines {
		buffer.Write([]byte(stdinLines[line] + "\n"))
	}

	cmd.Stdin = &buffer

	errorBuffer := bytes.Buffer{}
	cmd.Stderr = &errorBuffer
	cmd.Stdout = os.Stdout

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("git command failed to start: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("git command returned a failure: %w\nstderr: %s", err, errorBuffer.String())
	}

	return err
}

func CreateBundle(repoDir string, filename string) (bool, error) {
	err := GitCommand(
		"-C", repoDir, "bundle", "create",
		filename, "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CreateBundleFromRefs(repoDir string, filename string, refs map[string]string) error {
	refNames := []string{}

	for ref, oid := range refs {
		err := GitCommand("-C", repoDir, "branch", "-f", ref, oid)
		if err != nil {
			return fmt.Errorf("failed to create ref %s: %w", ref, err)
		}

		refNames = append(refNames, ref)
	}

	err := GitCommandWithStdin(
		refNames,
		"-C", repoDir, "bundle", "create",
		filename, "--stdin")
	if err != nil {
		return err
	}

	return nil
}

func CreateIncrementalBundle(repoDir string, filename string, prereqs []string) (bool, error) {
	err := GitCommandWithStdin(
		prereqs, "-C", repoDir, "bundle", "create",
		filename, "--stdin", "--all")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
