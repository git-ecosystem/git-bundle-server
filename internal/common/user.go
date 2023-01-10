package common

import (
	"os/user"
)

type UserProvider interface {
	CurrentUser() (*user.User, error)
}

type userProvider struct{}

func NewUserProvider() *userProvider {
	return &userProvider{}
}

func (u *userProvider) CurrentUser() (*user.User, error) {
	return user.Current()
}
