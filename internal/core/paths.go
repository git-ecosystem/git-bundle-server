package core

import "os"

func bundleroot() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		// TODO: respond better. For now, try creating in "/var"
		dirname = "/var"
	}

	return dirname + "/git-bundle-server/"
}

func webroot() string {
	return bundleroot() + "www/"
}

func reporoot() string {
	return bundleroot() + "git/"
}

func CrontabFile() string {
	return bundleroot() + "cron-schedule"
}
