package testhelpers

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"os/user"
	"runtime"

	"github.com/github/git-bundle-server/internal/cmd"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/stretchr/testify/mock"
)

func methodIsMocked(m *mock.Mock) bool {
	// Get the calling method name
	pc := make([]uintptr, 1)
	n := runtime.Callers(1, pc)
	if n == 0 {
		// No caller found - fall back on "not mocked"
		return false
	}
	caller := runtime.FuncForPC(pc[0] - 1)
	if caller == nil {
		// Caller not found - fall back on "not mocked"
		return false
	}

	for _, call := range m.ExpectedCalls {
		if call.Method == caller.Name() {
			return true
		}
	}

	return false
}

type notMocked struct{}

var NotMockedValue notMocked = notMocked{}

func mockWithDefault[T any](args mock.Arguments, index int, defaultValue T) T {
	if len(args) <= index {
		return defaultValue
	}

	mockedValue := args.Get(index)
	if _, ok := mockedValue.(notMocked); ok {
		return defaultValue
	}

	return mockedValue.(T)
}

type MockTraceLogger struct {
	mock.Mock
}

func (l *MockTraceLogger) Region(ctx context.Context, category string, label string) (context.Context, func()) {
	fnArgs := mock.Arguments{}
	if methodIsMocked(&l.Mock) {
		fnArgs = l.Called(ctx, category, label)
	}
	return mockWithDefault(fnArgs, 0, ctx), mockWithDefault(fnArgs, 1, func() {})
}

func (l *MockTraceLogger) ChildProcess(ctx context.Context, cmd *exec.Cmd) (func(error), func()) {
	fnArgs := mock.Arguments{}
	if methodIsMocked(&l.Mock) {
		fnArgs = l.Called(ctx, cmd)
	}
	return mockWithDefault(fnArgs, 0, func(error) {}), mockWithDefault(fnArgs, 1, func() {})
}

func (l *MockTraceLogger) LogCommand(ctx context.Context, commandName string) context.Context {
	fnArgs := mock.Arguments{}
	if methodIsMocked(&l.Mock) {
		fnArgs = l.Called(ctx, commandName)
	}
	return mockWithDefault(fnArgs, 0, ctx)
}

func (l *MockTraceLogger) Error(ctx context.Context, err error) error {
	// Input validation
	if err == nil {
		panic("err must be nil")
	}

	fnArgs := mock.Arguments{}
	if methodIsMocked(&l.Mock) {
		fnArgs = l.Called(ctx, err)
	}
	return mockWithDefault(fnArgs, 0, err)
}

func (l *MockTraceLogger) Errorf(ctx context.Context, format string, a ...any) error {
	fnArgs := mock.Arguments{}
	if methodIsMocked(&l.Mock) {
		fnArgs = l.Called(ctx, format, a)
	}
	return mockWithDefault(fnArgs, 0, fmt.Errorf(format, a...))
}

func (l *MockTraceLogger) Exit(ctx context.Context, exitCode int) {
	if methodIsMocked(&l.Mock) {
		l.Called(ctx, exitCode)
	}
}

func (l *MockTraceLogger) Fatal(ctx context.Context, err error) {
	if methodIsMocked(&l.Mock) {
		l.Called(ctx, err)
	}
}

func (l *MockTraceLogger) Fatalf(ctx context.Context, format string, a ...any) {
	if methodIsMocked(&l.Mock) {
		l.Called(ctx, format, a)
	}
}

type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) CurrentUser() (*user.User, error) {
	fnArgs := m.Called()
	return fnArgs.Get(0).(*user.User), fnArgs.Error(1)
}

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) RunStdout(ctx context.Context, command string, args ...string) (int, error) {
	fnArgs := m.Called(ctx, command, args)
	return fnArgs.Int(0), fnArgs.Error(1)
}

func (m *MockCommandExecutor) RunQuiet(ctx context.Context, command string, args ...string) (int, error) {
	fnArgs := m.Called(ctx, command, args)
	return fnArgs.Int(0), fnArgs.Error(1)
}

func (m *MockCommandExecutor) Run(ctx context.Context, command string, args []string, settings ...cmd.Setting) (int, error) {
	fnArgs := m.Called(ctx, command, args, settings)
	return fnArgs.Int(0), fnArgs.Error(1)
}

type MockLockFile struct {
	mock.Mock
}

func (m *MockLockFile) Commit() error {
	fnArgs := m.Called()
	return fnArgs.Error(0)
}

func (m *MockLockFile) Rollback() error {
	fnArgs := m.Called()
	return fnArgs.Error(0)
}

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) GetLocalExecutable(name string) (string, error) {
	fnArgs := m.Called(name)
	return fnArgs.String(0), fnArgs.Error(1)
}

func (m *MockFileSystem) FileExists(filename string) (bool, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockFileSystem) WriteFile(filename string, content []byte) error {
	fnArgs := m.Called(filename, content)
	return fnArgs.Error(0)
}

func (m *MockFileSystem) WriteLockFileFunc(filename string, writeFunc func(io.Writer) error) (common.LockFile, error) {
	fnArgs := m.Called(filename, writeFunc)
	return fnArgs.Get(0).(common.LockFile), fnArgs.Error(1)
}

func (m *MockFileSystem) DeleteFile(filename string) (bool, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockFileSystem) ReadFileLines(filename string) ([]string, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Get(0).([]string), fnArgs.Error(1)
}

type MockGitHelper struct {
	mock.Mock
}

func (m *MockGitHelper) CreateBundle(ctx context.Context, repoDir string, filename string) (bool, error) {
	fnArgs := m.Called(ctx, repoDir, filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockGitHelper) CreateBundleFromRefs(ctx context.Context, repoDir string, filename string, refs map[string]string) error {
	fnArgs := m.Called(ctx, repoDir, filename, refs)
	return fnArgs.Error(0)
}

func (m *MockGitHelper) CreateIncrementalBundle(ctx context.Context, repoDir string, filename string, prereqs []string) (bool, error) {
	fnArgs := m.Called(ctx, repoDir, filename, prereqs)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockGitHelper) CloneBareRepo(ctx context.Context, url string, destination string) error {
	fnArgs := m.Called(ctx, url, destination)
	return fnArgs.Error(0)
}

func (m *MockGitHelper) UpdateBareRepo(ctx context.Context, repoDir string) error {
	fnArgs := m.Called(ctx, repoDir)
	return fnArgs.Error(0)
}

func (m *MockGitHelper) GetRemoteUrl(ctx context.Context, repoDir string) (string, error) {
	fnArgs := m.Called(ctx, repoDir)
	return fnArgs.String(0), fnArgs.Error(1)
}
