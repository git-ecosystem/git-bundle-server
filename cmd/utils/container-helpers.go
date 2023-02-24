package utils

import (
	"context"

	"github.com/github/git-bundle-server/internal/bundles"
	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/daemon"
	"github.com/github/git-bundle-server/internal/log"
)

func BuildGitBundleServerContainer(logger log.TraceLogger) *DependencyContainer {
	container := NewDependencyContainer()
	registerDependency(container, func(ctx context.Context) common.UserProvider {
		return common.NewUserProvider()
	})
	registerDependency(container, func(ctx context.Context) cmd.CommandExecutor {
		return cmd.NewCommandExecutor(logger)
	})
	registerDependency(container, func(ctx context.Context) common.FileSystem {
		return common.NewFileSystem()
	})
	registerDependency(container, func(ctx context.Context) core.RepositoryProvider {
		return core.NewRepositoryProvider(
			logger,
			GetDependency[common.UserProvider](ctx, container),
			GetDependency[common.FileSystem](ctx, container),
		)
	})
	registerDependency(container, func(ctx context.Context) bundles.BundleProvider {
		return bundles.NewBundleProvider(logger)
	})
	registerDependency(container, func(ctx context.Context) core.CronScheduler {
		return core.NewCronScheduler(
			logger,
			GetDependency[common.UserProvider](ctx, container),
			GetDependency[cmd.CommandExecutor](ctx, container),
			GetDependency[common.FileSystem](ctx, container),
		)
	})
	registerDependency(container, func(ctx context.Context) CronHelper {
		return NewCronHelper(
			logger,
			GetDependency[common.FileSystem](ctx, container),
			GetDependency[core.CronScheduler](ctx, container),
		)
	})
	registerDependency(container, func(ctx context.Context) daemon.DaemonProvider {
		t, err := daemon.NewDaemonProvider(
			logger,
			GetDependency[common.UserProvider](ctx, container),
			GetDependency[cmd.CommandExecutor](ctx, container),
			GetDependency[common.FileSystem](ctx, container),
		)
		if err != nil {
			logger.Fatal(ctx, err)
		}
		return t
	})

	return container
}
