package daemon_test

import (
	"encoding/xml"
	"fmt"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/github/git-bundle-server/internal/daemon"
	. "github.com/github/git-bundle-server/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var launchdCreateBehaviorTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig
	force  BoolArg

	// Mocked responses
	fileExists         []Pair[bool, error]
	writeFile          []error
	launchctlPrint     []Pair[int, error]
	launchctlBootstrap []Pair[int, error]
	launchctlBootout   []Pair[int, error]

	// Expected values
	expectErr bool
}{
	{
		"Fresh config created if none exists",
		&basicDaemonConfig,
		Any,
		[]Pair[bool, error]{NewPair[bool, error](false, nil)}, // file exists
		[]error{nil}, // write file
		[]Pair[int, error]{NewPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{NewPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]Pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"Config exists & is not bootstrapped doesn't write file, bootstraps",
		&basicDaemonConfig,
		False,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{}, // write file
		[]Pair[int, error]{NewPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{NewPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]Pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"'force' option overwrites file and bootstraps when not already bootstrapped",
		&basicDaemonConfig,
		True,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]Pair[int, error]{NewPair[int, error](daemon.LaunchdServiceNotFoundErrorCode, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{NewPair[int, error](0, nil)},                                      // launchctl bootstrap
		[]Pair[int, error]{}, // launchctl bootout
		false,
	},
	{
		"Config exists & already bootstrapped does nothing",
		&basicDaemonConfig,
		False,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{}, // write file
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{},                            // launchctl bootstrap
		[]Pair[int, error]{},                            // launchctl bootout
		false,
	},
	{
		"'force' option unloads config, overwrites file, and bootstraps",
		&basicDaemonConfig,
		True,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // launchctl bootstrap
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // launchctl bootout
		false,
	},
	{
		"Config missing & already bootstrapped throws error",
		&basicDaemonConfig,
		Any,
		[]Pair[bool, error]{NewPair[bool, error](false, nil)}, // file exists
		[]error{}, // write file
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // launchctl print (isBootstrapped)
		[]Pair[int, error]{},                            // launchctl bootstrap
		[]Pair[int, error]{},                            // launchctl bootout
		true,
	},
}

var launchdCreatePlistTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig

	// Expected values
	expectedPlistLines []string
}{
	{
		title:  "Created plist is correct",
		config: &basicDaemonConfig,
		expectedPlistLines: []string{
			`<?xml version="1.0" encoding="UTF-8"?>`,
			`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
			`<plist version="1.0">`,
			"<dict>",

			"<key>Label</key>",
			fmt.Sprintf("<string>%s</string>", basicDaemonConfig.Label),

			"<key>Program</key>",
			fmt.Sprintf("<string>%s</string>", basicDaemonConfig.Program),

			"<key>LimitLoadToSessionType</key>",
			"<string>Background</string>",

			"<key>StandardOutPath</key>",
			"<string>/dev/null</string>",

			"<key>StandardErrorPath</key>",
			"<string>/dev/null</string>",

			"<key>ProgramArguments</key>",
			"<array>",
			fmt.Sprintf("<string>%s</string>", basicDaemonConfig.Program),
			"</array>",

			"</dict>",
			"</plist>",
		},
	},
	{
		title: "Plist contents are escaped",
		config: &daemon.DaemonConfig{
			// All of <'&"\t> should be replaced by the associated escape code
			// 🤔 is in-range for XML (no replacement), but ￿ (\uFFFF) is
			// out-of-range and replaced with � (\uFFFD)
			// See https://www.w3.org/TR/xml11/Overview.html#charsets for details
			Label:   "test-escape<'&\"	🤔￿>",
			Program: "/path/to/the/program with a space",
		},
		expectedPlistLines: []string{
			`<?xml version="1.0" encoding="UTF-8"?>`,
			`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
			`<plist version="1.0">`,
			"<dict>",

			"<key>Label</key>",
			"<string>test-escape&lt;&#39;&amp;&#34;&#x9;🤔�&gt;</string>",

			"<key>Program</key>",
			"<string>/path/to/the/program with a space</string>",

			"<key>LimitLoadToSessionType</key>",
			"<string>Background</string>",

			"<key>StandardOutPath</key>",
			"<string>/dev/null</string>",

			"<key>StandardErrorPath</key>",
			"<string>/dev/null</string>",

			"<key>ProgramArguments</key>",
			"<array>",
			"<string>/path/to/the/program with a space</string>",
			"</array>",

			"</dict>",
			"</plist>",
		},
	},
	{
		title: "Created plist captures args",
		config: &daemon.DaemonConfig{
			Label:     "test-with-args",
			Program:   "/path/to/the/program",
			Arguments: []string{"--test", "another-arg"},
		},
		expectedPlistLines: []string{
			`<?xml version="1.0" encoding="UTF-8"?>`,
			`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
			`<plist version="1.0">`,
			"<dict>",

			"<key>Label</key>",
			"<string>test-with-args</string>",

			"<key>Program</key>",
			"<string>/path/to/the/program</string>",

			"<key>LimitLoadToSessionType</key>",
			"<string>Background</string>",

			"<key>StandardOutPath</key>",
			"<string>/dev/null</string>",

			"<key>StandardErrorPath</key>",
			"<string>/dev/null</string>",

			"<key>ProgramArguments</key>",
			"<array>",
			"<string>/path/to/the/program</string>",
			"<string>--test</string>",
			"<string>another-arg</string>",
			"</array>",

			"</dict>",
			"</plist>",
		},
	},
}

func TestLaunchd_Create(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	testFileSystem := &MockFileSystem{}

	launchd := daemon.NewLaunchdProvider(testUserProvider, testCommandExecutor, testFileSystem)

	// Verify launchd commands called
	for _, tt := range launchdCreateBehaviorTests {
		forceArg := tt.force.ToBoolList()
		for _, force := range forceArg {
			t.Run(fmt.Sprintf("%s (force='%t')", tt.title, force), func(t *testing.T) {
				// Mock responses
				for _, retVal := range tt.launchctlPrint {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "print" }),
					).Return(retVal.First, retVal.Second).Once()
				}
				for _, retVal := range tt.launchctlBootstrap {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "bootstrap" }),
					).Return(retVal.First, retVal.Second).Once()
				}
				for _, retVal := range tt.launchctlBootout {
					testCommandExecutor.On("Run",
						"launchctl",
						mock.MatchedBy(func(args []string) bool { return args[0] == "bootout" }),
					).Return(retVal.First, retVal.Second).Once()
				}
				for _, retVal := range tt.fileExists {
					testFileSystem.On("FileExists",
						mock.AnythingOfType("string"),
					).Return(retVal.First, retVal.Second).Once()
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
	for _, tt := range launchdCreatePlistTests {
		t.Run(tt.title, func(t *testing.T) {
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

			err := launchd.Create(tt.config, false)
			assert.Nil(t, err)
			mock.AssertExpectationsForObjects(t, testCommandExecutor, testFileSystem)

			// Check filename
			expectedFilename := filepath.Clean(fmt.Sprintf("/my/test/dir/Library/LaunchAgents/%s.plist", tt.config.Label))
			assert.Equal(t, expectedFilename, actualFilename)

			// Check XML
			err = xml.Unmarshal(actualFileBytes, new(interface{}))
			if err != nil {
				assert.Fail(t, "plist XML is malformed")
			}
			fileContents := strings.TrimSpace(string(actualFileBytes))
			plistLines := strings.Split(
				regexp.MustCompile(`>\s*<`).ReplaceAllString(fileContents, ">\n<"), "\n")

			assert.ElementsMatch(t, tt.expectedPlistLines, plistLines)

			// Reset mocks
			testCommandExecutor.Mock = mock.Mock{}
			testFileSystem.Mock = mock.Mock{}
		})
	}
}

func TestLaunchd_Start(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	launchd := daemon.NewLaunchdProvider(testUserProvider, testCommandExecutor, nil)

	// Test #1: launchctl succeeds
	t.Run("Calls correct launchctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			"launchctl",
			[]string{"kickstart", fmt.Sprintf("user/123/%s", basicDaemonConfig.Label)},
		).Return(0, nil).Once()

		err := launchd.Start(basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: launchctl fails
	t.Run("Returns error when launchctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := launchd.Start(basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}

func TestLaunchd_Stop(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	launchd := daemon.NewLaunchdProvider(testUserProvider, testCommandExecutor, nil)

	// Test #1: launchctl succeeds
	t.Run("Calls correct launchctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			"launchctl",
			[]string{"kill", "SIGINT", fmt.Sprintf("user/123/%s", basicDaemonConfig.Label)},
		).Return(0, nil).Once()

		err := launchd.Stop(basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: launchctl fails with uncaught error
	t.Run("Returns error when launchctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := launchd.Stop(basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #3: launchctl fails with uncaught error
	t.Run("Exits without error if service not found", func(t *testing.T) {
		testCommandExecutor.On("Run",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(daemon.LaunchdServiceNotFoundErrorCode, nil).Once()

		err := launchd.Stop(basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}
