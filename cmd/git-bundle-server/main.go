package main

import (
	"context"
	"log"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
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
	ctx := context.Background()

	cmds := all()

	parser := argparse.NewArgParser("git-bundle-server <command> [<options>]")
	parser.SetIsTopLevel(true)
	for _, cmd := range cmds {
		parser.Subcommand(cmd)
	}
	parser.Parse(ctx, os.Args[1:])

	err := parser.InvokeSubcommand(ctx)
	if err != nil {
		log.Fatal("Failed with error: ", err)
	}
}
