package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/log"
	"github.com/github/git-bundle-server/internal/utils"
)

type cronSchedule string

const (
	CronDaily  cronSchedule = "0 0 * * *"
	CronWeekly cronSchedule = "0 0 0 * *"
)

type CronScheduler interface {
	AddJob(ctx context.Context, schedule cronSchedule,
		exePath string, args []string) error
}

type cronScheduler struct {
	logger     log.TraceLogger
	user       common.UserProvider
	cmdExec    cmd.CommandExecutor
	fileSystem common.FileSystem
}

func NewCronScheduler(
	l log.TraceLogger,
	u common.UserProvider,
	c cmd.CommandExecutor,
	fs common.FileSystem,
) CronScheduler {
	return &cronScheduler{
		logger:     l,
		user:       u,
		cmdExec:    c,
		fileSystem: fs,
	}
}

func (c *cronScheduler) loadExistingSchedule(ctx context.Context) ([]byte, error) {
	buffer := bytes.Buffer{}
	exitCode, err := c.cmdExec.Run(ctx, "crontab", []string{"-l"}, cmd.Stdout(&buffer))
	if err != nil {
		return nil, c.logger.Error(ctx, err)
	} else if exitCode != 0 {
		return nil, c.logger.Errorf(ctx, "'crontab' exited with status %d", exitCode)
	}

	return buffer.Bytes(), nil
}

func (c *cronScheduler) commitCronSchedule(ctx context.Context, filename string) error {
	exitCode, err := c.cmdExec.RunQuiet(ctx, "crontab", filename)
	if err != nil {
		return c.logger.Error(ctx, err)
	} else if exitCode != 0 {
		return c.logger.Errorf(ctx, "'crontab' exited with status %d", exitCode)
	}

	return nil
}

func (c *cronScheduler) AddJob(ctx context.Context,
	schedule cronSchedule,
	exePath string,
	args []string,
) error {
	newSchedule := fmt.Sprintf("%s \"%s\" %s",
		schedule,
		exePath,
		utils.Map(args, func(s string) string { return "\"" + s + "\"" }),
	)

	scheduleBytes, err := c.loadExistingSchedule(ctx)
	if err != nil {
		return c.logger.Errorf(ctx, "failed to get existing cron schedule: %w", err)
	}

	scheduleStr := string(scheduleBytes)

	// TODO: Use comments to indicate a "region" where our schedule
	// is set, so we can remove the entire region even if we update
	// the schedule in the future.
	if strings.Contains(scheduleStr, newSchedule) {
		// We already have this schedule, so skip modifying
		// the crontab schedule.
		return nil
	}

	scheduleBytes = append(scheduleBytes, []byte(schedule)...)

	user, err := c.user.CurrentUser()
	if err != nil {
		return c.logger.Error(ctx, err)
	}
	scheduleFile := CrontabFile(user)

	err = c.fileSystem.WriteFile(scheduleFile, scheduleBytes)
	if err != nil {
		return c.logger.Errorf(ctx, "failed to write new cron schedule to temp file: %w", err)
	}

	err = c.commitCronSchedule(ctx, scheduleFile)
	if err != nil {
		return c.logger.Errorf(ctx, "failed to commit new cron schedule: %w", err)
	}

	_, err = c.fileSystem.DeleteFile(scheduleFile)
	if err != nil {
		return c.logger.Errorf(ctx, "failed to clear schedule temp file: %w", err)
	}

	return nil
}
