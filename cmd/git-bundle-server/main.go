package main

import (
	"context"
	"log"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
	tracelog "github.com/github/git-bundle-server/internal/log"
)

func all() []argparse.Subcommand {
	return []argparse.Subcommand{
		NewDeleteCommand(),
		NewInitCommand(),
		NewStartCommand(),
		NewStopCommand(),
		NewUpdateCommand(),
		NewUpdateAllCommand(),
		NewWebServerCommand(),
	}
}

func main() {
	tracelog.WithTraceLogger(context.Background(), func(ctx context.Context, logger tracelog.TraceLogger) {
		cmds := all()

		parser := argparse.NewArgParser("git-bundle-server <command> [<options>]")
		parser.SetIsTopLevel(true)
		for _, cmd := range cmds {
			parser.Subcommand(cmd)
		}
		parser.Parse(ctx, os.Args[1:])

		err := parser.InvokeSubcommand(ctx)
		if err != nil {
			log.Fatalf("Failed with error: %s", err)
		}
	})
}
