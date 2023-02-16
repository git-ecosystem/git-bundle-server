package core

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/github/git-bundle-server/internal/common"
)

type Repository struct {
	Route   string
	RepoDir string
	WebDir  string
}

func CreateRepository(route string) (*Repository, error) {
	user, err := common.NewUserProvider().CurrentUser()
	if err != nil {
		return nil, err
	}
	fs := common.NewFileSystem()
	repos, err := GetRepositories(user, fs)
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
		log.Fatal("failed to create web directory: ", mkdirErr)
	}

	repo = Repository{
		Route:   route,
		RepoDir: repodir,
		WebDir:  web,
	}

	repos[route] = repo

	err = WriteRouteFile(repos)
	if err != nil {
		return nil, fmt.Errorf("warning: failed to write route file")
	}

	return &repo, nil
}

func RemoveRoute(route string) error {
	user, err := common.NewUserProvider().CurrentUser()
	if err != nil {
		return err
	}
	fs := common.NewFileSystem()
	repos, err := GetRepositories(user, fs)
	if err != nil {
		return fmt.Errorf("failed to parse routes file: %w", err)
	}

	_, contains := repos[route]
	if !contains {
		return fmt.Errorf("route '%s' is not registered", route)
	}

	delete(repos, route)

	return WriteRouteFile(repos)
}

func WriteRouteFile(repos map[string]Repository) error {
	user, err := common.NewUserProvider().CurrentUser()
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

func GetRepositories(user *user.User, fs common.FileSystem) (map[string]Repository, error) {
	repos := make(map[string]Repository)

	dir := bundleroot(user)
	routefile := dir + "/routes"

	lines, err := fs.ReadFileLines(routefile)
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
