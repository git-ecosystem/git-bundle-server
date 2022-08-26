package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/github/git-bundle-server/internal/core"
)

func parseRoute(path string) (string, string, string, error) {
	if len(path) == 0 {
		return "", "", "", fmt.Errorf("empty route")
	}

	if path[0] == '/' {
		path = path[1:]
	}

	slash1 := strings.Index(path, "/")
	if slash1 < 0 {
		return "", "", "", fmt.Errorf("route has owner, but no repo")
	}
	slash2 := strings.Index(path[slash1+1:], "/")
	if slash2 < 0 {
		// No trailing slash.
		return path[:slash1], path[slash1+1:], "", nil
	}
	slash2 += slash1 + 1
	slash3 := strings.Index(path[slash2+1:], "/")
	if slash3 >= 0 {
		return "", "", "", fmt.Errorf("path has depth exceeding three")
	}

	return path[:slash1], path[slash1+1 : slash2], path[slash2+1:], nil
}

func serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	owner, repo, file, err := parseRoute(path)
	if err != nil {
		// TODO: report a failure to the user somehow. 404?
		fmt.Printf("Failed to parse route: %s\n", err)
		return
	}

	route := owner + "/" + repo

	repos, err := core.GetRepositories()
	if err != nil {
		// TODO: report a 500
		fmt.Printf("Failed to load routes\n")
		return
	}

	repository, contains := repos[route]
	if !contains {
		// TODO: report a 404
		fmt.Printf("Failed to get route out of repos\n")
		return
	}

	if file == "" {
		file = "bundle-list"
	}

	fileToServe := repository.WebDir + "/" + file
	data, err := os.ReadFile(fileToServe)
	if err != nil {
		// TODO: return a 404
		fmt.Printf("Failed to read file\n")
		return
	}

	fmt.Printf("Successfully serving content for %s/%s\n", route, file)
	w.Write(data)
}

func main() {
	// API routes
	http.HandleFunc("/", serve)

	port := ":8080"
	fmt.Println("Server is running on port" + port)

	// Start server on port specified above
	log.Fatal(http.ListenAndServe(port, nil))
}
