package main

import (
	"context"
	"os"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/log"
)

func all(logger log.TraceLogger) []argparse.Subcommand {
	container := utils.BuildGitBundleServerContainer(logger)

	return []argparse.Subcommand{
		NewDeleteCommand(logger, container),
		NewInitCommand(logger, container),
		NewStartCommand(logger, container),
		NewStopCommand(logger, container),
		NewUpdateCommand(logger, container),
		NewUpdateAllCommand(logger, container),
		NewListCommand(logger, container),
		NewWebServerCommand(logger, container),
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
