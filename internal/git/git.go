package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/log"
)

type GitHelper interface {
	CreateBundle(ctx context.Context, repoDir string, filename string) (bool, error)
	CreateBundleFromRefs(ctx context.Context, repoDir string, filename string, refs map[string]string) error
	CreateIncrementalBundle(ctx context.Context, repoDir string, filename string, prereqs []string) (bool, error)
	CloneBareRepo(ctx context.Context, url string, destination string) error
}

type gitHelper struct {
	logger  log.TraceLogger
	cmdExec cmd.CommandExecutor
}

func NewGitHelper(l log.TraceLogger, c cmd.CommandExecutor) GitHelper {
	return &gitHelper{
		logger:  l,
		cmdExec: c,
	}
}

func (g *gitHelper) gitCommand(ctx context.Context, args ...string) error {
	exitCode, err := g.cmdExec.Run(ctx, "git", args,
		cmd.Stdout(os.Stdout),
		cmd.Stderr(os.Stderr),
		cmd.Env([]string{"LC_CTYPE=C"}),
	)

	if err != nil {
		return g.logger.Error(ctx, err)
	} else if exitCode != 0 {
		return g.logger.Errorf(ctx, "'git' exited with status %d", exitCode)
	}

	return nil
}

func (g *gitHelper) gitCommandWithStdin(ctx context.Context, stdinLines []string, args ...string) error {
	buffer := bytes.Buffer{}
	for line := range stdinLines {
		buffer.Write([]byte(stdinLines[line] + "\n"))
	}
	exitCode, err := g.cmdExec.Run(ctx, "git", args,
		cmd.Stdin(&buffer),
		cmd.Stdout(os.Stdout),
		cmd.Stderr(os.Stderr),
		cmd.Env([]string{"LC_CTYPE=C"}),
	)

	if err != nil {
		return g.logger.Error(ctx, err)
	} else if exitCode != 0 {
		return g.logger.Errorf(ctx, "'git' exited with status %d", exitCode)
	}

	return nil
}

func (g *gitHelper) CreateBundle(ctx context.Context, repoDir string, filename string) (bool, error) {
	err := g.gitCommand(ctx,
		"-C", repoDir, "bundle", "create",
		filename, "--branches")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (g *gitHelper) CreateBundleFromRefs(ctx context.Context, repoDir string, filename string, refs map[string]string) error {
	refNames := []string{}

	for ref, oid := range refs {
		err := g.gitCommand(ctx, "-C", repoDir, "branch", "-f", ref, oid)
		if err != nil {
			return fmt.Errorf("failed to create ref %s: %w", ref, err)
		}

		refNames = append(refNames, ref)
	}

	err := g.gitCommandWithStdin(ctx,
		refNames,
		"-C", repoDir, "bundle", "create",
		filename, "--stdin")
	if err != nil {
		return err
	}

	return nil
}

func (g *gitHelper) CreateIncrementalBundle(ctx context.Context, repoDir string, filename string, prereqs []string) (bool, error) {
	err := g.gitCommandWithStdin(ctx,
		prereqs, "-C", repoDir, "bundle", "create",
		filename, "--stdin", "--branches")
	if err != nil {
		if strings.Contains(err.Error(), "Refusing to create empty bundle") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (g *gitHelper) CloneBareRepo(ctx context.Context, url string, destination string) error {
	gitErr := g.gitCommand(ctx, "clone", "--bare", url, destination)

	if gitErr != nil {
		return g.logger.Errorf(ctx, "failed to clone repository: %w", gitErr)
	}

	gitErr = g.gitCommand(ctx, "-C", destination, "config", "remote.origin.fetch", "+refs/heads/*:refs/heads/*")
	if gitErr != nil {
		return g.logger.Errorf(ctx, "failed to configure refspec: %w", gitErr)
	}

	gitErr = g.gitCommand(ctx, "-C", destination, "fetch", "origin")
	if gitErr != nil {
		return g.logger.Errorf(ctx, "failed to fetch latest refs: %w", gitErr)
	}

	return nil
}
