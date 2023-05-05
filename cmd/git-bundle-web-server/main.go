package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
	"github.com/git-ecosystem/git-bundle-server/pkg/auth"
)

func parseAuthConfig(configPath string) (auth.AuthMiddleware, error) {
	var config authConfig
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(config.AuthMode) {
	default:
		return nil, fmt.Errorf("unrecognized auth mode '%s'", config.AuthMode)
	}
}

type authConfig struct {
	AuthMode string `json:"mode"`

	// Per-middleware custom config
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

func main() {
	log.WithTraceLogger(context.Background(), func(ctx context.Context, logger log.TraceLogger) {
		parser := argparse.NewArgParser(logger, "git-bundle-web-server [--port <port>] [--cert <filename> --key <filename>]")
		flags, validate := utils.WebServerFlags(parser)
		flags.VisitAll(func(f *flag.Flag) {
			parser.Var(f.Value, f.Name, f.Usage)
		})

		parser.Parse(ctx, os.Args[1:])
		validate(ctx)

		// Get the flag values
		port := utils.GetFlagValue[string](parser, "port")
		cert := utils.GetFlagValue[string](parser, "cert")
		key := utils.GetFlagValue[string](parser, "key")
		tlsMinVersion := utils.GetFlagValue[uint16](parser, "tls-version")
		clientCA := utils.GetFlagValue[string](parser, "client-ca")
		authConfig := utils.GetFlagValue[string](parser, "auth-config")

		// Configure auth
		var err error
		middlewareAuthorize := authFunc(nil)
		if authConfig != "" {
			middleware, err := parseAuthConfig(authConfig)
			if err != nil {
				logger.Fatalf(ctx, "Invalid auth config: %w", err)
			}
			if middleware == nil {
				// Up until this point, everything indicates that a user intends
				// to use - and has properly configured - custom auth. However,
				// despite there being no error from the initializer, the
				// middleware was empty. This is almost certainly incorrect, so
				// we exit.
				logger.Fatalf(ctx, "Middleware is nil, but no error was returned from initializer. "+
					"If no middleware is desired, remove the --auth-config option.")
			}
			middlewareAuthorize = middleware.Authorize
		}

		// Configure the server
		bundleServer, err := NewBundleWebServer(logger,
			port,
			cert, key,
			tlsMinVersion,
			clientCA,
			middlewareAuthorize,
		)
		if err != nil {
			logger.Fatal(ctx, err)
		}

		// Start the server asynchronously
		bundleServer.StartServerAsync(ctx)

		// Intercept interrupt signals
		bundleServer.HandleSignalsAsync(ctx)

		// Wait for server to shut down
		bundleServer.Wait()

		fmt.Println("Shutdown complete")
	})
}
