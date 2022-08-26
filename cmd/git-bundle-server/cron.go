package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/git-bundle-server/internal/core"
)

func SetCronSchedule() error {
	pathToExec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %w", err)
	}

	dailySchedule := "0 0 * * * \"" + pathToExec + "\" update-all\n"

	scheduleBytes, err := core.LoadExistingSchedule()
	if err != nil {
		return fmt.Errorf("failed to get existing cron schedule: %w", err)
	}

	scheduleStr := string(scheduleBytes)

	// TODO: Use comments to indicate a "region" where our schedule
	// is set, so we can remove the entire region even if we update
	// the schedule in the future.
	if strings.Contains(scheduleStr, dailySchedule) {
		// We already have this schedule, so skip modifying
		// the crontab schedule.
		return nil
	}

	scheduleBytes = append(scheduleBytes, []byte(dailySchedule)...)
	scheduleFile := core.CrontabFile()

	err = os.WriteFile(scheduleFile, scheduleBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write new cron schedule to temp file: %w", err)
	}

	err = core.CommitCronSchedule(scheduleFile)
	if err != nil {
		return fmt.Errorf("failed to commit new cron schedule: %w", err)
	}

	err = os.Remove(scheduleFile)
	if err != nil {
		return fmt.Errorf("failed to clear schedule file: %w", err)
	}

	return nil
}
