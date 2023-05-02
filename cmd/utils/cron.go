package utils

import (
	"context"

	"github.com/git-ecosystem/git-bundle-server/internal/common"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/git-ecosystem/git-bundle-server/internal/log"
)

type CronHelper interface {
	SetCronSchedule(ctx context.Context) error
}

type cronHelper struct {
	logger     log.TraceLogger
	fileSystem common.FileSystem
	scheduler  core.CronScheduler
}

func NewCronHelper(
	l log.TraceLogger,
	fs common.FileSystem,
	s core.CronScheduler,
) CronHelper {
	return &cronHelper{
		logger:     l,
		fileSystem: fs,
		scheduler:  s,
	}
}

func (c *cronHelper) SetCronSchedule(ctx context.Context) error {
	pathToExec, err := c.fileSystem.GetLocalExecutable("git-bundle-server")
	if err != nil {
		return c.logger.Errorf(ctx, "failed to get executable: %w", err)
	}

	err = c.scheduler.AddJob(ctx, core.CronDaily, pathToExec, []string{"update-all"})
	if err != nil {
		return c.logger.Errorf(ctx, "failed to set cron schedule: %w", err)
	}

	return nil
}
