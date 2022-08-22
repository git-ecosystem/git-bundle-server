package git

import (
	"log"
	"os/exec"
)

func GitCommand(args ...string) error {
	git, lookErr := exec.LookPath("git")

	if lookErr != nil {
		return lookErr
	}

	cmd := exec.Command(git, args...)
	err := cmd.Start()
	if err != nil {
		log.Fatal("Git command failed to start: ", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal("Git command returned a failure: ", err)
	}

	return err
}
