package main

import (
	"log"
	"os"

	"github.com/github/git-bundle-server/internal/argparse"
)

func all() []argparse.Subcommand {
	return []argparse.Subcommand{
		Delete{},
		Init{},
		Start{},
		Stop{},
		Update{},
		UpdateAll{},
	}
}

func main() {
	cmds := all()

	parser := argparse.NewArgParser("git-bundle-server <command> [<options>]")
	parser.SetIsTopLevel(true)
	for _, cmd := range cmds {
		parser.Subcommand(cmd)
	}
	parser.Parse(os.Args[1:])

	err := parser.InvokeSubcommand()
	if err != nil {
		log.Fatal("Failed with error: ", err)
	}
}
