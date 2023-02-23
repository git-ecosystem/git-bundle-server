package main

import (
	"context"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/log"
)

func all(logger log.TraceLogger) []argparse.Subcommand {
	return []argparse.Subcommand{
		NewDeleteCommand(logger),
		NewInitCommand(logger),
		NewStartCommand(logger),
		NewStopCommand(logger),
		NewUpdateCommand(logger),
		NewUpdateAllCommand(logger),
		NewWebServerCommand(logger),
	}
}

func main() {
	log.WithTraceLogger(context.Background(), func(ctx context.Context, logger log.TraceLogger) {
		cmds := all(logger)

		parser := argparse.NewArgParser(logger, "git-bundle-server <command> [<options>]")
		parser.SetIsTopLevel(true)
		for _, cmd := range cmds {
			parser.Subcommand(cmd)
		}
		parser.Parse(ctx, os.Args[1:])

		err := parser.InvokeSubcommand(ctx)
		if err != nil {
			logger.Fatalf(ctx, "Failed with error: %s", err)
		}
	})
}
