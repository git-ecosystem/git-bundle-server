package core

import (
	"log"
	"os"
)

type Repository struct {
	Route   string
	RepoDir string
	WebDir  string
}

func GetRepository(route string) Repository {
	repo := reporoot() + route
	web := webroot() + route

	mkdirErr := os.MkdirAll(web, os.ModePerm)
	if mkdirErr != nil {
		log.Fatal("failed to create web directory: ", mkdirErr)
	}

	return Repository{
		Route:   route,
		RepoDir: repo,
		WebDir:  web,
	}
}
