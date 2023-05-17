package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"plugin"
	"strings"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	auth_internal "github.com/git-ecosystem/git-bundle-server/internal/auth"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
	"github.com/git-ecosystem/git-bundle-server/pkg/auth"
)

func getPluginChecksum(pluginPath string) (hash.Hash, error) {
	file, err := os.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	checksum := sha256.New()
	if _, err := io.Copy(checksum, file); err != nil {
		return nil, err
	}

	return checksum, nil
}

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
	case "fixed":
		return auth_internal.NewFixedCredentialAuth(config.Parameters)
	case "plugin":
		if len(config.Path) == 0 {
			return nil, fmt.Errorf("plugin .so is empty")
		}
		if len(config.Initializer) == 0 {
			return nil, fmt.Errorf("plugin initializer symbol is empty")
		}
		if len(config.Checksum) == 0 {
			return nil, fmt.Errorf("SHA256 checksum of plugin file is empty")
		}

		// First, verify plugin checksum matches expected
		// Note: time-of-check/time-of-use could be exploited here (anywhere
		// between the checksum check and invoking the initializer). There's not
		// much we can realistically do about that short of rewriting the plugin
		// package, so we advise users to carefully control access to their
		// system & limit write permissions on their plugin files as a
		// mitigation (see docs/technical/auth-config.md).
		expectedChecksum, err := hex.DecodeString(config.Checksum)
		if err != nil {
			return nil, fmt.Errorf("plugin checksum is invalid: %w", err)
		}
		checksum, err := getPluginChecksum(config.Path)
		if err != nil {
			return nil, fmt.Errorf("could not calculate plugin checksum: %w", err)
		}

		if !bytes.Equal(expectedChecksum, checksum.Sum(nil)) {
			return nil, fmt.Errorf("specified hash does not match plugin checksum")
		}

		// Load the plugin and find the initializer function
		p, err := plugin.Open(config.Path)
		if err != nil {
			return nil, fmt.Errorf("could not load auth plugin: %w", err)
		}

		rawInit, err := p.Lookup(config.Initializer)
		if err != nil {
			return nil, fmt.Errorf("failed to load initializer: %w", err)
		}

		initializer, ok := rawInit.(func(json.RawMessage) (auth.AuthMiddleware, error))
		if !ok {
			return nil, fmt.Errorf("initializer function has incorrect signature")
		}

		// Call the initializer
		return initializer(config.Parameters)
	default:
		return nil, fmt.Errorf("unrecognized auth mode '%s'", config.AuthMode)
	}
}

type authConfig struct {
	AuthMode string `json:"mode"`

	// Plugin-specific settings
	Path        string `json:"path,omitempty"`
	Initializer string `json:"initializer,omitempty"`
	Checksum    string `json:"sha256,omitempty"`

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
