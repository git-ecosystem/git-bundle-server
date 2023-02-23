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

	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/core"
)

type bundleWebServer struct {
	server             *http.Server
	serverWaitGroup    *sync.WaitGroup
	listenAndServeFunc func() error
}

func NewBundleWebServer(port string, certFile string, keyFile string) *bundleWebServer {
	bundleServer := &bundleWebServer{
		serverWaitGroup: &sync.WaitGroup{},
	}

	// Configure the http.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/", bundleServer.serve)
	bundleServer.server = &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	if certFile != "" {
		bundleServer.listenAndServeFunc = func() error { return bundleServer.server.ListenAndServeTLS(certFile, keyFile) }
	} else {
		bundleServer.listenAndServeFunc = func() error { return bundleServer.server.ListenAndServe() }
	}

	return bundleServer
}

func (b *bundleWebServer) parseRoute(ctx context.Context, path string) (string, string, string, error) {
	elements := strings.FieldsFunc(path, func(char rune) bool { return char == '/' })
	switch len(elements) {
	case 0:
		return "", "", "", fmt.Errorf("empty route")
	case 1:
		return "", "", "", fmt.Errorf("route has owner, but no repo")
	case 2:
		return elements[0], elements[1], "", nil
	case 3:
		return elements[0], elements[1], elements[2], nil
	default:
		return "", "", "", fmt.Errorf("path has depth exceeding three")
	}
}

func (b *bundleWebServer) serve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := common.NewUserProvider().CurrentUser()
	if err != nil {
		return
	}
	fs := common.NewFileSystem()
	path := r.URL.Path

	owner, repo, file, err := b.parseRoute(ctx, path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to parse route: %s\n", err)
		return
	}

	route := owner + "/" + repo

	repos, err := core.GetRepositories(user, fs)
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

func (b *bundleWebServer) StartServerAsync(ctx context.Context) {
	// Add to wait group
	b.serverWaitGroup.Add(1)

	go func(ctx context.Context) {
		defer b.serverWaitGroup.Done()

		// Return error unless it indicates graceful shutdown
		err := b.listenAndServeFunc()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}(ctx)

	fmt.Println("Server is running at address " + b.server.Addr)
}

func (b *bundleWebServer) HandleSignalsAsync(ctx context.Context) {
	// Intercept interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func(ctx context.Context) {
		<-c
		fmt.Println("Starting graceful server shutdown...")
		b.server.Shutdown(ctx)
	}(ctx)
}

func (b *bundleWebServer) Wait() {
	b.serverWaitGroup.Wait()
}
