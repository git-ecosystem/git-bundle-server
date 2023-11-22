package core

import (
	"fmt"
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

func ParseRoute(route string, repoOnly bool) (string, string, string, error) {
	elements := strings.FieldsFunc(route, func(char rune) bool { return char == '/' })
	validElementPattern := regexp.MustCompile(`^[\w\.-]+$`)
	for _, e := range elements {
		if !validElementPattern.MatchString(e) {
			return "", "", "",
				fmt.Errorf("invalid element '%s'; route may only contain alphanumeric characters, '.', '_', and/or '-'", e)
		}
		if e == "." || e == ".." {
			return "", "", "", fmt.Errorf("invalid route element '%s'", e)
		}
	}

	switch len(elements) {
	case 0:
		return "", "", "", fmt.Errorf("empty route")
	case 1:
		return "", "", "", fmt.Errorf("route has owner, but no repo")
	case 2:
		return elements[0], elements[1], "", nil
	case 3:
		if repoOnly {
			return "", "", "", fmt.Errorf("route is too deep")
		}
		return elements[0], elements[1], elements[2], nil
	default:
		return "", "", "", fmt.Errorf("route is too deep")
	}
}
