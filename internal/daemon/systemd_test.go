package daemon_test

import (
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/github/git-bundle-server/internal/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var systemdCreateTests = []struct {
	title string

	// Inputs
	config *daemon.DaemonConfig
	force  boolArg

	// Mocked responses (ordered per list!)
	fileExists            []pair[bool, error]
	writeFile             []error
	systemctlDaemonReload []pair[int, error]

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
		[]pair[bool, error]{newPair[bool, error](false, nil)}, // file exists
		[]error{nil}, // write file
		[]pair[int, error]{newPair[int, error](0, nil)}, // systemctl daemon-reload
		false,
	},
	{
		"Service unit exists, doesn't write file or reload",
		&daemon.DaemonConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/local/bin/test/git-bundle-web-server",
		},
		False,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{},            // write file
		[]pair[int, error]{}, // systemctl daemon-reload
		false,
	},
	{
		"'force' option overwrites service unit and reloads daemon",
		&daemon.DaemonConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/local/bin/test/git-bundle-web-server",
		},
		True,
		[]pair[bool, error]{newPair[bool, error](true, nil)}, // file exists
		[]error{nil}, // write file
		[]pair[int, error]{newPair[int, error](0, nil)}, // systemctl daemon-reload
		false,
	},
}

func TestSystemd_Create(t *testing.T) {
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

	systemd := daemon.NewSystemdProvider(testUserProvider, testCommandExecutor, testFileSystem)

	for _, tt := range systemdCreateTests {
		forceArg := tt.force.toBoolList()
		for _, force := range forceArg {
			t.Run(fmt.Sprintf("%s (force='%t')", tt.title, force), func(t *testing.T) {
				// Mock responses
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
				for _, retVal := range tt.systemctlDaemonReload {
					testCommandExecutor.On("Run",
						"systemctl",
						[]string{"--user", "daemon-reload"},
					).Return(retVal.first, retVal.second).Once()
				}

				// Run "Create"
				err := systemd.Create(tt.config, force)

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

		err := systemd.Create(&basicDaemonConfig, false)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor, testFileSystem)

		// Check filename
		expectedFilename := filepath.Clean(fmt.Sprintf("/my/test/dir/.config/systemd/user/%s.service", basicDaemonConfig.Label))
		assert.Equal(t, expectedFilename, actualFilename)

		// Check file contents
		fileContents := string(actualFileBytes)
		assert.Contains(t, fileContents, fmt.Sprintf("ExecStart=%s", basicDaemonConfig.Program))
		assert.Contains(t, fileContents, fmt.Sprintf("Description=%s", basicDaemonConfig.Description))
	})
}

func TestSystemd_Start(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &mockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &mockCommandExecutor{}

	systemd := daemon.NewSystemdProvider(testUserProvider, testCommandExecutor, nil)

	// Test #1: systemctl succeeds
	t.Run("Calls correct systemctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			"systemctl",
			[]string{"--user", "start", basicDaemonConfig.Label},
		).Return(0, nil).Once()

		err := systemd.Start(basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: systemctl fails
	t.Run("Returns error when systemctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := systemd.Start(basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}

func TestSystemd_Stop(t *testing.T) {
	// Set up mocks
	testUser := &user.User{
		Uid:      "123",
		Username: "testuser",
		HomeDir:  "/my/test/dir",
	}
	testUserProvider := &mockUserProvider{}
	testUserProvider.On("CurrentUser").Return(testUser, nil)

	testCommandExecutor := &mockCommandExecutor{}

	systemd := daemon.NewSystemdProvider(testUserProvider, testCommandExecutor, nil)

	// Test #1: systemctl succeeds
	t.Run("Calls correct systemctl command", func(t *testing.T) {
		testCommandExecutor.On("Run",
			"systemctl",
			[]string{"--user", "stop", basicDaemonConfig.Label},
		).Return(0, nil).Once()

		err := systemd.Stop(basicDaemonConfig.Label)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})

	// Reset the mock structure between tests
	testCommandExecutor.Mock = mock.Mock{}

	// Test #2: systemctl fails with uncaught error
	t.Run("Returns error when systemctl fails", func(t *testing.T) {
		testCommandExecutor.On("Run",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("[]string"),
		).Return(1, nil).Once()

		err := systemd.Stop(basicDaemonConfig.Label)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, testCommandExecutor)
	})
}
