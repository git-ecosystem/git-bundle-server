package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/git-ecosystem/git-bundle-server/pkg/auth"
)

type simplePluginAuth struct {
	usernameHash [32]byte
}

// Example auth plugin: basic auth with username "admin" and password
// "{owner}_{repo}" (based on the owner & repo from the route).
// DO NOT USE THIS IN A PRODUCTION BUNDLE SERVER.
func NewSimplePluginAuth(_ json.RawMessage) (auth.AuthMiddleware, error) {
	return &simplePluginAuth{
		usernameHash: sha256.Sum256([]byte("admin")),
	}, nil
}

// Nearly identical to Basic auth, but with a per-request password
func (a *simplePluginAuth) Authorize(r *http.Request, owner string, repo string) auth.AuthResult {
	username, password, ok := r.BasicAuth()
	if ok {
		usernameHash := sha256.Sum256([]byte(username))
		passwordHash := sha256.Sum256([]byte(password))

		perRoutePasswordHash := sha256.Sum256([]byte(owner + "_" + repo))

		usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], a.usernameHash[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], perRoutePasswordHash[:]) == 1)

		if usernameMatch && passwordMatch {
			return auth.Allow()
		} else {
			return auth.Deny(404)
		}
	}

	return auth.Deny(401, auth.Header{"WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`})
}
