package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-ecosystem/git-bundle-server/internal/common"
	"github.com/git-ecosystem/git-bundle-server/internal/git"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
)

type Repository struct {
	Route   string
	RepoDir string
	WebDir  string
}

type RepositoryProvider interface {
	CreateRepository(ctx context.Context, route string) (*Repository, error)
	GetRepositories(ctx context.Context) (map[string]Repository, error)
	WriteAllRoutes(ctx context.Context, repos map[string]Repository) error
	ReadRepositoryStorage(ctx context.Context) (map[string]Repository, error)
	RemoveRoute(ctx context.Context, route string) error
}

type repoProvider struct {
	logger     log.TraceLogger
	user       common.UserProvider
	fileSystem common.FileSystem
	gitHelper  git.GitHelper
}

func NewRepositoryProvider(logger log.TraceLogger,
	u common.UserProvider,
	fs common.FileSystem,
	g git.GitHelper,
) RepositoryProvider {
	return &repoProvider{
		logger:     logger,
		user:       u,
		fileSystem: fs,
		gitHelper:  g,
	}
}

func (r *repoProvider) CreateRepository(ctx context.Context, route string) (*Repository, error) {
	ctx, exitRegion := r.logger.Region(ctx, "repo", "create_repo")
	defer exitRegion()

	user, err := r.user.CurrentUser()
	if err != nil {
		return nil, err
	}

	repos, err := r.GetRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes file: %w", err)
	}

	repo, contains := repos[route]
	if contains {
		return &repo, nil
	}

	repodir := filepath.Join(reporoot(user), route)
	web := filepath.Join(webroot(user), route)

	mkdirErr := os.MkdirAll(web, os.ModePerm)
	if mkdirErr != nil {
		return nil, fmt.Errorf("failed to create web directory: %w", mkdirErr)
	}

	repo = Repository{
		Route:   route,
		RepoDir: repodir,
		WebDir:  web,
	}

	repos[route] = repo

	err = r.WriteAllRoutes(ctx, repos)
	if err != nil {
		return nil, fmt.Errorf("warning: failed to write route file")
	}

	return &repo, nil
}

func (r *repoProvider) RemoveRoute(ctx context.Context, route string) error {
	ctx, exitRegion := r.logger.Region(ctx, "repo", "remove_route")
	defer exitRegion()

	repos, err := r.GetRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse routes file: %w", err)
	}

	_, contains := repos[route]
	if !contains {
		return fmt.Errorf("route '%s' is not registered", route)
	}

	delete(repos, route)

	return r.WriteAllRoutes(ctx, repos)
}

func (r *repoProvider) WriteAllRoutes(ctx context.Context, repos map[string]Repository) error {
	user, err := r.user.CurrentUser()
	if err != nil {
		return err
	}
	routefile := filepath.Join(bundleroot(user), "routes")

	contents := ""
	for routes := range repos {
		contents = contents + routes + "\n"
	}

	return r.fileSystem.WriteFile(routefile, []byte(contents))
}

func (r *repoProvider) GetRepositories(ctx context.Context) (map[string]Repository, error) {
	ctx, exitRegion := r.logger.Region(ctx, "repo", "get_repos") //lint:ignore SA4006 keep ctx up-to-date
	defer exitRegion()

	user, err := r.user.CurrentUser()
	if err != nil {
		return nil, err
	}

	repos := make(map[string]Repository)

	routefile := filepath.Join(bundleroot(user), "routes")

	lines, err := r.fileSystem.ReadFileLines(routefile)
	if err != nil {
		return nil, err
	}
	for _, route := range lines {
		if route == "" {
			continue
		}

		repo := Repository{
			Route:   route,
			RepoDir: filepath.Join(reporoot(user), route),
			WebDir:  filepath.Join(webroot(user), route),
		}
		repos[route] = repo
	}

	return repos, nil
}

func (r *repoProvider) ReadRepositoryStorage(ctx context.Context) (map[string]Repository, error) {
	ctx, exitRegion := r.logger.Region(ctx, "repo", "get_on_disk_repos")
	defer exitRegion()

	user, err := r.user.CurrentUser()
	if err != nil {
		return nil, err
	}

	entries, err := r.fileSystem.ReadDirRecursive(reporoot(user), 2, true)
	if err != nil {
		return nil, err
	}

	repos := make(map[string]Repository)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		_, err = r.gitHelper.GetRemoteUrl(ctx, entry.Path())
		if err != nil {
			continue
		}

		pathElems := strings.Split(entry.Path(), string(os.PathSeparator))
		if len(pathElems) < 2 {
			return nil, r.logger.Errorf(ctx, "invalid repo path '%s'", entry.Path())
		}
		route := strings.Join(pathElems[len(pathElems)-2:], "/")
		repos[route] = Repository{
			Route:   route,
			RepoDir: filepath.Join(reporoot(user), route),
			WebDir:  filepath.Join(webroot(user), route),
		}
	}

	return repos, nil
}
