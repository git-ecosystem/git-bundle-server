package core_test

import (
	"context"
	"errors"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/common"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	. "github.com/git-ecosystem/git-bundle-server/internal/testhelpers"
	"github.com/git-ecosystem/git-bundle-server/internal/utils"
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
	repoProvider := core.NewRepositoryProvider(testLogger, testUserProvider, testFileSystem, nil)

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
					assert.Equal(t, filepath.Clean(repo.RepoDir), a.RepoDir)
					assert.Equal(t, filepath.Clean(repo.WebDir), a.WebDir)
				}
			}
		})
	}
}

var readRepositoryStorageTests = []struct {
	title string

	foundPaths            Pair[[]Pair[string, bool], error] // list of (path, isDir), error
	foundRouteIsValidRepo map[string]bool                   // map of route -> whether GetRemoteUrl succeeds

	expectedRoutes []string
	expectedErr    bool
}{
	{
		"no dirs found",
		NewPair([]Pair[string, bool]{}, error(nil)),
		map[string]bool{},
		[]string{},
		false,
	},
	{
		"multiple valid repos found",
		NewPair(
			[]Pair[string, bool]{
				NewPair("my/repo", true),
				NewPair("another/route", true),
			},
			error(nil),
		),
		map[string]bool{
			"my/repo":       true,
			"another/route": true,
		},
		[]string{"my/repo", "another/route"},
		false,
	},
	{
		"ignores non-directories",
		NewPair(
			[]Pair[string, bool]{
				NewPair("this-is-a/file", false),
				NewPair("my/repo", true),
			},
			error(nil),
		),
		map[string]bool{
			"my/repo": true,
		},
		[]string{"my/repo"},
		false,
	},
	{
		"ignores invalid Git repos",
		NewPair(
			[]Pair[string, bool]{
				NewPair("is/a-repo", true),
				NewPair("not/a-repo", true),
			},
			error(nil),
		),
		map[string]bool{
			"is/a-repo":  true,
			"not/a-repo": false,
		},
		[]string{"is/a-repo"},
		false,
	},
}

func TestRepos_ReadRepositoryStorage(t *testing.T) {
	testLogger := &MockTraceLogger{}
	testFileSystem := &MockFileSystem{}
	testGitHelper := &MockGitHelper{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)
	repoProvider := core.NewRepositoryProvider(testLogger, testUserProvider, testFileSystem, testGitHelper)

	for _, tt := range readRepositoryStorageTests {
		t.Run(tt.title, func(t *testing.T) {
			testFileSystem.On("ReadDirRecursive",
				filepath.Clean("/my/test/dir/git-bundle-server/git"),
				2,
				true,
			).Return(utils.Map(tt.foundPaths.First, func(path Pair[string, bool]) common.ReadDirEntry {
				return TestReadDirEntry{
					PathVal:  filepath.Join("/my/test/dir/git-bundle-server/git", path.First),
					IsDirVal: path.Second,
				}
			}), tt.foundPaths.Second).Once()

			for route, isValid := range tt.foundRouteIsValidRepo {
				call := testGitHelper.On("GetRemoteUrl",
					mock.Anything,
					filepath.Join("/my/test/dir/git-bundle-server/git", route),
				).Once()

				if isValid {
					call.Return("https://localhost/example-remote", nil)
				} else {
					call.Return("", errors.New("could not get remote URL"))
				}
			}

			actual, err := repoProvider.ReadRepositoryStorage(context.Background())
			mock.AssertExpectationsForObjects(t, testUserProvider, testFileSystem, testGitHelper)

			if tt.expectedErr {
				assert.NotNil(t, err, "Expected error")
				assert.Nil(t, actual, "Expected nil map")
			} else {
				assert.Nil(t, err, "Expected success")
				assert.NotNil(t, actual, "Expected non-nil map")
				assert.NotNil(t, actual, "Expected non-nil list")
				assert.Equal(t, len(tt.expectedRoutes), len(actual), "Length mismatch")
				for _, route := range tt.expectedRoutes {
					a := actual[route]

					assert.Equal(t, route, a.Route)
					assert.Equal(t, filepath.Join("/my/test/dir/git-bundle-server/git", route), a.RepoDir)
					assert.Equal(t, filepath.Join("/my/test/dir/git-bundle-server/www", route), a.WebDir)
				}
			}

			// Reset mocks
			testFileSystem.Mock = mock.Mock{}
			testGitHelper.Mock = mock.Mock{}
		})
	}
}

var writeAllRoutesTests = []struct {
	title        string
	repos        map[string]core.Repository
	expectedFile []string
}{
	{
		"empty repo map",
		map[string]core.Repository{},
		[]string{""},
	},
	{
		"single repo",
		map[string]core.Repository{
			"test/route": {Route: "test/route"},
		},
		[]string{
			"test/route",
		},
	},
	{
		"multiple repos",
		map[string]core.Repository{
			"test/route":   {Route: "test/route"},
			"another/repo": {Route: "another/repo"},
		},
		[]string{
			"test/route",
			"another/repo",
		},
	},
}

func TestRepos_WriteAllRoutes(t *testing.T) {
	testLogger := &MockTraceLogger{}
	testFileSystem := &MockFileSystem{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)
	repoProvider := core.NewRepositoryProvider(testLogger, testUserProvider, testFileSystem, nil)

	for _, tt := range writeAllRoutesTests {
		t.Run(tt.title, func(t *testing.T) {
			var actualFilename string
			var actualFileBytes []byte

			testFileSystem.On("WriteFile",
				mock.MatchedBy(func(filename string) bool {
					actualFilename = filename
					return true
				}),
				mock.MatchedBy(func(fileBytes any) bool {
					// Save off value and always match
					actualFileBytes = fileBytes.([]byte)
					return true
				}),
			).Return(nil).Once()

			err := repoProvider.WriteAllRoutes(context.Background(), tt.repos)
			assert.Nil(t, err)
			mock.AssertExpectationsForObjects(t, testUserProvider, testFileSystem)

			// Check filename
			expectedFilename := filepath.Clean("/my/test/dir/git-bundle-server/routes")
			assert.Equal(t, expectedFilename, actualFilename)

			// Check routes file contents
			fileLines := strings.Split(strings.TrimSpace(string(actualFileBytes)), "\n")

			assert.ElementsMatch(t, tt.expectedFile, fileLines)

			// Reset mocks
			testFileSystem.Mock = mock.Mock{}
		})
	}
}
