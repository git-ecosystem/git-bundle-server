package core

import (
	"context"
	"fmt"
	"os"

	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/log"
)

type Repository struct {
	Route   string
	RepoDir string
	WebDir  string
}

type RepositoryProvider interface {
	CreateRepository(ctx context.Context, route string) (*Repository, error)
	GetRepositories(ctx context.Context) (map[string]Repository, error)
	RemoveRoute(ctx context.Context, route string) error
}

type repoProvider struct {
	logger     log.TraceLogger
	user       common.UserProvider
	fileSystem common.FileSystem
}

func NewRepositoryProvider(logger log.TraceLogger,
	u common.UserProvider,
	fs common.FileSystem,
) RepositoryProvider {
	return &repoProvider{
		logger:     logger,
		user:       u,
		fileSystem: fs,
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

	repodir := reporoot(user) + route
	web := webroot(user) + route

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

	err = r.writeRouteFile(repos)
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

	return r.writeRouteFile(repos)
}

func (r *repoProvider) writeRouteFile(repos map[string]Repository) error {
	user, err := r.user.CurrentUser()
	if err != nil {
		return err
	}
	dir := bundleroot(user)
	routefile := dir + "/routes"

	contents := ""

	for routes := range repos {
		contents = contents + routes + "\n"
	}

	return os.WriteFile(routefile, []byte(contents), 0o600)
}

func (r *repoProvider) GetRepositories(ctx context.Context) (map[string]Repository, error) {
	ctx, exitRegion := r.logger.Region(ctx, "repo", "get_repos") //lint:ignore SA4006 keep ctx up-to-date
	defer exitRegion()

	user, err := r.user.CurrentUser()
	if err != nil {
		return nil, err
	}

	repos := make(map[string]Repository)

	dir := bundleroot(user)
	routefile := dir + "/routes"

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
			RepoDir: reporoot(user) + route,
			WebDir:  webroot(user) + route,
		}
		repos[route] = repo
	}

	return repos, nil
}
