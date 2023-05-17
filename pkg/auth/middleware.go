package auth

import (
	"net/http"
)

// AuthMiddleware provides custom authN/authZ functionality to validate requests
// to the bundle web server.
//
// BE CAREFUL! Accesses to the loaded AuthMiddleware instance will *not* be
// thread-safe. Custom implementations should ensure any writes to any common
// state are properly locked.
type AuthMiddleware interface {
	// Authorize interprets the contents of a bundle server request of a valid
	// format (i.e., /<owner>/<repo>[/<bundle>]) and returns an AuthResult
	// indicating whether the request should be allowed or denied. If the
	// AuthResult is invalid (not created with Allow() or Deny()), the server
	// will respond with a 500 status.
	Authorize(r *http.Request, owner string, repo string) AuthResult
}
