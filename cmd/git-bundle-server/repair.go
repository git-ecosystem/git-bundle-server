package main

import (
	"context"
	"fmt"

	"github.com/git-ecosystem/git-bundle-server/cmd/utils"
	"github.com/git-ecosystem/git-bundle-server/internal/argparse"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
	typeutils "github.com/git-ecosystem/git-bundle-server/internal/utils"
)

type repairCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewRepairCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &repairCmd{
		logger:    logger,
		container: container,
	}
}

func (repairCmd) Name() string {
	return "repair"
}

func (repairCmd) Description() string {
	return `
Scan and correct inconsistencies in the bundle server's internal registries and
storage.`
}

func (r *repairCmd) repairRoutes(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(r.logger, "git-bundle-server repair routes [--start-all] [--dry-run]")
	enable := parser.Bool("start-all", false, "turn on bundle computation for all repositories found")
	dryRun := parser.Bool("dry-run", false, "report the repairs needed, but do not perform them")
	// TODO: add a '--cleanup' option to delete non-repo contents inside repo root
	parser.Parse(ctx, args)

	repoProvider := utils.GetDependency[core.RepositoryProvider](ctx, r.container)

	// Read the routes file
	repos, err := repoProvider.GetRepositories(ctx)
	if err != nil {
		// If routes file cannot be read, start over
		fmt.Println("warning: cannot load routes file; rebuilding from scratch...")
		repos = make(map[string]core.Repository)
	}

	// Read the repositories as represented by internal storage
	storedRepos, err := repoProvider.ReadRepositoryStorage(ctx)
	if err != nil {
		return r.logger.Errorf(ctx, "could not read internal repository storage: %w", err)
	}

	_, missingOnDisk, notRegistered := typeutils.SegmentKeys(repos, storedRepos)

	// Print the updates to be made
	fmt.Print("\n")

	if *enable && len(notRegistered) > 0 {
		fmt.Println("Unregistered routes to add")
		fmt.Println("--------------------------")
		for _, route := range notRegistered {
			fmt.Printf("* %s\n", route)
			repos[route] = storedRepos[route]
		}
		fmt.Print("\n")
	}

	if len(missingOnDisk) > 0 {
		fmt.Println("Missing or invalid routes to remove")
		fmt.Println("-----------------------------------")
		for _, route := range missingOnDisk {
			fmt.Printf("* %s\n", route)
			delete(repos, route)
		}
		fmt.Print("\n")
	}

	if (!*enable || len(notRegistered) == 0) && len(missingOnDisk) == 0 {
		fmt.Println("No repairs needed.")
		return nil
	}

	if *dryRun {
		fmt.Println("Skipping updates (dry run)")
	} else {
		fmt.Println("Applying route repairs...")
		err := repoProvider.WriteAllRoutes(ctx, repos)
		if err != nil {
			return err
		}

		// Start the global cron schedule (if it's not already running)
		cron := utils.GetDependency[utils.CronHelper](ctx, r.container)
		cron.SetCronSchedule(ctx)
		fmt.Println("Done")
	}

	return nil
}

func (r *repairCmd) Run(ctx context.Context, args []string) error {
	parser := argparse.NewArgParser(r.logger, "git-bundle-server repair <subcommand> [<options>]")
	parser.Subcommand(argparse.NewSubcommand("routes", "Correct the contents of the internal route registry", r.repairRoutes))
	parser.Parse(ctx, args)

	return parser.InvokeSubcommand(ctx)
}
