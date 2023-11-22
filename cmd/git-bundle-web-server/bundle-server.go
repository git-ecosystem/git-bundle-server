package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/git-ecosystem/git-bundle-server/internal/bundles"
	"github.com/git-ecosystem/git-bundle-server/internal/cmd"
	"github.com/git-ecosystem/git-bundle-server/internal/common"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/git-ecosystem/git-bundle-server/internal/git"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
	"github.com/git-ecosystem/git-bundle-server/pkg/auth"
)

type authFunc func(*http.Request, string, string) auth.AuthResult

type bundleWebServer struct {
	logger             log.TraceLogger
	server             *http.Server
	serverWaitGroup    *sync.WaitGroup
	listenAndServeFunc func() error
	authorize          authFunc
}

func NewBundleWebServer(logger log.TraceLogger,
	port string,
	certFile string, keyFile string,
	tlsMinVersion uint16,
	clientCAFile string,
	middlewareAuthorize authFunc,
) (*bundleWebServer, error) {
	bundleServer := &bundleWebServer{
		logger:          logger,
		serverWaitGroup: &sync.WaitGroup{},
		authorize:       middlewareAuthorize,
	}

	// Configure the http.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/", bundleServer.serve)
	bundleServer.server = &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	// No TLS configuration to be done, return
	if certFile == "" {
		bundleServer.listenAndServeFunc = func() error { return bundleServer.server.ListenAndServe() }
		return bundleServer, nil
	}

	// Configure for TLS
	tlsConfig := &tls.Config{
		MinVersion: tlsMinVersion,
	}
	bundleServer.server.TLSConfig = tlsConfig
	bundleServer.listenAndServeFunc = func() error { return bundleServer.server.ListenAndServeTLS(certFile, keyFile) }

	if clientCAFile != "" {
		caBytes, err := os.ReadFile(clientCAFile)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caBytes)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}

	return bundleServer, nil
}

func (b *bundleWebServer) serve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, exitRegion := b.logger.Region(ctx, "http", "serve")
	defer exitRegion()

	path := r.URL.Path
	owner, repo, filename, err := core.ParseRoute(path, false)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to parse route: %s\n", err)
		return
	}

	route := owner + "/" + repo

	if b.authorize != nil {
		authResult := b.authorize(r, owner, repo)
		if authResult.ApplyResult(w) {
			return
		}
	}

	userProvider := common.NewUserProvider()
	fileSystem := common.NewFileSystem()
	commandExecutor := cmd.NewCommandExecutor(b.logger)
	gitHelper := git.NewGitHelper(b.logger, commandExecutor)
	repoProvider := core.NewRepositoryProvider(b.logger, userProvider, fileSystem, gitHelper)

	repos, err := repoProvider.GetRepositories(ctx)
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

	var fileToServe string
	if filename == "" {
		if path[len(path)-1] == '/' {
			// Trailing slash, so the bundle URIs should be relative to the
			// request's URL as if it were a directory
			fileToServe = filepath.Join(repository.WebDir, bundles.BundleListFilename)
		} else {
			// No trailing slash, so the bundle URIs should be relative to the
			// request's URL as if it were a file
			fileToServe = filepath.Join(repository.WebDir, bundles.RepoBundleListFilename)
		}
	} else if filename == bundles.BundleListFilename || filename == bundles.RepoBundleListFilename {
		// If the request identifies a non-bundle "reserved" file, return 404
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to open file\n")
		return
	} else {
		fileToServe = filepath.Join(repository.WebDir, filename)
	}

	file, err := os.OpenFile(fileToServe, os.O_RDONLY, 0)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Failed to open file\n")
		return
	}

	fmt.Printf("Successfully serving content for %s/%s\n", route, filename)
	http.ServeContent(w, r, filename, time.UnixMicro(0), file)
}

func (b *bundleWebServer) StartServerAsync(ctx context.Context) {
	// Add to wait group
	b.serverWaitGroup.Add(1)

	go func(ctx context.Context) {
		defer b.serverWaitGroup.Done()

		// Return error unless it indicates graceful shutdown
		err := b.listenAndServeFunc()
		if err != nil && err != http.ErrServerClosed {
			b.logger.Fatal(ctx, err)
		}
	}(ctx)

	// Wait 0.1s before reporting that the server is started in case
	// 'listenAndServeFunc' exits immediately.
	//
	// It's a hack, but a necessary one because 'ListenAndServe[TLS]()' doesn't
	// have any mechanism of notifying if it starts successfully, only that it
	// fails. We could get around that by copying/reimplementing those functions
	// with a print statement inserted at the right place, but that's way more
	// cumbersome than just adding a delay here (see:
	// https://stackoverflow.com/questions/53332667/how-to-notify-when-http-server-starts-successfully).
	time.Sleep(time.Millisecond * 100)
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
