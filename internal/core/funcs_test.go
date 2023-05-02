package core_test

import (
	"fmt"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/core"
	"github.com/stretchr/testify/assert"
)

var urlToRouteTests = []struct {
	url           string
	expectedRoute string
	expectedMatch bool
}{
	// SSH tests
	{
		"git@example.com:git/git.git",
		"git/git",
		true,
	},
	{
		"some-user123@some.test.org:BUNDLE_SERVER.io/git-bundle-server",
		"BUNDLE_SERVER.io/git-bundle-server",
		true,
	},
	{
		"thing@some.site:imFineWith/trailingSlashes//",
		"imFineWith/trailingSlashes",
		true,
	},
	{
		"test@mydomain:deeper/toodeep/cannotmatch",
		"",
		false,
	},
	{
		"thing@some.site:imFineWith/trailingSlashes/",
		"imFineWith/trailingSlashes",
		true,
	},
	{
		"another-user@abc.def.ghi.jk:tooshallow",
		"",
		false,
	},

	// HTTP(S) tests
	{
		"hTTp://www.mysite.net/org/repo.git/",
		"org/repo",
		true,
	},
	{
		"https://domain.test/clone/me",
		"clone/me",
		true,
	},
	{
		"https://all.my.repos/having-some_fun/with.valid_ch4racters",
		"having-some_fun/with.valid_ch4racters",
		true,
	},
	{
		"http://completely.normal.site/with/invalid/repo",
		"",
		false,
	},
	{
		"HTTPS://SCREAM/INTOTHEVOID",
		"",
		false,
	},

	// Filesystem tests
	{
		"file:///root/path/to/a/repo.git",
		"a/repo",
		true,
	},
	{
		"FILE://RELATIVE/to/me",
		"to/me",
		true,
	},
	{
		"fIlE://spaces are allowed/in/path/to/repo",
		"to/repo",
		true,
	},
	{
		"fIlE:///butspaces/are/NOT/allowed in route",
		"",
		false,
	},
	{
		"file://somepathsaretooshort",
		"",
		false,
	},
}

func TestGetRouteFromUrl(t *testing.T) {
	for _, tt := range urlToRouteTests {
		var title string
		if tt.expectedMatch {
			title = fmt.Sprintf("%s => %s", tt.url, tt.expectedRoute)
		} else {
			title = fmt.Sprintf("%s (no match)", tt.url)
		}

		t.Run(title, func(t *testing.T) {
			route, isMatched := core.GetRouteFromUrl(tt.url)
			if tt.expectedMatch {
				assert.True(t, isMatched)
				assert.Equal(t, tt.expectedRoute, route)
			} else {
				assert.False(t, isMatched, "Expected no match, got route %s", route)
			}
		})
	}
}
