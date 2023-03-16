package core

import (
	"regexp"
	"strings"
)

// Standalone helper functions for core (repo, cron, etc.) functionality.

func GetRouteFromUrl(url string) (string, bool) {
	matchers := []*regexp.Regexp{
		// SSH, matches <username>@<domain>:<route>[.git]
		regexp.MustCompile(`^[\w-]+@[\w\.-]+:([\w\.-]+/[\w\.-]+)/*$`),

		// HTTP(S), matches http[s]://<domain>/<route>[.git]
		regexp.MustCompile(`^(?i:http[s]?)://[\w\.-]+/([\w\.-]+/[\w\.-]+)/*$`),

		// Filesystem, matches file://[<path>/]<route>[.git]
		regexp.MustCompile(`^(?i:file)://[\w\.-/ ]*/([\w\.-]+/[\w\.-]+)/*$`),
	}

	for _, matcher := range matchers {
		if groups := matcher.FindStringSubmatch(url); groups != nil {
			return strings.TrimSuffix(groups[1], ".git"), true
		}
	}

	return "", false
}
