package testhelpers

import (
	"os/user"

	"github.com/stretchr/testify/mock"
)

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

func (m *MockCommandExecutor) Run(command string, args ...string) (int, error) {
	fnArgs := m.Called(command, args)
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

func (m *MockFileSystem) ReadFileLines(filename string) ([]string, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Get(0).([]string), fnArgs.Error(1)
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	fnArgs := m.Called()
	return fnArgs.String(0), fnArgs.Error(1)
}
