package daemon_test

import (
	"os/user"

	"github.com/github/git-bundle-server/internal/daemon"
	"github.com/stretchr/testify/mock"
)

/*********************************************/
/***************** Constants *****************/
/*********************************************/

var basicDaemonConfig = daemon.DaemonConfig{
	Label:       "com.example.testdaemon",
	Description: "Test service",
	Program:     "/usr/local/bin/test/git-bundle-web-server",
}

/*********************************************/
/************* Types & Functions *************/
/*********************************************/

type pair[T any, R any] struct {
	first  T
	second R
}

func newPair[T any, R any](first T, second R) pair[T, R] {
	return pair[T, R]{
		first:  first,
		second: second,
	}
}

type boolArg int

const (
	False boolArg = iota
	True
	Any
)

func (b boolArg) toBoolList() []bool {
	switch b {
	case False:
		return []bool{false}
	case True:
		return []bool{true}
	case Any:
		return []bool{false, true}
	default:
		panic("invalid bool arg value")
	}
}

/*********************************************/
/******************* Mocks *******************/
/*********************************************/

type mockUserProvider struct {
	mock.Mock
}

func (m *mockUserProvider) CurrentUser() (*user.User, error) {
	fnArgs := m.Called()
	return fnArgs.Get(0).(*user.User), fnArgs.Error(1)
}

type mockCommandExecutor struct {
	mock.Mock
}

func (m *mockCommandExecutor) Run(command string, args ...string) (int, error) {
	fnArgs := m.Called(command, args)
	return fnArgs.Int(0), fnArgs.Error(1)
}

type mockFileSystem struct {
	mock.Mock
}

func (m *mockFileSystem) FileExists(filename string) (bool, error) {
	fnArgs := m.Called(filename)
	return fnArgs.Bool(0), fnArgs.Error(1)
}

func (m *mockFileSystem) WriteFile(filename string, content []byte) error {
	fnArgs := m.Called(filename, content)
	return fnArgs.Error(0)
}
