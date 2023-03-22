package core_test

import (
	"context"
	"errors"
	"os/user"
	"testing"

	"github.com/github/git-bundle-server/internal/core"
	. "github.com/github/git-bundle-server/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var getRepositoriesTests = []struct {
	title string

	// Expected values
	readFileLines Pair[[]string, error]

	// Expected output
	expectedRepos []core.Repository
	expectedErr   bool
}{
	{
		"empty file, empty list",
		NewPair[[]string, error]([]string{}, nil),
		[]core.Repository{},
		false,
	},
	{
		"error from filesystem",
		NewPair([]string{}, errors.New("error")),
		[]core.Repository{},
		true,
	},
	{
		"one repository",
		NewPair[[]string, error]([]string{
			"git/git",
		}, nil),
		[]core.Repository{
			{
				Route:   "git/git",
				RepoDir: "/my/test/dir/git-bundle-server/git/git/git",
				WebDir:  "/my/test/dir/git-bundle-server/www/git/git",
			},
		},
		false,
	},
	{
		"multiple repositories",
		NewPair[[]string, error]([]string{
			"git/git",
			"github/github",
			"org with spaces/repo with spaces",
			"", // Skips empty lines.
			"three/deep/repo",
		}, nil),
		[]core.Repository{
			{
				Route:   "git/git",
				RepoDir: "/my/test/dir/git-bundle-server/git/git/git",
				WebDir:  "/my/test/dir/git-bundle-server/www/git/git",
			},
			{
				Route:   "github/github",
				RepoDir: "/my/test/dir/git-bundle-server/git/github/github",
				WebDir:  "/my/test/dir/git-bundle-server/www/github/github",
			},
			{
				Route:   "org with spaces/repo with spaces",
				RepoDir: "/my/test/dir/git-bundle-server/git/org with spaces/repo with spaces",
				WebDir:  "/my/test/dir/git-bundle-server/www/org with spaces/repo with spaces",
			},
			{
				Route:   "three/deep/repo",
				RepoDir: "/my/test/dir/git-bundle-server/git/three/deep/repo",
				WebDir:  "/my/test/dir/git-bundle-server/www/three/deep/repo",
			},
		},
		false,
	},
}

func TestRepos_GetRepositories(t *testing.T) {
	testLogger := &MockTraceLogger{}
	testFileSystem := &MockFileSystem{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)
	repoProvider := core.NewRepositoryProvider(testLogger, testUserProvider, testFileSystem)

	for _, tt := range getRepositoriesTests {
		t.Run(tt.title, func(t *testing.T) {
			testFileSystem.On("ReadFileLines",
				mock.AnythingOfType("string"),
			).Return(tt.readFileLines.First, tt.readFileLines.Second).Once()

			actual, err := repoProvider.GetRepositories(context.Background())
			mock.AssertExpectationsForObjects(t, testUserProvider, testFileSystem)

			if tt.expectedErr {
				assert.NotNil(t, err, "Expected error")
				assert.Nil(t, actual, "Expected nil list")
			} else {
				assert.Nil(t, err, "Expected success")
				assert.NotNil(t, actual, "Expected non-nil list")
				assert.Equal(t, len(tt.expectedRepos), len(actual), "Length mismatch")
				for _, repo := range tt.expectedRepos {
					a := actual[repo.Route]

					assert.Equal(t, repo.Route, a.Route)
					assert.Equal(t, repo.RepoDir, a.RepoDir)
					assert.Equal(t, repo.WebDir, a.WebDir)
				}
			}
		})
	}
}
