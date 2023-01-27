package core_test

import (
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
}

func TestRepos_GetRepositories(t *testing.T) {
	testFileSystem := &MockFileSystem{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}

	for _, tt := range getRepositoriesTests {
		t.Run(tt.title, func(t *testing.T) {
			testFileSystem.On("UserHomeDir").Return("~", nil)
			testFileSystem.On("ReadFileLines",
				mock.AnythingOfType("string"),
			).Return(tt.readFileLines.First, tt.readFileLines.Second).Once()

			actual, err := core.GetRepositories(testUser, testFileSystem)

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
