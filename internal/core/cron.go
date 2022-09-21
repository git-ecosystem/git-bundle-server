package core

import (
	"bytes"
	"fmt"
	"os/exec"
)

func GetCrontabCommand(args ...string) (*exec.Cmd, error) {
	crontab, err := exec.LookPath("crontab")
	if err != nil {
		return nil, fmt.Errorf("failed to find 'crontab' on the path: %w", err)
	}

	cmd := exec.Command(crontab, args...)
	return cmd, nil
}

func LoadExistingSchedule() ([]byte, error) {
	cmd, err := GetCrontabCommand("-l")
	if err != nil {
		return nil, err
	}

	buffer := bytes.Buffer{}
	cmd.Stdout = &buffer

	errorBuffer := bytes.Buffer{}
	cmd.Stderr = &errorBuffer

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("crontab failed to start: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("crontab returned a failure: %w\nstderr: %s", err, errorBuffer.String())
	}

	return buffer.Bytes(), nil
}

func CommitCronSchedule(filename string) error {
	cmd, err := GetCrontabCommand(filename)
	if err != nil {
		return err
	}

	errorBuffer := bytes.Buffer{}
	cmd.Stderr = &errorBuffer

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("crontab failed to start: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("crontab returned a failure: %w\nstderr: %s", err, errorBuffer.String())
	}

	return nil
}
