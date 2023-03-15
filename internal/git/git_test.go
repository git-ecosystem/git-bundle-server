package git_test

import (
	"context"
	"io"
	"testing"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/git"
	. "github.com/github/git-bundle-server/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var createIncrementalBundleTests = []struct {
	title string

	// Inputs
	repoDir  string
	filename string
	prereqs  []string

	// Mocked responses
	bundleCreate       Pair[int, error]
	bundleCreateStderr string

	// Expected values
	expectedBundleCreated bool
	expectErr             bool
}{
	{
		"Successful bundle creation",

		"/test/home/git-bundle-server/git/test/myrepo/",
		"/test/home/git-bundle-server/www/test/myrepo/bundle-1234.bundle",
		[]string{"^018d4b8a"},

		NewPair[int, error](0, nil),
		"",

		true,
		false,
	},
	{
		"Successful no-op (empty bundle)",

		"/test/home/git-bundle-server/git/test/myrepo/",
		"/test/home/git-bundle-server/www/test/myrepo/bundle-5678.bundle",
		[]string{"^0793b0ce", "^3649daa0"},

		NewPair[int, error](128, nil),
		"fatal: Refusing to create empty bundle",

		false,
		false,
	},
}

func TestGit_CreateIncrementalBundle(t *testing.T) {
	// Set up mocks
	testLogger := &MockTraceLogger{}
	testCommandExecutor := &MockCommandExecutor{}

	gitHelper := git.NewGitHelper(testLogger, testCommandExecutor)

	for _, tt := range createIncrementalBundleTests {
		t.Run(tt.title, func(t *testing.T) {
			var stdin io.Reader
			var stdout io.Writer

			// Mock responses
			testCommandExecutor.On("Run",
				mock.Anything,
				"git",
				[]string{"-C", tt.repoDir, "bundle", "create", tt.filename, "--stdin", "--branches"},
				mock.MatchedBy(func(settings []cmd.Setting) bool {
					var ok bool
					stdin = nil
					stdout = nil
					for _, setting := range settings {
						switch setting.Key {
						case cmd.StdinKey:
							stdin, ok = setting.Value.(io.Reader)
							if !ok {
								return false
							}
						case cmd.StderrKey:
							stdout, ok = setting.Value.(io.Writer)
							if !ok {
								return false
							}
						}
					}
					return stdin != nil && stdout != nil
				}),
			).Run(func(mock.Arguments) {
				stdout.Write([]byte(tt.bundleCreateStderr))
			}).Return(tt.bundleCreate.First, tt.bundleCreate.Second)

			// Run 'CreateIncrementalBundle()'
			actualBundleCreated, err := gitHelper.CreateIncrementalBundle(context.Background(), tt.repoDir, tt.filename, tt.prereqs)

			// Assert on expected values
			assert.Equal(t, tt.expectedBundleCreated, actualBundleCreated)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, testCommandExecutor)

			// Check the content of stdin
			expectedStdin := ConcatLines(tt.prereqs)
			expectedStdinLen := len(expectedStdin)
			stdinBytes := make([]byte, expectedStdinLen+1)
			numRead, err := stdin.Read(stdinBytes)
			assert.NoError(t, err)
			assert.Equal(t, expectedStdinLen, numRead)
			assert.Equal(t, expectedStdin, string(stdinBytes[:expectedStdinLen]))

			// Reset mocks
			testCommandExecutor.Mock = mock.Mock{}
		})
	}
}
