package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
)

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

		// Configure auth
		middlewareAuthorize := authFunc(nil)

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
