package core

import (
	"os/user"
	"path/filepath"
)

func bundleroot(user *user.User) string {
	return filepath.Join(user.HomeDir, "git-bundle-server")
}

func webroot(user *user.User) string {
	return filepath.Join(bundleroot(user), "www")
}

func reporoot(user *user.User) string {
	return filepath.Join(bundleroot(user), "git")
}

func CrontabFile(user *user.User) string {
	return filepath.Join(bundleroot(user), "cron-schedule")
}
