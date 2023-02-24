package testhelpers

import (
	"context"
	"fmt"
	"os/user"
	"runtime"

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

func (m *MockCommandExecutor) Run(ctx context.Context, command string, args ...string) (int, error) {
	fnArgs := m.Called(ctx, command, args)
	return fnArgs.Int(0), fnArgs.Error(1)
}

type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) FileExists(filename string) (bool, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockFileSystem) WriteFile(filename string, content []byte) error {
	fnArgs := m.Called(filename, content)
	return fnArgs.Error(0)
}

func (m *MockFileSystem) DeleteFile(filename string) (bool, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *MockFileSystem) ReadFileLines(filename string) ([]string, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Get(0).([]string), fnArgs.Error(1)
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	fnArgs := m.Called()
	return fnArgs.String(0), fnArgs.Error(1)
}
