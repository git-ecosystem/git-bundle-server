package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type Repository struct {
	Route   string
	RepoDir string
	WebDir  string
}

func CreateRepository(route string) (*Repository, error) {
	repos, err := GetRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes file")
	}

	repo, contains := repos[route]
	if contains {
		return &repo, nil
	}

	repodir := reporoot() + route
	web := webroot() + route

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
	repos, err := GetRepositories()
	if err != nil {
		return fmt.Errorf("failed to parse routes file")
	}

	_, contains := repos[route]
	if !contains {
		return fmt.Errorf("route '%s' is not registered", route)
	}

	delete(repos, route)

	return WriteRouteFile(repos)
}

func WriteRouteFile(repos map[string]Repository) error {
	dir := bundleroot()
	routefile := dir + "/routes"

	contents := ""

	for routes := range repos {
		contents = contents + routes + "\n"
	}

	return os.WriteFile(routefile, []byte(contents), 0o600)
}

func GetRepositories() (map[string]Repository, error) {
	repos := make(map[string]Repository)

	dir := bundleroot()
	routefile := dir + "/routes"

	file, err := os.OpenFile(routefile, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		// Assume that the file doesn't exist?
		return repos, nil
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if line == "" || line[0] == '\n' ||
			(err != nil && err != io.EOF) {
			break
		}

		route := line[0 : len(line)-1]

		repo := Repository{
			Route:   route,
			RepoDir: reporoot() + route,
			WebDir:  webroot() + route,
		}
		repos[route] = repo
	}

	return repos, nil
}
