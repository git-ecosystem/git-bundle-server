package daemon_test

import (
	"encoding/xml"
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/github/git-bundle-server/internal/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var launchdCreateTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig
	force  boolArg

	// Mocked responses
	fileExists         []pair[bool, error]
	writeFile          []error
	launchctlPrint     []pair[int, error]
	launchctlBootstrap []pair[int, error]
	launchctlBootout   []pair[int, error]

	// Expected values
	expectErr bool
}{
	{
		"Fresh config created if none exists",
		&basicDaemonConfig,
		Any,
		[]pair[bool, error]{newPair[bool, error](false, nil)}, // file exists
		[]error{nil}, // write file
		[]pair[int, error]{newPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{newPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"Config exists & is not bootstrapped doesn't write file, bootstraps",
		&basicDaemonConfig,
		False,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{}, // write file
		[]pair[int, error]{newPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{newPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"'force' option overwrites file and bootstraps when not already bootstrapped",
		&basicDaemonConfig,
		True,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]pair[int, error]{newPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{newPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"Config exists & already bootstrapped does nothing",
		&basicDaemonConfig,
		False,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{}, // write file
		[]pair[int, error]{newPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{},                            // launchctl bootstrap
		[]pair[int, error]{},                            // launchctl bootout
		false,
	},
	{
		"'force' option unloads config, overwrites file, and bootstraps",
		&basicDaemonConfig,
		True,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]pair[int, error]{newPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{newPair[int, error](0, nil)}, // launchctl bootstrap
		[]pair[int, error]{newPair[int, error](0, nil)}, // launchctl bootout
		false,
	},
	{
		"Config missing & already bootstrapped throws error",
		&basicDaemonConfig,
		Any,
		[]pair[bool, error]{newPair[bool, error](false, nil)}, // file exists
		[]error{}, // write file
		[]pair[int, error]{newPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]pair[int, error]{},                            // launchctl bootstrap
		[]pair[int, error]{},                            // launchctl bootout
		true,
	},
}

func TestLaunchd_Create(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &mockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &mockCommandExecutor{}

	testFileSystem := &mockFileSystem{}

	launchd := daemon.NewLaunchdProvider(testUserProvider, testCommandExecutor, testFileSystem)

	// Verify launchd commands called
	for _, tt := range launchdCreateTests {
		forceArg := tt.force.toBoolList()
		for _, force := range forceArg {
			t.Run(fmt.Sprintf("%s (force='%t')", tt.title, force), func(t *testing.T) {
				// Mock responses
				for _, retVal := range tt.launchctlPrint {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "print" }),
					).Return(retVal.first, retVal.second).Once()
				}
				for _, retVal := range tt.launchctlBootstrap {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "bootstrap" }),
					).Return(retVal.first, retVal.second).Once()
				}
				for _, retVal := range tt.launchctlBootout {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "bootout" }),
					).Return(retVal.first, retVal.second).Once()
				}
				for _, retVal := range tt.fileExists {
					testFileSystem.On("FileExists",
						mock.AnythingOfType("string"),
					).Return(retVal.first, retVal.second).Once()
				}
				for _, retVal := range tt.writeFile {
					testFileSystem.On("WriteFile",
						mock.AnythingOfType("string"),
						mock.Anything,
					).Return(retVal).Once()
				}

				// Run "Create"
				err := launchd.Create(tt.config, force)

				// Assert on expected values
				if tt.expectErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
				mock.AssertExpectationsForObjects(t, testCommandExecutor, testFileSystem)

				// Reset mocks
				testCommandExecutor.Mock = mock.Mock{}
				testFileSystem.Mock = mock.Mock{}
			})
		}
	}

	// Verify content of created file
	t.Run("Created file content and path are correct", func(t *testing.T) {
		var actualFilename string
		var actualFileBytes []byte

		// Mock responses for successful fresh write
		testCommandExecutor.On("Run",
			"launchctl",
			mock.MatchedBy(func(args []string) bool { return args[0] == "print" }),
		).Return(daemon.LaunchdServiceNotFoundErrorCode, nil).Once()
		testCommandExecutor.On("Run",
			"launchctl",
			mock.MatchedBy(func(args []string) bool { return args[0] == "bootstrap" }),
		).Return(0, nil).Once()
		testFileSystem.On("FileExists",
			mock.AnythingOfType("string"),
		).Return(false, nil).Once()

		// Use mock to save off input args
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

		err := launchd.Create(&basicDaemonConfig, false)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor, testFileSystem)

		// Check filename
		expectedFilename := filepath.Clean(fmt.Sprintf("/my/test/dir/Library/LaunchAgents/%s.plist", basicDaemonConfig.Label))
		assert.Equal(t, expectedFilename, actualFilename)

		// Check file contents
		err = xml.Unmarshal(actualFileBytes, new(interface{}))
		if err != nil {
			assert.Fail(t, "plist XML is malformed")
		}
		fileContents := string(actualFileBytes)
		assert.Contains(t, fileContents, fmt.Sprintf("<key>Label</key><string>%s</string>", basicDaemonConfig.Label))
		assert.Contains(t, fileContents, fmt.Sprintf("<key>Program</key><string>%s</string>", basicDaemonConfig.Program))
	})
}
