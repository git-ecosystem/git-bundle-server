package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

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
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to parse route: %s\n", err)
		return
	}

	route := owner + "/" + repo

	repos, err := core.GetRepositories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Failed to load routes\n")
		return
	}

	repository, contains := repos[route]
	if !contains {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to get route out of repos\n")
		return
	}

	if file == "" {
		file = "bundle-list"
	}

	fileToServe := repository.WebDir + "/" + file
	data, err := os.ReadFile(fileToServe)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to read file\n")
		return
	}

	fmt.Printf("Successfully serving content for %s/%s\n", route, file)
	w.Write(data)
}

func startServer(server *http.Server, serverWaitGroup *sync.WaitGroup) {
	// Add to wait group
	serverWaitGroup.Add(1)

	go func() {
		defer serverWaitGroup.Done()

		// Return error unless it indicates graceful shutdown
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	fmt.Println("Server is running at address " + server.Addr)
}

func main() {
	// Configure the server
	mux := http.NewServeMux()
	mux.HandleFunc("/", serve)
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	serverWaitGroup := &sync.WaitGroup{}

	// Start the server asynchronously
	startServer(server, serverWaitGroup)

	// Intercept interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Starting graceful server shutdown...")
		server.Shutdown(context.Background())
	}()

	// Wait for server to shut down
	serverWaitGroup.Wait()

	fmt.Println("Shutdown complete")
}
