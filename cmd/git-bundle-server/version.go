package main

import (
	"context"
	"fmt"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
)

type versionCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewVersionCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &versionCmd{
		logger:    logger,
		container: container,
	}
}

func (versionCmd) Name() string {
	return "version"
}

func (versionCmd) Description() string {
	return `
Display the version information for the bundle server CLI.`
}

func (v *versionCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(v.logger, "git-bundle-server version")
	parser.Parse(ctx, args)

	versionStr := utils.Version
	if versionStr == "" {
		versionStr = "<no version>"
	}

	fmt.Printf("git-bundle-server version %s\n", versionStr)

	return nil
}
