package daemon_test

import (
	"context"
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

var systemdCreateBehaviorTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig
	force  BoolArg

	// Mocked responses (ordered per list!)
	fileExists            []Pair[bool, error]
	writeFile             []error
	systemctlDaemonReload []Pair[int, error]

	// Expected values
	expectErr bool
}{
	{
		"Fresh service unit created if none exists, daemon reloaded",
		&daemon.DaemonConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/local/bin/test/git-bundle-web-server",
		},
		Any,
		[]Pair[bool, error]{NewPair[bool, error](false, nil)}, // file exists
		[]error{nil}, // write file
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // systemctl daemon-reload
		false,
	},
	{
		"Service unit exists, doesn't write file or reload",
		&daemon.DaemonConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/local/bin/test/git-bundle-web-server",
		},
		False,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{},            // write file
		[]Pair[int, error]{}, // systemctl daemon-reload
		false,
	},
	{
		"'force' option overwrites service unit and reloads daemon",
		&daemon.DaemonConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/local/bin/test/git-bundle-web-server",
		},
		True,
		[]Pair[bool, error]{NewPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]Pair[int, error]{NewPair[int, error](0, nil)}, // systemctl daemon-reload
		false,
	},
}

var systemdCreateServiceUnitTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig

	// Expected values
	expectedServiceUnitLines []string
}{
	{
		title:  "Created service unit contents are correct",
		config: &basicDaemonConfig,
		expectedServiceUnitLines: []string{
			"[Unit]",
			fmt.Sprintf("Description=%s", basicDaemonConfig.Description),
			"[Service]",
			"Type=simple",
			fmt.Sprintf("ExecStart='%s'", basicDaemonConfig.Program),
		},
	},
	{
		title: "Service unit ExecStart with space is quoted",
		config: &daemon.DaemonConfig{
			Label:       "test-quoting",
			Description: "My program's description (double quotes \" are ok too)",
			Program:     "/path/to/the/program with a space",
		},
		expectedServiceUnitLines: []string{
			"[Unit]",
			"Description=My program's description (double quotes \" are ok too)",
			"[Service]",
			"Type=simple",
			"ExecStart='/path/to/the/program with a space'",
		},
	},
	{
		title: "Service unit ExecStart captures args, quoted and escaped",
		config: &daemon.DaemonConfig{
			Label:       "test-escape",
			Description: "Another program description",
			Program:     "/path/to/the/program with a space",
			Arguments: []string{
				"--my-option",
				"an arg with double quotes \", single quotes ', and spaces!",
			},
		},
		expectedServiceUnitLines: []string{
			"[Unit]",
			"Description=Another program description",
			"[Service]",
			"Type=simple",
			"ExecStart='/path/to/the/program with a space' '--my-option' 'an arg with double quotes \", single quotes \\', and spaces!'",
		},
	},
}

func TestSystemd_Create(t *testing.T) {
	// Set up mocks
	testLogger := &MockTraceLogger{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	testFileSystem := &MockFileSystem{}

	ctx := context.Background()

	systemd := daemon.NewSystemdProvider(testLogger, testUserProvider, testCommandExecutor, testFileSystem)

	for _, tt := range systemdCreateBehaviorTests {
		forceArg := tt.force.ToBoolList()
		for _, force := range forceArg {
			t.Run(fmt.Sprintf("%s (force='%t')", tt.title, force), func(t *testing.T) {
				// Mock responses
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
				for _, retVal := range tt.systemctlDaemonReload {
					testCommandExecutor.On("Run",
						ctx,
						"systemctl",
						[]string{"--user", "daemon-reload"},
					).Return(retVal.First, retVal.Second).Once()
				}

				// Run "Create"
				err := systemd.Create(ctx, tt.config, force)

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
	for _, tt := range systemdCreateServiceUnitTests {
		t.Run(tt.title, func(t *testing.T) {
			var actualFilename string
			var actualFileBytes []byte

			// Mock responses for successful fresh write
			testCommandExecutor.On("Run",
				ctx,
				"systemctl",
				[]string{"--user", "daemon-reload"},
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

			err := systemd.Create(ctx, tt.config, false)
			assert.Nil(t, err)
			mock.AssertExpectationsForObjects(t, testCommandExecutor, testFileSystem)

			// Check filename
			expectedFilename := filepath.Clean(fmt.Sprintf("/my/test/dir/.config/systemd/user/%s.service", tt.config.Label))
			assert.Equal(t, expectedFilename, actualFilename)

			// Check file contents
			// Ensure there's no more than one newline between each line
			// before splitting the file.
			fileContents := strings.TrimSpace(string(actualFileBytes))
			serviceUnitLines := strings.Split(
				regexp.MustCompile(`\n+`).ReplaceAllString(fileContents, "\n"), "\n")
			assert.ElementsMatch(t, tt.expectedServiceUnitLines, serviceUnitLines)

			// Reset mocks
			testCommandExecutor.Mock = mock.Mock{}
			testFileSystem.Mock = mock.Mock{}
		})
	}
}

func TestSystemd_Start(t *testing.T) {
	// Set up mocks
	testLogger := &MockTraceLogger{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	ctx := context.Background()

	systemd := daemon.NewSystemdProvider(testLogger, testUserProvider, testCommandExecutor, nil)

	// Test #1: systemctl succeeds
	t.Run("Calls correct systemctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			ctx,
			"systemctl",
			[]string{"--user", "start", basicDaemonConfig.Label},
		).Return(0, nil).Once()

		err := systemd.Start(ctx, basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: systemctl fails
	t.Run("Returns error when systemctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := systemd.Start(ctx, basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}

func TestSystemd_Stop(t *testing.T) {
	// Set up mocks
	testLogger := &MockTraceLogger{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}

	ctx := context.Background()

	systemd := daemon.NewSystemdProvider(testLogger, testUserProvider, testCommandExecutor, nil)

	// Test #1: systemctl succeeds
	t.Run("Calls correct systemctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			ctx,
			"systemctl",
			[]string{"--user", "stop", basicDaemonConfig.Label},
		).Return(0, nil).Once()

		err := systemd.Stop(ctx, basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: systemctl fails with uncaught error
	t.Run("Returns error when systemctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := systemd.Stop(ctx, basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #3: service unit not found still succeeds
	t.Run("Succeeds if service unit not installed", func(t *testing.T) {
		testCommandExecutor.On("Run",
			ctx,
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(daemon.SystemdUnitNotInstalledErrorCode, nil).Once()

		err := systemd.Stop(ctx, basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}

var systemdRemoveTests = []struct {
	title string

	// Inputs
	label string

	// Mocked responses
	deleteFile            *Pair[bool, error]
	systemctlDaemonReload *Pair[int, error]

	// Expected values
	expectErr bool
}{
	{
		"Unloads and deletes service unit",
		"com.test.service",
		PtrTo(NewPair[bool, error](true, nil)), // delete file
		PtrTo(NewPair[int, error](0, nil)),     // systemctl daemon-reload
		false,
	},
	{
		"Reloads daemon even if service unit missing",
		"com.test.service",
		PtrTo(NewPair[bool, error](false, nil)), // delete file
		PtrTo(NewPair[int, error](0, nil)),      // systemctl daemon-reload
		false,
	},
	{
		"Daemon not reloaded if file cannot be deleted",
		"com.test.service",
		PtrTo(NewPair(false, fmt.Errorf("unhandled error"))), // delete file
		nil, // systemctl daemon-reload
		true,
	},
}

func TestSystemd_Remove(t *testing.T) {
	// Set up mocks
	testLogger := &MockTraceLogger{}
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &MockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &MockCommandExecutor{}
	testFileSystem := &MockFileSystem{}

	ctx := context.Background()

	systemd := daemon.NewSystemdProvider(testLogger, testUserProvider, testCommandExecutor, testFileSystem)

	for _, tt := range systemdRemoveTests {
		t.Run(tt.title, func(t *testing.T) {
			// Setup expected values
			expectedFilename := filepath.Clean(fmt.Sprintf("/my/test/dir/.config/systemd/user/%s.service", tt.label))

			// Mock responses
			if tt.deleteFile != nil {
				testFileSystem.On("DeleteFile",
					expectedFilename,
				).Return(tt.deleteFile.First, tt.deleteFile.Second).Once()
			}
			if tt.systemctlDaemonReload != nil {
				testCommandExecutor.On("Run",
					ctx,
					"systemctl",
					[]string{"--user", "daemon-reload"},
				).Return(tt.systemctlDaemonReload.First, tt.systemctlDaemonReload.Second).Once()
			}

			// Call function
			err := systemd.Remove(ctx, tt.label)
			mock.AssertExpectationsForObjects(t, testCommandExecutor)
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})

		// Reset the mocks between tests
		testCommandExecutor.Mock = mock.Mock{}
		testFileSystem.Mock = mock.Mock{}
	}
}
