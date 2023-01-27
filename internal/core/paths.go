package core

import (
	"os/user"
)

func bundleroot(user *user.User) string {
	return user.HomeDir + "/git-bundle-server/"
}

func webroot(user *user.User) string {
	return bundleroot(user) + "www/"
}

func reporoot(user *user.User) string {
	return bundleroot(user) + "git/"
}

func CrontabFile(user *user.User) string {
	return bundleroot(user) + "cron-schedule"
}
